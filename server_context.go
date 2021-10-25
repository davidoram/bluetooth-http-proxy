package main

import "github.com/davidoram/bluetooth-http-proxy/hps"

type ServerContext struct {
	Request  hps.Request
	Response *hps.Response

	// Wait for response on this channel
	responseChannel chan bool

	Error error
}
