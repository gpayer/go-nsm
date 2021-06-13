package nsm

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gpayer/go-osc/osc"
)

type ServerCapability string

const (
	CapabilityServerControl     ServerCapability = "server_control"
	CapabilityServerBroadcast   ServerCapability = "broadcast"
	CapabilityServerOptionalGui ServerCapability = "optional-gui"
)

type ClientCapability string

const (
	CapabilityClientSwitch      ClientCapability = "switch"
	CapabilityClientDirty       ClientCapability = "dirty"
	CapabilityClientProgress    ClientCapability = "progress"
	CapabilityClientMessage     ClientCapability = "message"
	CapabilityClientOptionalGUI ClientCapability = "optional-gui"
)

const (
	ServerAnnounce = "/nsm/server/announce"
	ClientOpen     = "/nsm/client/open"
	ClientSave     = "/nsm/client/save"
)

type Client struct {
	Osc                *osc.Client
	Server             string
	Servername         string
	serverCapabilities []ServerCapability
	clientCapabilities []ClientCapability
	clientOpen         func(projectPath, displayName, clientID string) error
	clientSave         func() error
}

type Option interface {
	configure(*Client)
}

func NewClient(name string, opts ...Option) (*Client, error) {
	nsmURL, ok := os.LookupEnv("NSM_URL")
	if !ok {
		return nil, errors.New("NSM_URL not defined")
	}
	u, err := url.Parse(nsmURL)
	if err != nil {
		return nil, err
	}
	serverport, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}

	client := &Client{
		Osc:    osc.NewClient(u.Hostname(), serverport),
		Server: u.Host,
	}

	for _, o := range opts {
		o.configure(client)
	}

	if client.clientOpen == nil {
		return nil, errors.New("no client open handler configured")
	}
	if client.clientSave == nil {
		return nil, errors.New("no client save handler configured")
	}

	announceReceived := make(chan error)

	// TODO: setup message handlers
	d := osc.NewStandardDispatcher()
	d.AddMsgHandler("/reply", func(msg *osc.Message) {
		tags, _ := msg.TypeTags()
		if len(tags) < 2 {
			return
		}
		replyPath, ok := msg.Arguments[0].(string)
		if !ok {
			return
		}
		if replyPath == ServerAnnounce {
			if tags != ",ssss" {
				return
			}
			client.Servername = msg.Arguments[2].(string)
			capabilities := msg.Arguments[3].(string)

			for _, cap := range strings.Split(capabilities, ":") {
				if cap != "" {
					client.serverCapabilities = append(client.serverCapabilities, ServerCapability(cap))
				}
			}
			announceReceived <- nil
		}
		// TODO: other replies
	})
	d.AddMsgHandler("/error", func(msg *osc.Message) {
		tags, _ := msg.TypeTags()
		if tags != ",sis" {
			return
		}
		replyPath := msg.Arguments[0].(string)
		errCode := msg.Arguments[1].(int32)
		errMsg := msg.Arguments[2].(string)
		fmt.Printf("DEBUG: /error %s %d %s\n", replyPath, errCode, errMsg)
		if replyPath == ServerAnnounce {
			announceReceived <- fmt.Errorf("server replied with error %d: %s", errCode, errMsg)
		}
		// TODO: send other error messages back to client
	})
	d.AddMsgHandler(ClientOpen, func(msg *osc.Message) {
		tags, _ := msg.TypeTags()
		if tags != ",sss" {
			return
		}
		projectPath := msg.Arguments[0].(string)
		displayName := msg.Arguments[1].(string)
		clientID := msg.Arguments[1].(string)
		err := client.clientOpen(projectPath, displayName, clientID)
		if err != nil {
			msg := osc.NewMessage("/error", ClientOpen, -1, err.Error())
			client.Osc.Send(msg)
		} else {
			msg := osc.NewMessage("/reply", ClientOpen, "ok")
			client.Osc.Send(msg)
		}
	})
	d.AddMsgHandler(ClientSave, func(msg *osc.Message) {
		err := client.clientSave()
		if err != nil {
			msg := osc.NewMessage("/error", ClientSave, -1, err.Error())
			client.Osc.Send(msg)
		} else {
			msg := osc.NewMessage("/reply", ClientSave, "ok")
			client.Osc.Send(msg)
		}
	})

	client.Osc.SetDispatcher(d)
	// TODO: connect
	// TODO: listen and serve thread
	// TODO: wait for connection
	// TODO: send announce message
	// TODO: wait for initial communication to finish
	select {
	case err := <-announceReceived:
		if err != nil {
			return nil, err
		}
	case <-time.After(10 * time.Second):
		return nil, errors.New("timeout while waiting for server announce reply")
	}

	return client, nil
}

func (c *Client) ServerHasCapability(cap ServerCapability) bool {
	// TODO: implement
	return false
}
