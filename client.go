// Package wsman implements a simple WSMAN client interface.
// It assumes you are talking to WSMAN over http(s) and using
// basic authentication.
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
	"crypto/tls"
	"fmt"
	"github.com/VictorLowther/simplexml/dom"
	"github.com/VictorLowther/soap"
	"io/ioutil"
	"net/http"
	"time"
)

// Client is a thin wrapper around http.Client.
type Client struct {
	http.Client
	target, username, password string
}

// NewClient creates a new wsman.Client.
//
// target must be a URL, and username and password must be
// the username and password to authenticate to the controller with.
// If username or password are empty, we will not try to authenticate.
func NewClient(target, username, password string) *Client {
	res := &Client{
		target:   target,
		username: username,
		password: password,
	}
	res.Timeout = 10 * time.Second
	res.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return res
}

func (c *Client) Endpoint() string {
	return c.target
}

func (c *Client) Post(msg *soap.Message) (response *soap.Message, err error) {
	req, err := http.NewRequest("POST", c.target, msg.Reader())
	if err != nil {
		return nil, err
	}
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	req.Header.Add("content-type", soap.ContentType)
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		b, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("wsman.Client: post recieved %v\n'%v'", res.Status, string(b))
	}
	response, err = soap.Parse(res.Body)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Identify performs a basic WSMAN IDENTIFY call.
// The response will provide the version of WSMAN the endpoint
// speaks, along with some details about the WSMAN endpoint itself.
// Note that identify uses soap.Message directly instead of wsman.Message.
func (c *Client) Identify() (*soap.Message, error) {
	message := soap.NewMessage()
	message.SetBody(dom.Elem("Identify", NS_WSMID))
	return c.Post(message)
}
