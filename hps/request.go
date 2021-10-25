package hps

type Request struct {
	URI     string
	Headers string
	Body    []byte
	Method  string
	Scheme  string
	Error   error
}
