package main

import (
	"github.com/kkyr/fig"
)

type Config struct {
	Debug     bool   `fig:"debug"`
	Serial    string `fig:"streamdeck_serial"`
	Instances []struct {
		Port     string `fig:"port"`
		Password string `fig:"password"`
	} `fig:"instances" validate:"required"`
	MicSource   string `fig:"mic_source_name" validate:"required"`
	AudioSource string `fig:"audio_source_name" validate:"required"`
}

func loadConfig() *Config {
	var cfg Config
	if err := fig.Load(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}
