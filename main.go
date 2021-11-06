package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/davidoram/bluetooth-http-proxy/hps"
	"github.com/pkg/errors"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

var (
	device = flag.String("device", hps.DeviceName, "device name")
	du     = flag.Duration("du", 5*time.Minute, "advertising duration, 0 for indefinitely")
)

func main() {
	flag.Parse()

	d, err := dev.NewDevice(*device)
	if err != nil {
		log.Fatalf("Can't new device, maybe permissions? : %s", err)
	}
	ble.SetDefaultDevice(d)

	svr := ServerContext{responseChannel: make(chan bool, 1)}

	hpsService := ble.NewService(hps.HpsServiceID)
	log.Printf("Service: %s", hps.HpsServiceID.String())
	hpsService.AddCharacteristic(svr.NewURIChar())
	hpsService.AddCharacteristic(svr.NewControlChar())
	hpsService.AddCharacteristic(svr.NewHeadersChar())
	hpsService.AddCharacteristic(svr.NewBodyChar())

	if err := ble.AddService(hpsService); err != nil {
		log.Fatalf("can't add service: %s", err)
	}

	// Advertise for specified durantion, or until interrupted by user.
	fmt.Printf("Advertising for %s...\n", *du)
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), *du))
	chkErr(ble.AdvertiseNameAndServices(ctx, hps.DeviceName, hpsService.UUID))
}

func chkErr(err error) {
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		fmt.Printf("done\n")
	case context.Canceled:
		fmt.Printf("canceled\n")
	default:
		log.Fatalf(err.Error())
	}
}
