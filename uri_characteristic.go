package main

import (
	"log"

	"github.com/davidoram/bluetooth-http-proxy/hps"
	"github.com/go-ble/ble"
)

// NewURIChar ...
func (svrCtx *ServerContext) NewURIChar() *ble.Characteristic {
	c := ble.NewCharacteristic(hps.URIUUID)
	c.HandleWrite(ble.WriteHandlerFunc(svrCtx.writeUri))
	return c
}

func (svrCtx *ServerContext) writeUri(req ble.Request, rsp ble.ResponseWriter) {
	log.Printf("writeUri: %s", string(req.Data()))
	svrCtx.Request.URI = string(req.Data())
}
