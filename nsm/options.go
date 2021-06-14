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

type OptionShowGuiHandler struct {
	handler func(showGui bool)
}

func (o *OptionShowGuiHandler) configure(c *Client) {
	c.clientShowGui = o.handler
}

var _ Option = (*OptionShowGuiHandler)(nil)

func SetShowGuiHandler(handler func(bool)) Option {
	return &OptionShowGuiHandler{
		handler: handler,
	}
}

type OptionSessionLoadedHandler struct {
	handler func()
}

func (o *OptionSessionLoadedHandler) configure(c *Client) {
	c.clientSessionLoaded = o.handler
}

var _ Option = (*OptionSessionLoadedHandler)(nil)

func SetSessionLoadedHandler(handler func()) Option {
	return &OptionSessionLoadedHandler{
		handler: handler,
	}
}
