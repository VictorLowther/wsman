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
	searchEnd := search.Tag("EndOfSequence", NS_WSMAN)
	if search.First(searchEnd, resp.AllBodyElements()) != nil {
		return nil
	}
	context := search.First(searchContext, resp.AllBodyElements())
	items := search.First(search.Tag("Items", "*"), resp.AllBodyElements())
	resource := firstreq.GetHeader(dom.Elem("ResourceURI", NS_WSMAN))
	maxElem := search.First(search.Tag("MaxElements", NS_WSMAN), firstreq.AllBodyElements())
	enumEpr := search.First(search.Tag("EnumerationMode", NS_WSMAN), firstreq.AllBodyElements())
	if resource == nil {
		return fmt.Errorf("WSMAN Enumerate request did not have RequestURI")
	}
	if items == nil {
		enumResp := search.First(search.Tag("EnumerateResponse", "*"), resp.AllBodyElements())
		if enumResp == nil {
			return fmt.Errorf("Enumeration response did not have EnumerateResponse body element")
		}
		items = dom.Elem("Items", NS_WSMAN)
		enumResp.AddChild(items)
	}

	for context != nil {
		req := resp.client.NewMessage(PULL)
		req.SetHeader(resource)
		body := dom.Elem("Pull", NS_WSMEN)
		req.SetBody(body)
		body.AddChild(context)
		if maxElem != nil {
			body.AddChild(maxElem)
		}
		if enumEpr != nil {
			body.AddChild(enumEpr)
		}
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
		if search.First(searchEnd, nextResp.AllBodyElements()) != nil {
			break
		}
	}
	return nil
}

func (c *Client) enumerate(resource string, epr, optimize bool) *Message {
	req := c.NewMessage(ENUMERATE).ResourceURI(resource)
	body := dom.Elem("Enumerate", NS_WSMEN)
	req.SetBody(body)
	if optimize {
		optimizeEnum := dom.Elem("OptimizeEnumeration", NS_WSMAN)
		maxElem := dom.ElemC("MaxElements", NS_WSMAN, "100")
		maxElem.Content = []byte("100")
		body.AddChildren(optimizeEnum, maxElem)
	}
	if epr {
		enumEpr := dom.Elem("EnumerationMode", NS_WSMAN)
		enumEpr.Content = []byte("EnumerateEPR")
		body.AddChild(enumEpr)
	}
	req.replyHelper = enumHelper
	return req
}

// Enumerate creates a wsman.Message that will enumerate all the objects
// available at resource.  If there are many objects, it will arrange
// for the appropriate series of wsman Pull calls to be performed, so you can
// be certian that the response to this message has all the objects you specify.
func (c *Client) Enumerate(resource string) *Message {
	return c.enumerate(resource, false, c.OptimizeEnum)
}

// EnumerateEPR creates a message that will enumerate the endpoints for a given resource.
func (c *Client) EnumerateEPR(resource string) *Message {
	return c.enumerate(resource, true, c.OptimizeEnum)
}

func (m *Message) EnumItems() ([]*dom.Element, error) {
	action, err := m.GHC("Action")
	if err != nil || action != ENUMERATE+"Response" {
		return nil, fmt.Errorf("Not an EnumerateResponse message!")
	}
	items := search.First(search.Tag("Items", NS_WSMAN), m.AllBodyElements())
	if items == nil {
		return nil, fmt.Errorf("No items returned from EnumItems")
	}
	return items.Children(), nil
}
