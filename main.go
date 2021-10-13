package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/davidoram/bluetool-http-proxy/hps"
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
		log.Fatalf("can't new device : %s", err)
	}
	ble.SetDefaultDevice(d)

	req := Request{}

	testSvc := ble.NewService(hps.HpsServiceID)
	testSvc.AddCharacteristic(req.NewURIChar())
	// testSvc.AddCharacteristic(lib.NewEchoChar())

	if err := ble.AddService(testSvc); err != nil {
		log.Fatalf("can't add service: %s", err)
	}

	// Advertise for specified durantion, or until interrupted by user.
	fmt.Printf("Advertising for %s...\n", *du)
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), *du))
	chkErr(ble.AdvertiseNameAndServices(ctx, "Gopher", testSvc.UUID))
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
