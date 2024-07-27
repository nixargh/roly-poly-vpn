package main

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/zalando/go-keyring"
)

type Config struct {
	Connection string `json:"connection"`
	Password   string `json:"password"`
	OtpSecret  string `json:"otp_secret"`
}

func (c *Config) read(instance string) {
	service := "roly-poly-vpn"

	config, err := keyring.Get(service, instance)

	if err == keyring.ErrNotFound {
		return
	}

	if err != nil {
		clog.WithFields(log.Fields{"instance": instance, "error": err}).Fatal("Failed to read config.")
	}

	clog.WithFields(log.Fields{"instance": instance}).Info("Got config from keyring.")

	err = json.Unmarshal([]byte(config), &c)
	if err != nil {
		clog.WithFields(log.Fields{"instance": instance}).Fatal("Failed to unmarshal config.")
	}
}

func (c *Config) write(instance string) {
	service := "roly-poly-vpn"

	jsonData, err := json.Marshal(c)

	if err != nil {
		clog.WithFields(log.Fields{
			"instance": instance,
			"error":    err,
		}).Error("Config convertion to JSON failed.")
	}

	jsonConfig := fmt.Sprintf(string(jsonData))

	err = keyring.Set(service, instance, jsonConfig)

	if err != nil {
		clog.WithFields(log.Fields{"instance": instance, "error": err}).Fatal("Failed to write config.")
	}

	clog.WithFields(log.Fields{"instance": instance}).Info("Wrote config to keyring.")
}
