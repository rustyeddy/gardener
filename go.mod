module github.com/rustyeddy/gardener

go 1.24.5

require (
	github.com/rustyeddy/devices v0.0.2
	github.com/rustyeddy/otto v0.0.10
)

replace github.com/rustyeddy/otto => ../otto

replace github.com/rustyeddy/devices => ../devices

require (
	github.com/creack/goselect v0.1.2 // indirect
	github.com/eclipse/paho.mqtt.golang v1.5.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/maciej/bme280 v0.2.0 // indirect
	github.com/mochi-mqtt/server/v2 v2.7.9 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/rs/xid v1.4.0 // indirect
	github.com/warthog618/go-gpiocdev v0.9.1 // indirect
	go.bug.st/serial v1.6.4 // indirect
	golang.org/x/exp v0.0.0-20251009144603-d2f985daa21b // indirect
	golang.org/x/image v0.23.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	periph.io/x/conn/v3 v3.7.2 // indirect
	periph.io/x/devices/v3 v3.7.4 // indirect
	periph.io/x/host/v3 v3.8.5 // indirect
)
