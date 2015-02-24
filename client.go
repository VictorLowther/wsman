// Package wsman implements a simple WSMAN client interface.
// It assumes you are talking to WSMAN over http(s) and using
// basic authentication.
package wsman

import (
	"fmt"
	"github.com/VictorLowther/soap"
	"crypto/tls"
	"net/http"
	"strings"
	"time"
)

// Client is a thin wrapper around http.Client.
type Client struct {
	http.Client
	target, username, password string
}

const identifyDoc = `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"
xmlns:wsmid="http://schemas.dmtf.org/wbem/wsman/identity/1/wsmanidentity.xsd">
 <s:Header/>
 <s:Body>
  <wsmid:Identify/>
 </s:Body>
</s:Envelope>`

// NewClient creates a new wsman.Client.
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

func (c *Client) post(msg *soap.Message) (response *soap.Message, err error) {
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
		return nil, fmt.Errorf("wsman.Client: post recieved %v", res.Status)
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
func (c *Client) Identify() (*soap.Message, error) {
	message, err := soap.Parse(strings.NewReader(identifyDoc))
	if err != nil {
		panic(err)
	}
	return c.post(message)
}
