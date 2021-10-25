package main

import (
	"log"

	"github.com/davidoram/bluetooth-http-proxy/hps"
	"github.com/go-ble/ble"
)

// NewControlChar ...
func (svrCtx *ServerContext) NewControlChar() *ble.Characteristic {
	c := ble.NewCharacteristic(hps.ControlPointUUID)

	var err error
	c.HandleWrite(ble.WriteHandlerFunc(func(req ble.Request, rsp ble.ResponseWriter) {
		data := req.Data()
		log.Printf("control: Wrote %d", uint(data[0]))
		svrCtx.Request.Method, svrCtx.Error = hps.DecodeHttpMethod(data[0])
		if svrCtx.Error != nil {
			log.Printf("Error: Write control %v", err)
			return
		}

		svrCtx.Request.Scheme, svrCtx.Error = hps.DecodeURLScheme(data[0])
		if svrCtx.Error != nil {
			log.Printf("Error: Decode scheme %v", err)
			return
		}

		// Make the API call in the background
		go svrCtx.ProxyRequest()

	}))

	c.HandleNotify(ble.NotifyHandlerFunc(func(req ble.Request, n ble.Notifier) {
		log.Printf("control: Notification subscribed")
		for {
			select {
			case <-n.Context().Done():
				log.Printf("control: Notification unsubscribed")
				return
			case proxiedOk := <-svrCtx.responseChannel:
				log.Printf("control: Notify: proxiedOk %t", proxiedOk)
				if proxiedOk {

					if _, err := n.Write(svrCtx.Response.NotifyStatus.Encode()); err != nil {
						// Client disconnected prematurely before unsubscription.
						log.Printf("control: Failed to notify : %s", err)
						return
					}
					// TODO return body + headers
				}
			}
		}
	}))
	return c
}
