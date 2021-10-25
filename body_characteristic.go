package main

import (
	"fmt"
	"log"

	"github.com/davidoram/bluetooth-http-proxy/hps"
	"github.com/go-ble/ble"
)

// NewBodyChar ...
func (svrCtx *ServerContext) NewBodyChar() *ble.Characteristic {
	c := ble.NewCharacteristic(hps.HTTPEntityBodyID)

	c.HandleWrite(ble.WriteHandlerFunc(func(req ble.Request, rsp ble.ResponseWriter) {
		log.Printf("Body: Wrote %s", string(req.Data()))
		svrCtx.Request.Body = req.Data()
	}))

	c.HandleRead(ble.ReadHandlerFunc(func(req ble.Request, rsp ble.ResponseWriter) {
		if svrCtx.Response == nil {
			log.Printf("Body: Read <empty> (no response)")
			fmt.Fprint(rsp, "")
		}
		log.Printf("Body: Read %s", string(svrCtx.Response.Body))
		fmt.Fprintf(rsp, "%s", svrCtx.Response.Body)
	}))
	return c
}
