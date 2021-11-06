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

	FindServerError = errors.New("HPS Server not found")

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

	addr      ble.Addr
	device    ble.Device
	bleClient ble.Client

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
	c.ResponseTimeout, _ = time.ParseDuration("10s")
	return &c
}

func (client *Client) filterByName(a ble.Advertisement) bool {
	// Default to search device with name specified by user
	match := strings.ToUpper(a.LocalName()) == strings.ToUpper(client.DeviceName)
	log.Printf("filterByName '%s' == '%s' match: %t", a.LocalName(), client.DeviceName, match)
	return match
}
func (client *Client) advHandler(a ble.Advertisement) {
	log.Printf("advHandler %s %s", a.LocalName(), a.Addr().String())
	client.addr = a.Addr()
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
	log.Printf("client.lastError %v", client.lastError)
	client.lastError = client.bleClient.WriteCharacteristic(uriChr, []byte(client.uri), true)
	if client.lastError != nil {
		log.Printf("Error setting URI")
		return *client.response, client.lastError
	}
	log.Printf("Set URI '%s' ok", client.uri)

	// Send Headers
	hdrChr := p.FindCharacteristic(ble.NewCharacteristic(HTTPHeadersID))
	if hdrChr == nil {
		client.lastError = FindHeadersChrError
		return *client.response, client.lastError
	}
	client.lastError = client.bleClient.WriteCharacteristic(hdrChr, []byte(client.headers.String()), true)
	if client.lastError != nil {
		log.Printf("Error setting Headers")
		return *client.response, client.lastError
	}
	log.Printf("Set Headers ok")

	// Send Body
	bodyChr := p.FindCharacteristic(ble.NewCharacteristic(HTTPEntityBodyID))
	if bodyChr == nil {
		client.lastError = FindBodyChrError
		return *client.response, client.lastError
	}
	client.lastError = client.bleClient.WriteCharacteristic(bodyChr, []byte(client.body), true)
	if client.lastError != nil {
		log.Printf("Error setting Body")
		return *client.response, client.lastError
	}
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
	client.lastError = client.bleClient.Subscribe(controlChr, true, client.controlHandler)
	if client.lastError != nil {
		log.Printf("Error subscribing to Control point")
		return *client.response, client.lastError
	}

	client.lastError = client.bleClient.WriteCharacteristic(controlChr, []byte{cpBuf}, true)
	if client.lastError != nil {
		log.Printf("Error setting Control point")
		return *client.response, client.lastError
	}
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
