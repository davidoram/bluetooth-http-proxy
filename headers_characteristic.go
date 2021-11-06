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
	c.HandleWrite(ble.WriteHandlerFunc(svrCtx.writeHeaders))
	c.HandleRead(ble.ReadHandlerFunc(svrCtx.readHeaders))
	return c
}

func (svrCtx *ServerContext) writeHeaders(req ble.Request, rsp ble.ResponseWriter) {
	log.Printf("writeHeaders: %s", string(req.Data()))
	svrCtx.Request.Headers = string(req.Data())
	rsp.SetStatus(ble.ErrSuccess)
}

func (svrCtx *ServerContext) readHeaders(req ble.Request, rsp ble.ResponseWriter) {
	if svrCtx.Response == nil {
		log.Printf("readHeaders: <empty> (no response)")
		fmt.Fprint(rsp, "")
		rsp.SetStatus(ble.ErrSuccess)
		return
	}
	log.Printf("readHeaders: %s", string(svrCtx.Response.Headers))
	fmt.Fprintf(rsp, "%s", svrCtx.Response.Headers)
	rsp.SetStatus(ble.ErrSuccess)
}
