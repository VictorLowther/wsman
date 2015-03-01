package wsman

import (
	"github.com/VictorLowther/simplexml/dom"
	"github.com/VictorLowther/soap"
	uuid "github.com/satori/go.uuid"
	"fmt"
)

// Message represents WSMAN messages
type Message struct {
	*soap.Message
	client *Client
	// Options are used to modify how certian WSMAN
	// operations work.  For now, the only thing we use it for
	// is to make EnumerateEPR work.  See
	// http://www.dmtf.org/sites/default/files/standards/documents/DSP0226_1.2.0.pdf,
	// section 6.4 for more information.
	Options map[string]string
}

// Resource turns a resource URI into an appropriate DOM element
// for inclusion in the SOAP header.
func Resource(uri string) *dom.Element {
	return soap.MuElemC("ResourceURI", NS_WSMAN, uri)
}

// NewMessage creates a new wsman.Message that can be sent
// via c.  It populates the message with the passed action
// and some other necessary headers.
func (c *Client) NewMessage(action string) (msg *Message) {
	msg = &Message{
		Message: soap.NewMessage(),
		client:  c,
		Options: map[string]string{},
	}
	msg.SetHeader(
		soap.MuElemC("Action", NS_WSA, action),
		soap.MuElemC("To", NS_WSA, c.target),
		soap.MuElemC("MessageID", NS_WSA, fmt.Sprintf("uuid:%s", uuid.NewV4())),
		dom.Elem("ReplyTo", NS_WSA).AddChild(
			soap.MuElemC("Address", NS_WSA,
				"http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous")))
	return msg
}

// Send sends a message to the endpoint of the Client it was
// constructed with, and returns either the Message that was
// returned, or an error statung what went wrong.
func (m *Message) Send() (*Message, error) {
	if len(m.Options) > 0 {
		optset := dom.Elem("OptionSet", NS_WSMAN)
		for k, v := range m.Options {
			elem := dom.Elem("Option", NS_WSMAN).Attr("Name", "", k)
			elem.Content = []byte(v)
			optset.AddChild(elem)
		}
		m.SetHeader(optset)
	}
	res, err := m.client.post(m.Message)
	if err != nil {
		return nil, err
	}
	return &Message{Message: res}, nil
}
