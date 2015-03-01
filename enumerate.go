package wsman

import (
	"github.com/VictorLowther/simplexml/dom"
	"github.com/VictorLowther/simplexml/search"
)

func (c *Client) enumRelease(context *dom.Element) {
	req := c.NewMessage(RELEASE)
	body := dom.Elem("Release", NS_WSMEN)
	req.SetBody(body)
	body.AddChild(context)
	req.Send()
}

// Enumerate enumerates items available at a given resource.
func (client *Client) Enumerate(resource string, filter *dom.Element, options map[string]string) (*Message, error) {
	searchContext := search.Tag("EnumerationContext", NS_WSMEN)
	req := client.NewMessage(ENUMERATE)
	req.Options = options
	req.SetHeader(Resource(resource))
	body := dom.Elem("Enumerate", NS_WSMEN)
	req.SetBody(body)
	optimizeEnum := dom.Elem("OptimizeEnumeration", NS_WSMAN)
	maxElem := dom.Elem("MaxElements", NS_WSMAN)
	maxElem.Content = []byte("100")
	body.AddChildren(optimizeEnum, maxElem)
	if filter != nil {
		body.AddChild(filter)
	}
	resp, err := req.Send()
	if err != nil {
		return nil, err
	}
	context := search.First(searchContext, resp.AllBodyElements())
	items := search.First(search.Tag("Items", "*"), resp.AllBodyElements())
	for context != nil {
		req = client.NewMessage(PULL)
		req.SetHeader(Resource(resource))
		body = dom.Elem("Pull", NS_WSMEN)
		req.SetBody(body)
		body.AddChild(maxElem)
		body.AddChild(context)
		nextResp, err := req.Send()
		if err != nil {
			client.enumRelease(context)
			return nil, err
		}
		context = search.First(searchContext, nextResp.AllBodyElements())
		extraItems := search.First(search.Tag("Items", "*"), nextResp.AllBodyElements())
		if extraItems != nil {
			items.AddChildren(extraItems.Children()...)
		}
	}
	return resp, nil
}

func (client *Client) EnumerateEPR(resource string, filter *dom.Element) (*Message, error) {
	opts := map[string]string{
		"EnumerationMode": "EnumerateEPR",
	}
	return client.Enumerate(resource, filter, opts)
}
