package nsm

import (
	"context"
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
	ServerAnnounce        = "/nsm/server/announce"
	ClientOpen            = "/nsm/client/open"
	ClientSave            = "/nsm/client/save"
	ClientIsDirty         = "/nsm/client/is_dirty"
	ClientIsClean         = "/nsm/client/is_clean"
	ClientSessionLoaded   = "/nsm/client/session_is_loaded"
	ClientShowOptionalGui = "/nsm/client/show_optional_gui"
	ClientHideOptionalGui = "/nsm/client/hide_optional_gui"
)

type ClientState int

const (
	StateInitializing ClientState = iota
	StateConnecting
	StateConnected
	StateError
)

type Client struct {
	Osc                 *osc.Client
	Server              string
	Servername          string
	State               ClientState
	Error               error
	serverCapabilities  []ServerCapability
	clientCapabilities  []ClientCapability
	clientOpen          func(projectPath, displayName, clientID string) error
	clientSave          func() error
	clientShowGui       func(showGui bool)
	clientSessionLoaded func()
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
		State:  StateInitializing,
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
	if client.HasCapability(CapabilityClientOptionalGUI) && client.clientShowGui == nil {
		return nil, errors.New("option optional-gui set, but no optional gui handler configured")
	}

	announceReceived := make(chan error)

	// setup message handlers
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
			code := ErrGeneral
			if nsmErr, ok := err.(*Error); ok {
				code = nsmErr.Code
			}
			msg := osc.NewMessage("/error", ClientSave, code, err.Error())
			client.Osc.Send(msg)
		} else {
			msg := osc.NewMessage("/reply", ClientSave, "ok")
			client.Osc.Send(msg)
		}
	})
	d.AddMsgHandler(ClientSessionLoaded, func(msg *osc.Message) {
		if client.clientSessionLoaded != nil {
			client.clientSessionLoaded()
		}
	})
	d.AddMsgHandler(ClientShowOptionalGui, func(msg *osc.Message) {
		if client.clientShowGui != nil {
			client.clientShowGui(true)
		}
	})
	d.AddMsgHandler(ClientHideOptionalGui, func(msg *osc.Message) {
		if client.clientShowGui != nil {
			client.clientShowGui(false)
		}
	})

	client.Osc.SetDispatcher(d)

	// connect
	err = client.Osc.Connect()
	if err != nil {
		return nil, err
	}
	client.State = StateConnecting

	// listen and serve thread
	go func() {
		err := client.Osc.ListenAndServe()
		if err != nil {
			client.Osc.Close()
			client.State = StateError
			client.Error = err
		}
	}()

	// wait for connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for !client.Osc.Connected() {
		select {
		case <-time.After(100 * time.Millisecond):
		case <-ctx.Done():
			client.Osc.Close()
			return nil, errors.New("connection failed")
		}
	}
	client.State = StateConnected

	// send announce message
	msg := osc.NewMessage(ServerAnnounce)
	var capListBuilder strings.Builder
	capListBuilder.WriteRune(':')
	for _, cap := range client.clientCapabilities {
		capListBuilder.WriteString(string(cap))
	}
	capListBuilder.WriteRune(':')
	msg.Append(name, capListBuilder.String(), os.Args[0], int32(1), int32(0), int32(os.Getpid()))
	err = client.Osc.Send(msg)
	if err != nil {
		client.Osc.Close()
		return nil, fmt.Errorf("error sending msg: %v", err)
	}

	// wait for initial communication to finish
	select {
	case err := <-announceReceived:
		if err != nil {
			return nil, err
		}
	case <-time.After(10 * time.Second):
		// TODO: how to end ListenAndServer?
		return nil, errors.New("timeout while waiting for server announce reply")
	}

	return client, nil
}

func (c *Client) ServerHasCapability(cap ServerCapability) bool {
	for _, scap := range c.serverCapabilities {
		if cap == scap {
			return true
		}
	}
	return false
}

func (c *Client) HasCapability(cap ClientCapability) bool {
	for _, ccap := range c.clientCapabilities {
		if cap == ccap {
			return true
		}
	}
	return false
}

func (c *Client) SetDirty(dirty bool) {
	if c.HasCapability(CapabilityClientDirty) {
		if dirty {
			c.Osc.Send(osc.NewMessage(ClientIsDirty))
		} else {
			c.Osc.Send(osc.NewMessage(ClientIsClean))
		}
	}
}
