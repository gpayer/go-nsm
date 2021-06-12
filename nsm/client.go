package nsm

import (
	"errors"
	"net/url"
	"os"
	"strconv"

	"github.com/gpayer/go-osc/osc"
)

type Client struct {
	Osc    *osc.Client
	Server string
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

	// TODO: setup message handlers
	// TODO: connect
	// TODO: listen and serve thread
	// TODO: wait for connection
	// TODO: wait for initial communication to finish

	return client, nil
}
