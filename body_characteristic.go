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
	c.HandleWrite(ble.WriteHandlerFunc(svrCtx.readBody))
	c.HandleRead(ble.ReadHandlerFunc(svrCtx.writeBody))
	return c
}

func (svrCtx *ServerContext) writeBody(req ble.Request, rsp ble.ResponseWriter) {
	log.Printf("writeBody: %s", string(req.Data()))
	svrCtx.Request.Body = req.Data()
	rsp.SetStatus(ble.ErrSuccess)
}

func (svrCtx *ServerContext) readBody(req ble.Request, rsp ble.ResponseWriter) {
	if svrCtx.Response == nil {
		log.Printf("readBody: <empty> (no response)")
		fmt.Fprint(rsp, "")
		rsp.SetStatus(ble.ErrSuccess)
		return
	}
	log.Printf("readBody: %s", string(svrCtx.Response.Body))
	fmt.Fprintf(rsp, "%s", svrCtx.Response.Body)
	rsp.SetStatus(ble.ErrSuccess)
}
