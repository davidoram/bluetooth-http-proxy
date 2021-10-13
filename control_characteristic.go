package main

import (
	"log"

	"github.com/davidoram/bluetooth-http-proxy/hps"
	"github.com/go-ble/ble"
)

// NewControlChar ...
func (hreq *Request) NewControlChar() *ble.Characteristic {
	c := ble.NewCharacteristic(hps.ControlPointUUID)

	c.HandleWrite(ble.WriteHandlerFunc(func(req ble.Request, rsp ble.ResponseWriter) {
		log.Printf("URI: Wrote %s", string(req.Data()))
		hreq.URI = string(req.Data())
	}))

	return c
}
