package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/requests/sources"
	"github.com/andreykaipov/goobs/api/requests/streaming"
	"github.com/muesli/streamdeck"
)

// This is the main loop that listens to, and controls the streamdeck
func streamdeckLoop(ctx context.Context, errCh chan error, clients Connections, d *streamdeck.Device, cfg *Config) {
	kch, err := d.ReadKeys()
	if err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping streamdeck listener")
			return

		case k, ok := <-kch:
			if !ok {
				continue
			}

			row := k.Index / 5       // Row (corresponds to OBS client)
			offset := row * 5        // key offset (5 keys per row)
			kIdx := k.Index - offset // Key (column) Index -> maps to 'function'

			if k.Pressed && kIdx == 0 {
				// Toggle OBS instance 1 Live stream on/off
				resp, err := clients[row].Streaming.GetStreamingStatus()
				if err != nil {
					errCh <- err
					continue
				}
				if resp.Streaming {
					clients[row].Streaming.StopStreaming()
				} else {
					clients[row].Streaming.StartStreaming(&streaming.StartStreamingParams{})
				}
			}
			if k.Pressed && kIdx > 0 && kIdx < 4 {
				// Toggle OBS instance[row] Scene 1..3
				if err := toggleCurrentScene(clients[row], kIdx); err != nil {
					errCh <- err
					continue
				}
			}
			if k.Pressed && kIdx == 4 {
				// Toggle between microphone and music audio source
				if err := toggleCurrentAudioSource(clients[row], cfg); err != nil {
					errCh <- err
					continue
				}
			}
		}
	}
}

// This function will find a streamdeck device by its serial number
func getStreamdeckDeviceBySerial(serial string) (*streamdeck.Device, error) {
	// Populate list with streamdeck devices
	devs, err := streamdeck.Devices()
	if err != nil {
		return nil, err
	}

	// We should at least find 1 streamdeck..
	if len(devs) == 0 {
		return nil, errors.New("no streamdeck devices found")
	}

	// Loop through all streamdecks
	for _, d := range devs {
		log.Printf("Found: %s (%d DPI, %d keys) - %s\n", d.ID, d.DPI, d.Keys, d.Serial)
		if strings.EqualFold(d.Serial, serial) {
			log.Println("Streamdeck selected")
			return &d, nil
		}
	}

	return nil, fmt.Errorf("could not find streamdeck with serial: %s", serial)
}

// setupDeck initializes the entire deck by querying OBS status
// and setting the buttons accordingly
func setupDeck(clients Connections, d *streamdeck.Device, cfg *Config) error {
	for idx, client := range clients {
		// Key offset (0-4 row 1, 5-9 row 2, 10-14 row 3)
		offset := idx * 5

		// Key 0 -> Stream on/off
		sstat, err := client.Streaming.GetStreamingStatus()
		if err != nil {
			return fmt.Errorf("could not get OBS streaming status: %s", err)
		}
		if err := d.SetImage(offset+0, imgs["live"][sstat.Streaming]); err != nil {
			return fmt.Errorf("could not set streamdeck image on key 0: %s", err)
		}

		// Keys 1,2,3 -> Scene 1..3 selection
		current, err := getActiveSceneNumber(client)
		if err != nil {
			return err
		}
		for i := uint8(1); i <= 3; i++ {
			if i == current {
				// this one is active
				if err := d.SetImage(uint8(offset+i), imgs[fmt.Sprintf("scene%d", i)][true]); err != nil {
					return fmt.Errorf("could not set streamdeck image on key %d: %s", i, err)
				}
			} else {
				// this one is not
				if err := d.SetImage(uint8(offset+i), imgs[fmt.Sprintf("scene%d", i)][false]); err != nil {
					return fmt.Errorf("could not set streamdeck image on key %d: %s", i, err)
				}
			}
		}

		// Key 4 -> microphone/music toggle
		resp, err := client.Sources.GetVolume(&sources.GetVolumeParams{Source: cfg.MicSource})
		if err != nil {
			return err
		}

		if err := d.SetImage(offset+4, imgs["mic"][!resp.Muted]); err != nil {
			return fmt.Errorf("could not set streamdeck image on key 4: %s", err)
		}
	}
	return nil
}

// Toggles the button for the scene depending wheter it's active or not
func toggleCurrentScene(client *goobs.Client, sceneNumber uint8) error {
	cs, err := getActiveSceneNumber(client)
	if err != nil {
		return err
	}
	if cs != sceneNumber {
		// Scene is not currently active
		name, err := getSceneNameByNumber(client, sceneNumber)
		if err != nil {
			return err
		}
		if _, err = client.Scenes.SetCurrentScene(&scenes.SetCurrentSceneParams{SceneName: name}); err != nil {
			return err
		}
	}

	return nil
}

func toggleCurrentAudioSource(client *goobs.Client, cfg *Config) error {
	// First get the currently active audio source
	resp, err := client.Sources.GetVolume(&sources.GetVolumeParams{Source: cfg.MicSource})
	if err != nil {
		return err
	}
	if resp.Muted {
		// Microphone is muted (inactive); mute audio/music and unmute microphone
		client.Sources.SetMute(&sources.SetMuteParams{Source: cfg.AudioSource, Mute: true})
		client.Sources.SetMute(&sources.SetMuteParams{Source: cfg.MicSource, Mute: false})
	} else {
		// Microphone not muted; mute microphone and unmute audio/music
		client.Sources.SetMute(&sources.SetMuteParams{Source: cfg.MicSource, Mute: true})
		client.Sources.SetMute(&sources.SetMuteParams{Source: cfg.AudioSource, Mute: false})
	}
	return nil
}
