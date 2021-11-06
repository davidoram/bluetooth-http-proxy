package main

import (
	"log"

	"github.com/davidoram/bluetooth-http-proxy/hps"
	"github.com/go-ble/ble"
)

// NewControlChar ...
func (svrCtx *ServerContext) NewControlChar() *ble.Characteristic {
	c := ble.NewCharacteristic(hps.ControlPointUUID)
	c.HandleWrite(ble.WriteHandlerFunc(svrCtx.writeControl))
	c.HandleNotify(ble.NotifyHandlerFunc(svrCtx.notifyControl))
	c.HandleIndicate(ble.NotifyHandlerFunc(svrCtx.notifyControl))
	return c
}

func (svrCtx *ServerContext) writeControl(req ble.Request, rsp ble.ResponseWriter) {
	var err error
	data := req.Data()
	log.Printf("writeControl: %d", uint(data[0]))
	svrCtx.Request.Method, svrCtx.Error = hps.DecodeHttpMethod(data[0])
	if svrCtx.Error != nil {
		log.Printf("Error: writeControl decode HTTP method %v", err)
		rsp.SetStatus(ble.ErrReqNotSupp)
		return
	}

	svrCtx.Request.Scheme, svrCtx.Error = hps.DecodeURLScheme(data[0])
	if svrCtx.Error != nil {
		log.Printf("Error: writeControl decode scheme %v", err)
		rsp.SetStatus(ble.ErrReqNotSupp)
		return
	}

	// Make the API call in the background
	go svrCtx.ProxyRequest()
	rsp.SetStatus(ble.ErrSuccess)
}

func (svrCtx *ServerContext) notifyControl(req ble.Request, n ble.Notifier) {

	log.Printf("notifyControl: subscribed")
	for {
		select {
		case <-n.Context().Done():
			log.Printf("notifyControl: unsubscribed")
			return
		case proxiedOk := <-svrCtx.responseChannel:
			log.Printf("notifyControl: proxiedOk %t", proxiedOk)
			if proxiedOk {

				if _, err := n.Write(svrCtx.Response.NotifyStatus.Encode()); err != nil {
					// Client disconnected prematurely before unsubscription.
					log.Printf("notifyControl: Failed to notify : %s", err)
					return
				}
				// TODO return body + headers
			}
		}
	}
}
