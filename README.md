# bluetooth-http-proxy - NOT WORKING

Bluetooth low energy HTTP Proxy Service (HPS)


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
