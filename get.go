package wsman

// Get creates a wsman.Message that will get an instance
// at the passed-in resource.
func (c *Client) Get(resource string) *Message {
	res := c.NewMessage(GET)
	res.SetHeader(Resource(resource))
	return res
}
