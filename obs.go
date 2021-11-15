package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/events"
	"github.com/muesli/streamdeck"
)

func obsEventLoop(ctx context.Context, errCh chan error, client *goobs.Client, d *streamdeck.Device, cfg *Config, idx uint8) {
	offset := idx * 5
	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping OBS intance %d event listener", idx)
			return

		case e := <-client.IncomingEvents:
			switch ee := e.(type) {
			case *events.StreamStarted:
				// Switch stream button to 'on' image
				if err := d.SetImage(offset+0, imgs["live"][true]); err != nil {
					errCh <- err
					continue
				}

			case *events.StreamStopped:
				// Switch stream button to 'off' image
				if err := d.SetImage(offset+0, imgs["live"][false]); err != nil {
					errCh <- err
					continue
				}

			case *events.SwitchScenes:
				// Reset keys to current selected scene state
				current, err := getNumberForScene(client, ee.SceneName)
				if err != nil {
					errCh <- err
					continue
				}
				for i := uint8(1); i <= 3; i++ {
					if i == current {
						// this one is active
						if err := d.SetImage(offset+i, imgs[fmt.Sprintf("scene%d", i)][true]); err != nil {
							errCh <- err
							continue
						}
					} else {
						// this one is not
						if err := d.SetImage(offset+i, imgs[fmt.Sprintf("scene%d", i)][false]); err != nil {
							errCh <- err
							continue
						}
					}
				}

			case *events.StreamStatus:
				// Stream is live and periodically telling us the status..
				// 'swallow' event...

			case *events.SourceMuteStateChanged:
				// Audio/Microphone toggle
				if ee.SourceName == cfg.MicSource {
					if err := d.SetImage(offset+4, imgs["mic"][!ee.Muted]); err != nil {
						errCh <- err
						continue
					}
				}

			case *events.Error:
				errCh <- ee.Err
				continue

			default:
				// just ignore for now..
				//log.Printf("Unhandled event: %#v", e.GetUpdateType())
			}
		}
	}
}

// This function will return the number of the active scene
func getActiveSceneNumber(client *goobs.Client) (uint8, error) {
	currSceneResp, err := client.Scenes.GetCurrentScene()
	if err != nil {
		return 0, err
	}

	activeScene := currSceneResp.Name
	resp, err := client.Scenes.GetSceneList()
	if err != nil {
		return 0, err
	}
	for idx, s := range resp.Scenes {
		if s.Name == activeScene {
			return uint8(idx + 1), nil
		}
	}
	return 0, errors.New("unexpected: could not find active stream index")
}

// Fetch the name of the scene by number (1,2,3...)
func getSceneNameByNumber(client *goobs.Client, nr uint8) (string, error) {
	resp, err := client.Scenes.GetSceneList()
	if err != nil {
		return "", err
	}
	for idx, s := range resp.Scenes {
		if uint8(idx) == nr-1 {
			return s.Name, nil
		}
	}

	return "", errors.New("unexpected: could not find name for this scene number")
}

func getNumberForScene(client *goobs.Client, name string) (uint8, error) {
	resp, err := client.Scenes.GetSceneList()
	if err != nil {
		return 0, err
	}
	for idx, s := range resp.Scenes {
		if s.Name == name {
			return uint8(idx + 1), nil
		}
	}

	return 0, errors.New("unexpected: could not find name for this scene number")
}
