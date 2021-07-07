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
	"crypto/md5"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/VictorLowther/simplexml/dom"
	"github.com/VictorLowther/soap"
)

type challenge struct {
	Username   string
	Password   string
	Realm      string
	CSRFToken  string
	Domain     string
	Nonce      string
	Opaque     string
	Stale      string
	Algorithm  string
	Qop        string
	Cnonce     string
	NonceCount int
}

func h(data string) string {
	hf := md5.New()
	io.WriteString(hf, data)
	return fmt.Sprintf("%x", hf.Sum(nil))
}

func kd(secret, data string) string {
	return h(fmt.Sprintf("%s:%s", secret, data))
}

func (c *challenge) ha1() string {
	return h(fmt.Sprintf("%s:%s:%s", c.Username, c.Realm, c.Password))
}

func (c *challenge) ha2(method, uri string) string {
	return h(fmt.Sprintf("%s:%s", method, uri))
}

func (c *challenge) resp(method, uri, cnonce string) (string, error) {
	c.NonceCount++
	if c.Qop == "auth" {
		if cnonce != "" {
			c.Cnonce = cnonce
		} else {
			b := make([]byte, 8)
			io.ReadFull(rand.Reader, b)
			c.Cnonce = fmt.Sprintf("%x", b)[:16]
		}
		return kd(c.ha1(), fmt.Sprintf("%s:%08x:%s:%s:%s",
			c.Nonce, c.NonceCount, c.Cnonce, c.Qop, c.ha2(method, uri))), nil
	} else if c.Qop == "" {
		return kd(c.ha1(), fmt.Sprintf("%s:%s", c.Nonce, c.ha2(method, uri))), nil
	}
	return "", fmt.Errorf("Alg not implemented")
}

// source https://code.google.com/p/mlab-ns2/source/browse/gae/ns/digest/digest.go#178
func (c *challenge) authorize(method, uri string) (string, error) {
	// Note that this is only implemented for MD5 and NOT MD5-sess.
	// MD5-sess is rarely supported and those that do are a big mess.
	if c.Algorithm != "MD5" {
		return "", fmt.Errorf("Alg not implemented")
	}
	// Note that this is NOT implemented for "qop=auth-int".  Similarly the
	// auth-int server side implementations that do exist are a mess.
	if c.Qop != "auth" && c.Qop != "" {
		return "", fmt.Errorf("Alg not implemented")
	}
	resp, err := c.resp(method, uri, "")
	if err != nil {
		return "", fmt.Errorf("Alg not implemented")
	}
	sl := []string{fmt.Sprintf(`username="%s"`, c.Username)}
	sl = append(sl, fmt.Sprintf(`realm="%s"`, c.Realm))
	sl = append(sl, fmt.Sprintf(`nonce="%s"`, c.Nonce))
	sl = append(sl, fmt.Sprintf(`uri="%s"`, uri))
	sl = append(sl, fmt.Sprintf(`response="%s"`, resp))
	if c.Algorithm != "" {
		sl = append(sl, fmt.Sprintf(`algorithm="%s"`, c.Algorithm))
	}
	if c.Opaque != "" {
		sl = append(sl, fmt.Sprintf(`opaque="%s"`, c.Opaque))
	}
	if c.Qop != "" {
		sl = append(sl, fmt.Sprintf("qop=%s", c.Qop))
		sl = append(sl, fmt.Sprintf("nc=%08x", c.NonceCount))
		sl = append(sl, fmt.Sprintf(`cnonce="%s"`, c.Cnonce))
	}
	return fmt.Sprintf("Digest %s", strings.Join(sl, ",")), nil
}

// origin https://code.google.com/p/mlab-ns2/source/browse/gae/ns/digest/digest.go#90
func (c *challenge) parseChallenge(input string) error {
	const ws = " \n\r\t"
	const qs = `"`
	s := strings.Trim(input, ws)
	if !strings.HasPrefix(s, "Digest ") {
		return fmt.Errorf("Challenge is bad, missing prefix: %s", input)
	}
	s = strings.Trim(s[7:], ws)
	sl := strings.Split(s, ",")
	c.Algorithm = "MD5"
	var r []string
	for i := range sl {
		r = strings.SplitN(sl[i], "=", 2)
		switch strings.TrimSpace(r[0]) {
		case "realm":
			c.Realm = strings.Trim(r[1], qs)
		case "domain":
			c.Domain = strings.Trim(r[1], qs)
		case "nonce":
			c.Nonce = strings.Trim(r[1], qs)
		case "opaque":
			c.Opaque = strings.Trim(r[1], qs)
		case "stale":
			c.Stale = strings.Trim(r[1], qs)
		case "algorithm":
			c.Algorithm = strings.Trim(r[1], qs)
		case "qop":
			//TODO(gavaletz) should be an array of strings?
			c.Qop = strings.Trim(r[1], qs)
		default:
			return fmt.Errorf("Challenge is bad, unexpected token: %s", sl)
		}
	}
	return nil
}

// Client is a thin wrapper around http.Client.
type Client struct {
	http.Client
	target, username, password     string
	useDigest, Debug, OptimizeEnum bool
	challenge                      *challenge
}

// NewClient creates a new wsman.Client.
//
// target must be a URL, and username and password must be the
// username and password to authenticate to the controller with.  If
// username or password are empty, we will not try to authenticate.
// If useDigest is true, we will try to use digest auth instead of
// basic auth.
func NewClient(target, username, password string, useDigest bool) (*Client, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target as url %v", err)
	}
	res := &Client{
		target:    target,
		username:  username,
		password:  password,
		useDigest: useDigest,
	}
	res.Timeout = 10 * time.Second
	res.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if res.useDigest {
		res.challenge = &challenge{Username: res.username, Password: res.password}
		resp, err := res.PostForm(res.target, nil)
		if err != nil {
			return nil, fmt.Errorf("Unable to perform digest auth with %s: %v", res.target, err)
		}
		if resp.StatusCode != 401 {
			return nil, fmt.Errorf("No digest auth at %s", res.target)
		}
		if err := res.challenge.parseChallenge(resp.Header.Get("WWW-Authenticate")); err != nil {
			return nil, fmt.Errorf("Failed to parse auth header %v", err)
		}
	}
	return res, nil
}

// Endpoint returns the endpoint that the Client will try to ocmmunicate with.
func (c *Client) Endpoint() string {
	return c.target
}

// Post overrides http.Client's Post method and adds digext auth handling
// and SOAP pre and post processing.
func (c *Client) Post(msg *soap.Message) (response *soap.Message, err error) {
	req, err := http.NewRequest("POST", c.target, msg.Reader())
	if err != nil {
		return nil, err
	}
	if c.username != "" && c.password != "" {
		if c.useDigest {
			auth, err := c.challenge.authorize("POST", c.target)
			if err != nil {
				return nil, fmt.Errorf("Failed digest auth %v", err)
			}
			req.Header.Set("Authorization", auth)
		} else {
			req.SetBasicAuth(c.username, c.password)
		}
	}
	req.Header.Add("content-type", soap.ContentType)
	if c.Debug {
		log.Printf("req:%#v\nbody:\n%s\n", req, msg.String())
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if c.useDigest && res.StatusCode == 401 {
		if c.Debug {
			log.Printf("Digest reauthorizing")
		}
		if err := c.challenge.parseChallenge(res.Header.Get("WWW-Authenticate")); err != nil {
			return nil, err
		}
		auth, err := c.challenge.authorize("POST", c.target)
		if err != nil {
			return nil, fmt.Errorf("Failed digest auth %v", err)
		}
		req, err = http.NewRequest("POST", c.target, msg.Reader())
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", auth)
		req.Header.Add("content-type", soap.ContentType)
		res, err = c.Do(req)
		if err != nil {
			return nil, err
		}
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
	if c.Debug {
		log.Printf("res: %#v\nbody:\n%s\n", res, response.String())
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
