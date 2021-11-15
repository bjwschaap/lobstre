package main

import (
	"image"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
)

var (
	imgs = make(map[string]map[bool]image.Image)
)

func loadAssets() error {
	log.Println("Finding image assets")
	err := filepath.Walk("./assets", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process files with .png extension
		if !info.IsDir() && filepath.Ext(info.Name()) == ".png" {
			log.Printf("Reading asset: %s", info.Name())
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				return err
			}

			nameParts := strings.Split(strings.Split(info.Name(), ".")[0], "_")

			if imgs[nameParts[0]] == nil {
				imgs[nameParts[0]] = make(map[bool]image.Image)
			}
			tf, err := strconv.ParseBool(nameParts[1])
			if err != nil {
				return err
			}
			imgs[nameParts[0]][tf] = resize.Resize(72, 72, img, resize.Lanczos3)
			log.Printf("Asset: %s/%s processed", nameParts[0], nameParts[1])
		}
		return nil
	})

	return err
}
