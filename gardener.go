package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/bme280"
	"github.com/rustyeddy/devices/button"
	"github.com/rustyeddy/devices/oled"
	"github.com/rustyeddy/devices/relay"
	"github.com/rustyeddy/devices/vh400"
	"github.com/rustyeddy/otto/messenger"
	"github.com/rustyeddy/otto/server"
	"github.com/rustyeddy/otto/station"
)

type Gardener struct {
	*messenger.Messenger
	*station.StationManager
	*server.Server
	*station.DeviceManager // is this really needed?

	soil    *vh400.VH400
	env     *bme280.BME280
	pump    *relay.Relay
	on      *button.Button
	off     *button.Button
	display *oled.OLED

	Done chan any
}

func (g *Gardener) GetDeviceManager() *station.DeviceManager {
	if g.DeviceManager == nil {
		g.DeviceManager = station.NewDeviceManager()
	}
	return g.DeviceManager
}

var (
	pinmap = map[string]int{
		"on":   17,
		"off":  27,
		"soil": 22,
		"pump": 5,
		"env":  6,
	}
)

func (g *Gardener) Init() {
	g.Messenger = messenger.GetMessenger()
	g.DeviceManager = g.GetDeviceManager()
	g.StationManager = station.NewStationManager()
	g.Server = server.GetServer()
	g.Done = make(chan any)

	g.initButtons()
	g.initPump()
	g.initEnv()
	g.initDisplay()
	g.InitSoil()
}

func (g *Gardener) initButtons() {
	var err error
	g.on, err = button.New("on", pinmap["on"])
	if err != nil {
		panic(err)
	}
	g.DeviceManager.Add(g.on)
	g.on.RegisterEventHandler(func(evt *devices.DeviceEvent) {
		switch evt.Type {
		case devices.DeviceEventRisingEdge:
			slog.Info("button pressed", "button", "on", "action", "pump_on")
			g.Messenger.Pub("d/on", []byte("on"))
		}
	})

	g.off, err = button.New("off", pinmap["off"])
	if err != nil {
		panic(err)
	}
	g.DeviceManager.Add(g.off)
	g.off.RegisterEventHandler(func(evt *devices.DeviceEvent) {
		switch evt.Type {
		case devices.DeviceEventRisingEdge:
			slog.Info("button pressed", "button", "off", "action", "pump_off")
			g.Messenger.Pub("d/off", []byte("off"))
		}
	})
}

func (g *Gardener) InitSoil() {
	var err error
	g.soil, err = vh400.New("soil", pinmap["soil"])
	if err != nil {
		panic(err)
	}
	g.DeviceManager.Add(g.soil)
	cb := func(t time.Time) {
		value, err := g.soil.Get()
		if err != nil {
			slog.Error("soil sensor read failed", "error", err)
			return
		}
		slog.Info("soil moisture reading", "value", value)
		g.Messenger.Pub("d/soil", []byte(fmt.Sprintf("%5.2f", value)))
	}
	g.soil.StartTicker(10*time.Second, &cb)
}

func (g *Gardener) initEnv() {
	var err error
	g.env, err = bme280.New("env", "/dev/i2c-1", 0x76)
	if err != nil {
		panic(err)
	}
	g.DeviceManager.Add(g.env)
	ticker := func(t time.Time) {
		resp, err := g.env.Get()
		if err != nil {
			slog.Error("env sensor read failed", "error", err)
			return
		}
		slog.Info("env sensor reading",
			"temperature", resp.Temperature,
			"humidity", resp.Humidity,
			"pressure", resp.Pressure)

		jbuf, err := resp.JSON()
		if err != nil {
			slog.Error("env sensor marshal failed", "error", err)
			return
		}
		slog.Info("env sensor json", "data", string(jbuf))
		g.Messenger.Pub("d/env", jbuf)
	}
	g.env.StartTicker(10*time.Second, &ticker)
}

func (g *Gardener) initPump() {
	var err error
	g.pump, err = relay.New("pump", pinmap["pump"])
	if err != nil {
		panic(err)
	}
	g.Messenger.Sub("c/pump", g.pump.HandleMsg)
}

func (g *Gardener) initDisplay() {
	display, err := oled.New("c/lcd", 0x27, 1)
	if err != nil {
		panic(err)
	}
	display.Clear()

	// Register devices
	g.DeviceManager.Add(display)
}

func (g *Gardener) Start() {
	err := g.Messenger.Connect()
	if err != nil {
		slog.Error("gardener failed to connect to broker ", "error", err)
		return
	}

	topics := []string{"soil", "env", "on", "off", "pump", "display"}
	for _, topic := range topics {
		g.Sub(topic, g.MsgHandler)
	}
	if config.Mock {
		md := g.DeviceManager.GetDevice("soil")
		soil := md.(*vh400.VH400)
		g.emulator(soil)
	}

}

func (g *Gardener) MsgHandler(msg *messenger.Msg) error {
	slog.Info("MQTT [I]", "topic", msg.Topic, "value", msg.Data)

	switch msg.Topic {
	case "soil":
		fmt.Println("Got soil: ")

	case "env":
		fmt.Println("Got env: ")

	case "on", "off":
		fmt.Println("Got a button")

	default:
		slog.Error("unknown msg type", "topic", msg.Topic, "msg", msg)
	}
	fmt.Printf("msg: %#v\n", msg)
	return nil
}

func (g *Gardener) Stop() {
	// Implement stop logic if needed
	g.Done <- true
}

func (g *Gardener) emulator(soil *vh400.VH400) {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-g.Done:
				return // Exit the goroutine when done signal is received
			case _ = <-ticker.C:
				// Execute this code at each tick
				v, err := soil.Pin.Get()
				if err != nil {
					slog.Error("emulator failure", "error", err)
					continue
				}
				v += 0.02
				soil.Pin.Set(v)
			}
		}
	}()
}
