package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/davidoram/bluetooth-http-proxy/hps"
)

func (svrCtx *ServerContext) ProxyRequest() error {
	log.Printf("ProxyRequest %s %s://%s", svrCtx.Request.Method, svrCtx.Request.Scheme, svrCtx.Request.URI)
	// Alloc space for the response, and a channel for response comms
	svrCtx.Response = &hps.Response{}

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest(svrCtx.Request.Method, fmt.Sprintf("%s://%s", svrCtx.Request.Scheme, svrCtx.Request.URI), bytes.NewReader(svrCtx.Request.Body))
	if err != nil {
		svrCtx.Error = err
		log.Printf("Error: Creating proxy http request %v", svrCtx.Error)
		svrCtx.responseChannel <- false
		return svrCtx.Error
	}

	// Headers
	if svrCtx.Request.Headers != "" {
		for _, h := range strings.Split(svrCtx.Request.Headers, "\n") {
			values := strings.Split(h, "=")
			if len(values) != 2 {
				log.Printf("Warn: ignoring invalid header %s", h)
				continue
			}
			req.Header.Add(values[0], values[1])
		}
	}

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		svrCtx.Response.Error = err
		log.Printf("Error: HTTP call failed: %v", svrCtx.Response.Error)
		svrCtx.Response.NotifyStatus = hps.NotifyStatus{
			StatusCode:       http.StatusBadGateway,
			HeadersReceived:  false,
			HeadersTruncated: false,
			BodyReceived:     false,
			BodyTruncated:    false,
		}
		svrCtx.responseChannel <- false
		return svrCtx.Response.Error
	}

	// Read Response Body
	svrCtx.Response.Body, svrCtx.Response.Error = ioutil.ReadAll(resp.Body)
	if svrCtx.Response.Error != nil {
		log.Printf("Error: Read response body failed, err %v", svrCtx.Response.Error)
		svrCtx.Response.NotifyStatus = hps.NotifyStatus{
			StatusCode:       http.StatusInternalServerError,
			HeadersReceived:  false,
			HeadersTruncated: false,
			BodyReceived:     false,
			BodyTruncated:    false,
		}
		svrCtx.responseChannel <- false
		return svrCtx.Response.Error
	}

	log.Printf("Call worked ok!")
	var trunc bool
	svrCtx.Response.Headers, trunc = hps.EncodeHeaders(resp.Header)
	svrCtx.Response.NotifyStatus = hps.NotifyStatus{
		StatusCode:       resp.StatusCode,
		HeadersReceived:  true,
		HeadersTruncated: trunc,
		BodyReceived:     len(svrCtx.Response.Body) > 0,
		BodyTruncated:    len(svrCtx.Response.Body) > hps.BodyMaxOctets,
	}
	svrCtx.responseChannel <- true
	return nil
}
