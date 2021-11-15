package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/andreykaipov/goobs"
)

type Connections map[uint8]*goobs.Client

func main() {
	// Parse config file
	cfg := loadConfig()
	if err := loadAssets(); err != nil {
		panic(err)
	}

	// Find the streamdeck device through configured serial
	d, err := getStreamdeckDeviceBySerial(cfg.Serial)
	if err != nil {
		panic(err)
	}

	// Open connection to streamdeck
	if err = d.Open(); err != nil {
		panic(err)
	}
	defer d.Close()

	// Get firmware revision and reset device
	ver, err := d.FirmwareVersion()
	if err != nil {
		panic(err)
	}
	if err = d.Reset(); err != nil {
		panic(err)
	}
	log.Printf("Opened device with serial %s (firmware %s)\n", d.Serial, ver)

	// Connections list
	var clients Connections = make(map[uint8]*goobs.Client)

	// Open connections to OBS Websockets
	for idx, icfg := range cfg.Instances {
		log.Printf("Connecting to instance %d on port %s", idx, icfg.Port)
		client, err := goobs.New(
			"localhost:"+icfg.Port,
			goobs.WithPassword(icfg.Password),
			goobs.WithDebug(cfg.Debug),
		)
		if err != nil {
			panic(err)
		}
		clients[uint8(idx)] = client
	}

	// Setup deck
	if err = setupDeck(clients, d, cfg); err != nil {
		panic(err)
	}

	// error channel for receiving errors from the event loops
	errCh := make(chan error, 5)

	// Create context and start event loops
	ctx, cancel := context.WithCancel(context.Background())
	for idx, client := range clients {
		// Goroutine with event listener for each OBS client
		go obsEventLoop(ctx, errCh, client, d, cfg, idx)
	}
	// Single goroutine to listen for streamdeck keypresses
	go streamdeckLoop(ctx, errCh, clients, d, cfg)

	// Wait for OS signal (ctrl-c)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

waitloop:
	for {
		select {
		case <-sigs:
			break waitloop

		case e := <-errCh:
			log.Printf("event error: %s", e)
		}
	}

	// Received a signal, stop the program..
	log.Println("Halting")
	cancel()

	// Give goroutines some time to finish
	time.Sleep(1 * time.Second)
	_ = d.Reset()
}
