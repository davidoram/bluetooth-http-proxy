# bluetooth-http-proxy

Bluetooth low energy HTTP Proxy Service (HPS)

To compile:

```
GOOS=linux GOARCH=arm64 go build -o btserver *.go
GOOS=linux GOARCH=arm64 go build -o btclient client/*.go
```



Test using `https://github.dev/go-ble/ble`

```
sudo ./blesh_lnx sh
scan -name davidoram/HPS -dup=false
connect <id>
status
discover
status
disconnect
```

Unit tests

```
go test -v hps/*
```