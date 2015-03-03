package wsman

// Invoke creates a wsman.Message that will invoke method on resource.
// After creating the Message, you need to add the appropriate selectors
// with msg.Selectors(), and the appropriate parameters with msg.Parameters()
func (c *Client) Invoke(resource, method string) *Message {
	res := c.NewMessage(resource + "/" + method)
	res.SetHeader(Resource(resource))
	return res
}
