package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/otto/messanger"
	"github.com/rustyeddy/otto/utils"
)

type Config struct {
	StationName string
	Mock        bool
	Log         utils.LogConfig
	messanger.Config
}

var (
	config Config
)

func init() {
	flag.BoolVar(&config.Mock, "mock", false, "mock gpio")
	flag.StringVar(&config.Broker, "mqtt-broker", "otto", "MQTT broker address")
	flag.StringVar(&config.Username, "mqtt-username", "", "MQTT broker address")
	flag.StringVar(&config.Password, "mqtt-password", "", "MQTT broker address")
	flag.StringVar(&config.StationName, "station-name", "gardener", "station name")

	// Logging flags
	flag.StringVar(&config.Log.Level, "log-level", "info", "log level: debug, info, warn, error")
	flag.Var(&config.Log.Output, "log-output", "log output: stdout, stderr, file")
	flag.Var(&config.Log.Format, "log-format", "log format: text, json")
	flag.StringVar(&config.Log.FilePath, "log-file", "garden-station.log", "log file path (when log-output=file)")
	config.Log.Output.Set("file")
	config.Log.Format.Set("text")
}

func main() {
	flag.Parse()

	// Initialize structured logging
	_, err := utils.InitLoggerWithConfig(config.Log)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	slog.Info("starting garden-station",
		"station", config.StationName,
		"mock", config.Mock,
		"broker", config.Broker,
		"log_level", config.Log.Level,
		"log_output", config.Log.Output,
	)

	// Enable mocking in devices if mock flag is set
	if config.Mock {
		devices.SetMock(true)
	}

	gardener := &Gardener{}
	gardener.Init()
	go gardener.Start()

	// Handle OS signals and call Stop() for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		slog.Info("received signal, stopping gardener", "signal", sig)
		gardener.Stop()
	}()

	<-gardener.Done
	slog.Info("gardener stopped")
}
