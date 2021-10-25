package hps

// Package hps provides HPS/HTTP client  implementations
import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

var (
	UnknownError           = errors.New("Unknown error")
	ConnectionTimeoutError = errors.New("Connection timeout")
	DisconnectedError      = errors.New("Disconnected")
	ResponseTimeoutError   = errors.New("Response timeout")

	FindURIChrError     = errors.New("Missing URI characterstic")
	FindHeadersChrError = errors.New("Missing Headers characterstic")
	FindBodyChrError    = errors.New("Missing Body characterstic")
	FindControlChrError = errors.New("Missing Control characterstic")
)

type Client struct {
	DebugLog        bool
	DeviceName      string
	ConnectTimeout  time.Duration
	ResponseTimeout time.Duration

	uri     string
	u       *url.URL
	headers ArrayStr
	body    string
	method  string

	responseChannel chan bool
	response        *Response
	lastError       error

	foundServer bool
	device      ble.Device
	bleClient   ble.Client

	// uriChr, hdrsChr, bodyChr, controlChr, statusChr *gatt.Characteristic

	done chan bool
}

func MakeClient(deviceName string) *Client {

	c := Client{
		DeviceName:      deviceName,
		lastError:       UnknownError,
		responseChannel: make(chan bool, 1),
		done:            make(chan bool, 1),
		response:        &Response{},
	}
	if deviceName == "" {
		c.DeviceName = DeviceName
	}
	c.ConnectTimeout, _ = time.ParseDuration("5s")
	c.ResponseTimeout, _ = time.ParseDuration("5s")
	return &c
}

func (client *Client) filterByName(a ble.Advertisement) bool {
	// Default to search device with name specified by user
	match := strings.ToUpper(a.LocalName()) == strings.ToUpper(client.DeviceName)
	log.Printf("filterByName %s == %s match: %t", a.LocalName(), client.DeviceName, match)
	return match
}
func (client *Client) advHandler(a ble.Advertisement) {
}

func (client *Client) connect() error {
	var d ble.Device
	d, client.lastError = dev.NewDevice("default")
	if client.lastError != nil {
		return client.lastError
	}
	ble.SetDefaultDevice(d)

	client.bleClient = nil
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), client.ConnectTimeout))
	client.lastError = ble.Scan(ctx, false, client.advHandler, client.filterByName)
	if client.lastError == nil {
		log.Printf("Error scanning: %v", client.lastError)
		return client.lastError
	}
	client.bleClient, client.lastError = ble.Connect(ctx, client.filterByName)
	if client.lastError == nil {
		// Function to automatically cleanup on disconnect
		go func() {
			<-client.bleClient.Disconnected()
			client.bleClient = nil
			log.Printf("\n%s disconnected\n", client.bleClient.Addr().String())
		}()
	}
	return client.lastError
}

func (client *Client) controlHandler(req []byte) {
	var ns NotifyStatus
	ns, client.lastError = DecodeNotifyStatus(req)
	if client.lastError != nil {
		log.Printf("Error decoding notify status err: %v", client.lastError)
		client.responseChannel <- false
		return
	}
	client.response = &Response{NotifyStatus: ns}
	client.responseChannel <- true
}

func (client *Client) Do(uri, body, method string, headers ArrayStr) (Response, error) {
	client.uri = uri
	client.u, client.lastError = url.Parse(client.uri)
	if client.lastError != nil {
		log.Printf("Error Parsing URI, err: %v", client.lastError)
		return *client.response, client.lastError
	}
	client.method = method
	client.body = body
	client.headers = headers

	// Connect
	client.lastError = client.connect()
	if client.lastError != nil {
		log.Printf("Error connecting, err: %v", client.lastError)
		return *client.response, client.lastError
	}
	log.Printf("Connected ok to addr: %s", client.bleClient.Addr().String())
	defer client.bleClient.CancelConnection()

	// retrieve HPS profile
	var p *ble.Profile
	p, client.lastError = client.bleClient.DiscoverProfile(true)
	if client.lastError != nil {
		log.Printf("Error connecting, err: %v", client.lastError)
		return *client.response, client.lastError
	}
	log.Printf("Got profile")

	// Send URI
	uriChr := p.FindCharacteristic(ble.NewCharacteristic(URIUUID))
	if uriChr == nil {
		client.lastError = FindURIChrError
		return *client.response, client.lastError
	}
	uriChr.SetValue([]byte(client.uri))
	log.Printf("Set URI ok")

	// Send Headers
	hdrChr := p.FindCharacteristic(ble.NewCharacteristic(HTTPHeadersID))
	if hdrChr == nil {
		client.lastError = FindHeadersChrError
		return *client.response, client.lastError
	}
	hdrChr.SetValue([]byte(client.headers.String()))
	log.Printf("Set Headers ok")

	// Send Body
	bodyChr := p.FindCharacteristic(ble.NewCharacteristic(HTTPEntityBodyID))
	if bodyChr == nil {
		client.lastError = FindBodyChrError
		return *client.response, client.lastError
	}
	bodyChr.SetValue([]byte(client.body))
	log.Printf("Set Body ok")

	// Send control point (eg: GET/PUT/POST/... + HTTP/HTTPS)
	// This will trigger the server to proxy the request.
	var cpBuf uint8
	cpBuf, client.lastError = EncodeMethodScheme(client.method, client.u.Scheme)
	if client.lastError != nil {
		return *client.response, client.lastError
	}
	controlChr := p.FindCharacteristic(ble.NewCharacteristic(ControlPointUUID))
	if controlChr == nil {
		client.lastError = FindControlChrError
		return *client.response, client.lastError
	}

	// Add a handler to receive notifications
	client.bleClient.Subscribe(controlChr, true, client.controlHandler)

	controlChr.SetValue([]byte{cpBuf})
	log.Printf("Set Control point ok")

	// Wait for response, or timeout
	select {
	case res := <-client.responseChannel:
		log.Printf("Got response %b", res)
	case <-time.After(client.ResponseTimeout):
		log.Printf("Response timeout")
		client.lastError = ResponseTimeoutError
	}

	return *client.response, client.lastError
}

// func (client *Client) scanPeriodically(d gatt.Device) {
// 	log.Printf("start periodic scan")

// 	// Create a new context
// 	ctx := context.Background()
// 	// Create a new context, with its cancellation function
// 	// from the original context
// 	ctx, _ = context.WithTimeout(ctx, client.ConnectTimeout)

// 	timeout := false
// 	for !client.foundServer && !timeout {
// 		select {
// 		case <-ctx.Done():
// 			log.Printf("Connection timeout")
// 			timeout = true
// 			client.lastError = ConnectionTimeoutError
// 			client.done <- false
// 		default:
// 			d.Scan([]gatt.UUID{}, false)
// 			time.Sleep(time.Millisecond * 100)
// 		}
// 	}
// 	log.Printf("stop periodic scan")
// }

// func (client *Client) onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
// 	if p.Name() != client.DeviceName {
// 		log.Printf("Skip peripheral_id: %s, name: %s", p.ID(), p.Name())
// 		return
// 	}
// 	client.foundServer = true

// 	// Stop scanning once we've got the peripheral we're looking for.
// 	log.Printf("Found HPS server")
// 	p.Device().StopScanning()
// 	p.Device().Connect(p)
// }

// func (client *Client) onPeriphConnected(p gatt.Peripheral, err error) {
// 	log.Printf("connected")

// 	if client.lastError = p.SetMTU(500); client.lastError != nil {
// 		log.Printf("Error setting MTU, err: %v", client.lastError)
// 		return
// 	}

// 	// Discovery services
// 	var ss []*gatt.Service
// 	ss, client.lastError = p.DiscoverServices(nil)
// 	if client.lastError != nil {
// 		log.Printf("Error Discover services, err: %v", client.lastError)
// 		return
// 	}

// 	for _, s := range ss {
// 		if s.UUID().Equal(gatt.MustParseUUID(HpsServiceID)) {
// 			client.hpsService = s
// 			err := client.parseService(p)
// 			if err != nil {
// 				log.Printf("Warning Parsing service, err: %v", err)
// 				continue
// 			}
// 			client.lastError = client.callService(p)
// 			if client.lastError != nil {
// 				log.Printf("Error Calling service, err: %v", client.lastError)
// 			}
// 			break
// 		}
// 	}
// }

// func (client *Client) onPeriphDisconnected(p gatt.Peripheral, err error) {
// 	log.Printf("disconnected")
// 	client.done <- false
// }

// func (client *Client) parseService(p gatt.Peripheral) error {
// 	log.Printf("parse service")

// 	// Discovery characteristics
// 	var cs []*gatt.Characteristic
// 	cs, client.lastError = p.DiscoverCharacteristics(nil, client.hpsService)
// 	if client.lastError != nil {
// 		return client.lastError
// 	}
// 	for _, c := range cs {
// 		// log.Printf("discovered characteristic name: %s", c.Name())
// 		switch c.UUID().String() {
// 		case gatt.UUID16(HTTPURIID).String():
// 			client.uriChr = c
// 		case gatt.UUID16(HTTPHeadersID).String():
// 			client.hdrsChr = c
// 		case gatt.UUID16(HTTPEntityBodyID).String():
// 			client.bodyChr = c
// 		case gatt.UUID16(HTTPControlPointID).String():
// 			client.controlChr = c
// 		case gatt.UUID16(HTTPStatusCodeID).String():
// 			client.statusChr = c
// 		}

// 		// Discovery descriptors
// 		ds, err := p.DiscoverDescriptors(nil, c)
// 		if err != nil {
// 			log.Printf("Warn discover descriptors, err: %v", err)
// 			continue
// 		}

// 		for _, d := range ds {
// 			// Read descriptor (could fail, if it's not readable)
// 			_, err := p.ReadDescriptor(d)
// 			if err != nil {
// 				log.Printf("Warn reading descriptor: %s, err: %v", d.Name(), err)
// 				continue
// 			}
// 		}

// 		// Subscribe the characteristic, if possible.
// 		if (c.Properties() & (gatt.CharNotify | gatt.CharIndicate)) != 0 {
// 			f := func(c *gatt.Characteristic, b []byte, err error) {
// 				if c.UUID().Equal(gatt.UUID16(HTTPStatusCodeID)) {
// 					var ns NotifyStatus
// 					ns, client.lastError = DecodeNotifyStatus(b)
// 					if client.lastError != nil {
// 						log.Printf("Error decoding notify status err: %v", client.lastError)
// 						return
// 					}
// 					log.Printf("got headers?       %t", ns.HeadersReceived)
// 					log.Printf("headers truncated? %t", ns.HeadersTruncated)
// 					log.Printf("body received?     %t", ns.BodyReceived)
// 					log.Printf("body truncated?    %t", ns.BodyTruncated)
// 					log.Printf("status:  %d", ns.StatusCode)
// 					client.response = &Response{NotifyStatus: ns}
// 					client.responseChannel <- true
// 				}
// 			}
// 			if client.lastError = p.SetNotifyValue(c, f); client.lastError != nil {
// 				log.Printf("Error subscribing to notifications, err: %v", client.lastError)
// 				continue
// 			}
// 		}

// 	}
// 	return nil
// }

// func (client *Client) callService(p gatt.Peripheral) error {
// 	defer p.Device().CancelConnection(p)

// 	log.Printf("call service")

// 	urlStr := fmt.Sprintf("%s%s", client.u.Host, client.u.EscapedPath())
// 	log.Printf("write method + uri: %s %s", client.method, client.u.String())
// 	client.lastError = p.WriteCharacteristic(client.uriChr, []byte(urlStr), true)
// 	if client.lastError != nil {
// 		return client.lastError
// 	}

// 	log.Printf("write headers: %v", client.headers)
// 	client.lastError = p.WriteCharacteristic(client.hdrsChr, []byte(client.headers.String()), true)
// 	if client.lastError != nil {
// 		return client.lastError
// 	}

// 	log.Printf("write body: %s", client.body)
// 	client.lastError = p.WriteCharacteristic(client.bodyChr, []byte(client.body), true)
// 	if client.lastError != nil {
// 		return client.lastError
// 	}

// 	var code uint8
// 	code, client.lastError = EncodeMethodScheme(client.method, client.u.Scheme)
// 	if client.lastError != nil {
// 		return client.lastError
// 	}
// 	log.Printf("write control: %d", code)
// 	client.lastError = p.WriteCharacteristic(client.controlChr, []byte{code}, false)
// 	if client.lastError != nil {
// 		return client.lastError
// 	}

// 	log.Printf("waiting for notification, timeout after %v", client.ResponseTimeout)
// 	time.AfterFunc(client.ResponseTimeout, func() {
// 		log.Printf("timeout expired, no notification received")
// 		client.responseChannel <- false
// 	})
// 	gotResponse := <-client.responseChannel
// 	if gotResponse {
// 		client.response.Body, client.lastError = p.ReadCharacteristic(client.bodyChr)
// 		if client.lastError != nil {
// 			return client.lastError
// 		}
// 		log.Printf("body:    %s", string(client.response.Body))

// 		client.response.Headers, client.lastError = p.ReadCharacteristic(client.hdrsChr)
// 		if client.lastError != nil {
// 			return client.lastError
// 		}
// 		log.Printf("headers: %v", formatHeaders(client.response.Headers))

// 		// all done no errors!
// 		client.lastError = nil
// 		client.done <- true
// 	}
// 	return nil
// }

// func formatHeaders(b []byte) []string {
// 	s := string(b)
// 	sa := strings.Split(s, "\n")
// 	return sa
// }
