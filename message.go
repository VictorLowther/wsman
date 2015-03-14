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
	"github.com/VictorLowther/soap"
	uuid "github.com/satori/go.uuid"
	"strings"
)

// Message represents WSMAN messages
type Message struct {
	*soap.Message
	client *Client
	// First arg is the request, second is the initial response.
	// For now, this is used to allow Enumerate to Pull additional
	// replys without having to make API users do it.
	replyHelper func(*Message, *Message) error
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

// Options are used to modify how certian WSMAN operations work.
// For now, the only thing we use it for is to make EnumerateEPR work.
// See http://www.dmtf.org/sites/default/files/standards/documents/DSP0226_1.2.0.pdf,
// section 6.4 for more information.
// Options takes an even number of strings, which should be key/value pairs.
// They will be added to the soap Header as Options in an OptionSet.
//
// Example:
//
// client.Enumerate("http://my.resource/url") -> msg:
//    <?xml version="1.0" encoding="UTF-8"?>
//    <ns0:Envelope xmlns:ns0="http://www.w3.org/2003/05/soap-envelope" xmlns:ns1="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:ns2="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xmlns:ns3="http://schemas.xmlsoap.org/ws/2004/09/enumeration">
//     <ns0:Header>
//     <ns1:Action ns0:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/enumeration/Enumerate</ns1:Action>
//      <ns1:To ns0:mustUnderstand="true">https://192.168.128.41:443/wsman</ns1:To>
//      <ns1:MessageID ns0:mustUnderstand="true">uuid:0590dc5f-369b-4ed3-9da4-8461a6f66dae</ns1:MessageID>
//      <ns1:ReplyTo>
//       <ns1:Address ns0:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</ns1:Address>
//      </ns1:ReplyTo>
//      <ns2:ResourceURI ns0:mustUnderstand="true">http://my.resource/url</ns2:ResourceURI>
//     </ns0:Header>
//     <ns0:Body>
//      <ns3:Enumerate>
//       <ns2:OptimizeEnumeration/>
//       <ns2:MaxElements>100</ns2:MaxElements>
//      </ns3:Enumerate>
//     </ns0:Body>
//    </ns0:Envelope>
// msg.Options("EnumerationMode", "EnumerateEPR") -> msg:
//    <?xml version="1.0" encoding="UTF-8"?>
//    <ns0:Envelope xmlns:ns0="http://www.w3.org/2003/05/soap-envelope" xmlns:ns1="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:ns2="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xmlns:ns3="http://schemas.xmlsoap.org/ws/2004/09/enumeration">
//     <ns0:Header>
//      <ns1:Action ns0:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/enumeration/Enumerate</ns1:Action>
//      <ns1:To ns0:mustUnderstand="true">https://192.168.128.41:443/wsman</ns1:To>
//      <ns1:MessageID ns0:mustUnderstand="true">uuid:0590dc5f-369b-4ed3-9da4-8461a6f66dae</ns1:MessageID>
//      <ns1:ReplyTo>
//       <ns1:Address ns0:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</ns1:Address>
//      </ns1:ReplyTo>
//      <ns2:ResourceURI ns0:mustUnderstand="true">http://my.resource/url</ns2:ResourceURI>
//      <ns2:OptionSet>
//       <ns2:Option Name="EnumerationMode">EnumerateEPR</ns2:Option>
//      </ns2:OptionSet>
//     </ns0:Header>
//     <ns0:Body>
//      <ns3:Enumerate>
//       <ns2:OptimizeEnumeration/>
//       <ns2:MaxElements>100</ns2:MaxElements>
//      </ns3:Enumerate>
//     </ns0:Body>
//    </ns0:Envelope>
//
// Why not use a map?  Because that would make it much more painful to handle
// arrays, and I am not that interested in adding evem more magic.
func (m *Message) Options(opts ...string) *Message {
	if len(opts)%2 != 0 {
		panic("message.Options passed an odd number of args!")
	}
	optset := dom.Elem("OptionSet", NS_WSMAN)
	for i := 0; i < len(opts); i += 2 {
		elem := dom.ElemC("Option", NS_WSMAN, opts[i+1]).Attr("Name", "", opts[i])
		optset.AddChild(elem)
	}
	m.SetHeader(optset)
	return m
}

func (m *Message) ResourceURI(resource string) *Message {
	m.SetHeader(Resource(resource))
	return m
}

// Selectors are used to target the resource that Get, Put, and Invoke
// should work with.  They work like Options does, except they add a SelectorSet
// element with Selectors instead of Options.
func (m *Message) Selectors(args ...string) *Message {
	if len(args)%2 != 0 {
		panic("message.Selectors passed an odd number of args!")
	}
	selset := dom.Elem("SelectorSet", NS_WSMAN)
	for i := 0; i < len(args); i += 2 {
		elem := dom.ElemC("Selector", NS_WSMAN, args[i+1]).Attr("Name", "", args[i])
		selset.AddChild(elem)
	}
	m.SetHeader(selset)
	return m
}

// Parameters sets the parameters for an invoke call.
// It takes an even number of strings, which should be key:value pairs.
// It works alot like Options, except it adds the parameters to the Body
// in the format that WSMAN expects parameter elements to be in.
func (m *Message) Parameters(args ...string) *Message {
	if len(args)%2 != 0 {
		panic("message.Selectors passed an odd number of args!")
	}
	resource := m.GetHeader(dom.Elem("Action", NS_WSA))
	if resource == nil {
		panic("message has no Action")
	}
	resourceNS := string(resource.Content)
	idx := strings.LastIndex(resourceNS, "/")
	if idx == -1 {
		panic("Action is malformed!")
	}
	resourceName := fmt.Sprintf("%s_INPUT", resourceNS[idx+1:])
	resourceNS = resourceNS[:idx]

	paramSet := dom.Elem(resourceName, resourceNS)
	for i := 0; i < len(args); i += 2 {
		elem := dom.ElemC(args[i], resourceNS, args[i+1])
		paramSet.AddChild(elem)
	}
	m.SetBody(paramSet)
	return m
}

// Send sends a message to the endpoint of the Client it was
// constructed with, and returns either the Message that was
// returned, or an error statung what went wrong.
func (m *Message) Send() (*Message, error) {
	res, err := m.client.Post(m.Message)
	if err != nil {
		return nil, err
	}
	msg := &Message{Message: res}
	if msg.replyHelper != nil {
		err = msg.replyHelper(m, msg)
	}
	return msg, nil
}
