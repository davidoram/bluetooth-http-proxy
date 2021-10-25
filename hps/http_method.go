package hps

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	// See https://btprodspecificationrefs.blob.core.windows.net/assigned-values/16-bit%20UUID%20Numbers%20Document.pdf
	HTTPReserved      uint8 = 0x00
	HTTPGet           uint8 = 0x01
	HTTPHead          uint8 = 0x02
	HTTPPost          uint8 = 0x03
	HTTPPut           uint8 = 0x04
	HTTPDelete        uint8 = 0x05
	HTTPSGet          uint8 = 0x06
	HTTPSHead         uint8 = 0x07
	HTTPSPost         uint8 = 0x08
	HTTPSPut          uint8 = 0x09
	HTTPSDelete       uint8 = 0x0a
	HTTPRequestCancel uint8 = 0x0b
)

var (
	UnsupportedSchemeError     = errors.New("Unsupported scheme, valid values are http and https")
	UnsupportedHttpMethodError = errors.New("Unsupported method, valid values are GET, HEAD, POST, PUT, DELETE")
)

type DecodeHttpMethodError struct {
	Data byte
}

func (r *DecodeHttpMethodError) Error() string {
	return fmt.Sprintf("Unable to decode HTTP method from byte '%v'", r.Data)
}

type DecodeURLSchemeError struct {
	Data byte
}

func (r *DecodeURLSchemeError) Error() string {
	return fmt.Sprintf("Unable to decode URL scheme from byte '%v'", r.Data)
}

func DecodeHttpMethod(b byte) (string, error) {
	switch b {
	case HTTPGet, HTTPSGet:
		return http.MethodGet, nil
	case HTTPHead, HTTPSHead:
		return http.MethodHead, nil
	case HTTPPut, HTTPSPut:
		return http.MethodPut, nil
	case HTTPPost, HTTPSPost:
		return http.MethodPost, nil
	case HTTPDelete, HTTPSDelete:
		return http.MethodDelete, nil
	default:
		return "", &DecodeHttpMethodError{b}
	}
}

func DecodeURLScheme(b byte) (string, error) {
	switch b {
	case HTTPSGet, HTTPSHead, HTTPSPut, HTTPSPost, HTTPSDelete:
		return "https", nil
	case HTTPGet, HTTPHead, HTTPPut, HTTPPost, HTTPDelete:
		return "http", nil
	default:
		return "", &DecodeURLSchemeError{b}
	}
}

func EncodeMethodScheme(method, scheme string) (uint8, error) {
	switch strings.ToUpper(strings.Trim(method, " ")) {
	case http.MethodGet:
		switch scheme {
		case "http":
			return HTTPGet, nil
		case "https":
			return HTTPSGet, nil
		default:
			return 0, UnsupportedSchemeError
		}
	case http.MethodHead:
		switch scheme {
		case "http":
			return HTTPHead, nil
		case "https":
			return HTTPSHead, nil
		default:
			return 0, UnsupportedSchemeError
		}
	case http.MethodPost:
		switch scheme {
		case "http":
			return HTTPPost, nil
		case "https":
			return HTTPSPost, nil
		default:
			return 0, UnsupportedSchemeError
		}
	case http.MethodPut:
		switch scheme {
		case "http":
			return HTTPPut, nil
		case "https":
			return HTTPSPut, nil
		default:
			return 0, UnsupportedSchemeError
		}
	case http.MethodDelete:
		switch scheme {
		case "http":
			return HTTPDelete, nil
		case "https":
			return HTTPSDelete, nil
		default:
			return 0, UnsupportedSchemeError
		}
	default:
		return 0, UnsupportedHttpMethodError
	}
}
