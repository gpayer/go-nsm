package nsm

type OptionCapabilities struct {
	caps []ClientCapability
}

func (o *OptionCapabilities) configure(c *Client) {
	c.clientCapabilities = o.caps
}

var _ Option = (*OptionCapabilities)(nil)

func SetOptCapabilities(caps ...ClientCapability) Option {
	return &OptionCapabilities{
		caps: caps,
	}
}

type OptionOpenHandler struct {
	handler func(projectPath, displayName, clientID string) error
}

func (o *OptionOpenHandler) configure(c *Client) {
	c.clientOpen = o.handler
}

var _ Option = (*OptionOpenHandler)(nil)

func SetOpenHandler(handler func(projectPath, displayName, clientID string) error) Option {
	return &OptionOpenHandler{
		handler: handler,
	}
}

type OptionSaveHandler struct {
	handler func() error
}

func (o *OptionSaveHandler) configure(c *Client) {
	c.clientSave = o.handler
}

var _ Option = (*OptionSaveHandler)(nil)

func SetSaveHandler(handler func() error) Option {
	return &OptionSaveHandler{
		handler: handler,
	}
}
