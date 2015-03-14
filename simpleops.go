package wsman

// Invoke creates a wsman.Message that will invoke method on resource.
// After creating the Message, you need to add the appropriate selectors
// with msg.Selectors(), and the appropriate parameters with msg.Parameters()
func (c *Client) Invoke(resource, method string) *Message {
	return c.NewMessage(resource + "/" + method).ResourceURI(resource)
}

// Get creates a wsman.Message that will get an instance
// at the passed-in resource.
func (c *Client) Get(resource string) *Message {
	return c.NewMessage(GET).ResourceURI(resource)
}

// Put creates a wsman.Message that will update the passed-in
// resource.  The updated resource should be passed in as the
// only element in the Body of the messate.
func (c *Client) Put(resource string) *Message {
	return c.NewMessage(PUT).ResourceURI(resource)
}

// Create creates a wsman.Message that will update the passed-in
// resource.  The updated resource should be passed in as the
// only element in the Body of the messate.
func (c *Client) Create(resource string) *Message {
	return c.NewMessage(CREATE).ResourceURI(resource)
}

// Delete creates a wsman.Message that will update the passed-in
// resource.  The updated resource should be passed in as the
// only element in the Body of the messate.
func (c *Client) Delete(resource string) *Message {
	return c.NewMessage(DELETE).ResourceURI(resource)
}
