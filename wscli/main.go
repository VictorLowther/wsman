package main

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
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/VictorLowther/simplexml/dom"
	"github.com/VictorLowther/wsman"
)

const (
	_ = iota
	soapFault
	transportError
	argError
)

var Endpoint, Username, Password, Action, Method, ResourceURI string
var useDigest, debug, optimizeEnum, useStdin bool
var selStr, optStr, paramStr string
var timeout int64

func init() {
	flag.StringVar(&Endpoint, "e", "", "The WSMAN endpoint to communicate with. Right now, only URLs are accepted.")
	flag.StringVar(&Username, "u", "", "The username to authenticate with")
	flag.StringVar(&Password, "p", "", "The password to authenticate with")
	flag.BoolVar(&useDigest, "d", false, "Use digest authentication instead of basic auth")
	flag.BoolVar(&debug, "D", false, "Run the WSMAN client in debug mode")
	flag.StringVar(&Action, "a", "Identify", `The WSMAN Action to perform. Can be one of :
      Identify
      Enumerate
      EnumerateEPR
      Get
      Put
      Create
      Delete
      Invoke
      Any URL for a custom WSMAN Action`)
	flag.BoolVar(&optimizeEnum, "q", false, "Optimize returning items from an Emumerate or EnumerateEPR")
	flag.BoolVar(&useStdin, "i", false, "Read body of request from stdin")
	flag.StringVar(&ResourceURI, "r", "", "The ResourceURI for the action")
	flag.StringVar(&Method, "m", "", "The method to invoke if the action is Invoke")
	flag.StringVar(&selStr, "s", "", "The comma-seperated list of selector:value pairs")
	flag.StringVar(&optStr, "o", "", "The comma-seperated set of WSMAN option:value pairs")
	flag.StringVar(&paramStr, "x", "", "The comma-seperated list of parameter:value pairs for Invoke actions")
	flag.Int64Var(&timeout, "t", 60, "The number of seconds to wait for a response from the WSMAN endpoint")
}

func handleSlice(p string) []string {
	res := []string{}
	if len(p) == 0 {
		return res
	}
	segments := strings.Split(p, ",")
	for _, s := range segments {
		segment := strings.SplitN(s, ":", 2)
		if len(segment) != 2 {
			log.Printf("Segment %s does not have 2 : seperated elements!", s)
			os.Exit(argError)
		}
		for _, t := range segment {
			res = append(res, strings.TrimSpace(t))
		}
	}
	return res
}

func getStdin() *dom.Element {
	doc, err := dom.Parse(os.Stdin)
	if err != nil {
		log.Printf("Failed to parse XML doc on stdin: %v\n", err)
		os.Exit(argError)
	}
	return doc.Root()
}

func main() {
	flag.Parse()
	Selectors := handleSlice(selStr)
	Options := handleSlice(optStr)
	Parameters := handleSlice(paramStr)
	if Endpoint == "" || Action == "" {
		flag.Usage()
		os.Exit(argError)
	}
	if len(flag.Args()) > 0 {
		fmt.Printf("%v", flag.Args())
		os.Exit(argError)
	}
	client := wsman.NewClient(Endpoint, Username, Password, useDigest)
	client.Debug = debug
	client.OptimizeEnum = optimizeEnum
	client.Timeout = (time.Duration(timeout) * time.Second)
	var msg *wsman.Message
	if Action == "Identify" {
		reply, err := client.Identify()
		if err != nil {
			log.Println(err.Error())
			os.Exit(transportError)
		}
		fmt.Println(reply.String())
		os.Exit(0)
	}
	if len(ResourceURI) == 0 {
		log.Printf("%s requires a resource URI passed in with -r\n", Action)
		os.Exit(argError)
	}
	switch Action {
	case "Enumerate":
		msg = client.Enumerate(ResourceURI)
	case "EnumerateEPR":
		msg = client.EnumerateEPR(ResourceURI)
	case "Get":
		msg = client.Get(ResourceURI)
	case "Put":
		msg = client.Put(ResourceURI)
	case "Create":
		msg = client.Create(ResourceURI)
	case "Delete":
		msg = client.Delete(ResourceURI)
	case "Invoke":
		if len(Method) == 0 {
			log.Printf("%s requires a method passed in with -m", Action)
			os.Exit(argError)
		}
		msg = client.Invoke(ResourceURI, Method)
	default:
		msg = client.NewMessage(Action)
		if len(ResourceURI) > 0 {
			msg.SetHeader(wsman.Resource(ResourceURI))
		}
	}
	if len(Options) > 0 {
		msg.Options(Options...)
	}
	if len(Selectors) > 0 {
		msg.Selectors(Selectors...)
	}
	if len(Parameters) > 0 {
		if Action == "Put" {
			msg.Values(Parameters...)
		} else {
			msg.Parameters(Parameters...)
		}
	}
	if useStdin {
		msg.SetBody(getStdin())
	}
	reply, err := msg.Send()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(transportError)
	}
	fmt.Println(reply.String())
	if reply.Fault() != nil {
		os.Exit(soapFault)
	}
	os.Exit(0)
}
