package wsman

/*
Copyright 2015 Victor Lowther <victor.lowther@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"fmt"
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

func enumHelper(firstreq, resp *Message) error {
	searchContext := search.Tag("EnumerationContext", NS_WSMEN)
	context := search.First(searchContext, resp.AllBodyElements())
	items := search.First(search.Tag("Items", "*"), resp.AllBodyElements())
	resource := firstreq.GetHeader(dom.Elem("ResourceURI", NS_WSMAN))
	maxElem := dom.ElemC("MaxElements", NS_WSMAN, "100")
	if resource == nil {
		return fmt.Errorf("WSMAN Enumerate request did not have RequestURI")
	}
	for context != nil {
		req := resp.client.NewMessage(PULL)
		req.SetHeader(resource)
		body := dom.Elem("Pull", NS_WSMEN)
		req.SetBody(body)
		body.AddChild(maxElem)
		body.AddChild(context)
		nextResp, err := req.Send()
		if err != nil {
			resp.client.enumRelease(context)
			return err
		}
		context = search.First(searchContext, nextResp.AllBodyElements())
		extraItems := search.First(search.Tag("Items", "*"), nextResp.AllBodyElements())
		if extraItems != nil {
			items.AddChildren(extraItems.Children()...)
		}
	}
	return nil
}

// Enumerate creates a wsman.Message that will enumerate all the objects
// available at resource.  If there are many objects, it will arrange
// for the appropriate series of wsman Pull calls to be performed, so you can
// be certian that the response to this message has all the objects you specify.
func (client *Client) Enumerate(resource string) *Message {
	req := client.NewMessage(ENUMERATE).ResourceURI(resource)
	body := dom.Elem("Enumerate", NS_WSMEN)
	req.SetBody(body)
	optimizeEnum := dom.Elem("OptimizeEnumeration", NS_WSMAN)
	maxElem := dom.ElemC("MaxElements", NS_WSMAN, "100")
	maxElem.Content = []byte("100")
	body.AddChildren(optimizeEnum, maxElem)
	req.replyHelper = enumHelper
	return req
}

// EnumerateEPR creates a message that will enumerate the endpoints for a given resource.
func (client *Client) EnumerateEPR(resource string) *Message {
	return client.Enumerate(resource).Options("EnumerationMode", "EnumerateEPR")
}
