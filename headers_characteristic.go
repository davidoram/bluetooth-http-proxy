package main

import (
	"fmt"
	"log"

	"github.com/davidoram/bluetooth-http-proxy/hps"
	"github.com/go-ble/ble"
)

// NewHeadersChar ...
func (svrCtx *ServerContext) NewHeadersChar() *ble.Characteristic {
	c := ble.NewCharacteristic(hps.HTTPHeadersID)

	c.HandleWrite(ble.WriteHandlerFunc(func(req ble.Request, rsp ble.ResponseWriter) {
		log.Printf("Headers: Wrote %s", string(req.Data()))
		svrCtx.Request.Headers = string(req.Data())
	}))

	c.HandleRead(ble.ReadHandlerFunc(func(req ble.Request, rsp ble.ResponseWriter) {
		if svrCtx.Response == nil {
			log.Printf("Headers: Read <empty> (no response)")
			fmt.Fprint(rsp, "")
		}
		log.Printf("Headers: Read %s", string(svrCtx.Response.Headers))
		fmt.Fprintf(rsp, "%s", svrCtx.Response.Headers)
	}))
	return c
}
