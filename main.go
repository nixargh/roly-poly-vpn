package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/zalando/go-keyring"

	log "github.com/sirupsen/logrus"
	//	"github.com/pkg/profile"
)

var version string = "1.0.0"

var clog *log.Entry

func main() {
	//	defer profile.Start().Stop()

	var config string
	var password string
	var otpSecret string
	var debug bool
	var showVersion bool

	flag.StringVar(&config, "config", "", "VPN configuration name (use 'nmcli connection' to find out)")
	flag.StringVar(&password, "password", "", "VPN user password")
	flag.StringVar(&otpSecret, "otpSecret", "", "VPN OTP secret")
	flag.BoolVar(&debug, "debug", false, "Log debug messages")
	flag.BoolVar(&showVersion, "version", false, "FunVPN version")

	flag.Parse()

	// Setup logging
	log.SetOutput(os.Stdout)

	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if debug == true {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	clog = log.WithFields(log.Fields{
		"pid":     os.Getpid(),
		"thread":  "main",
		"version": version,
	})

	clog.Info("Let's have some fun with 2FA VPN via NM!")

	// Validate variables
	if config == "" {
		clog.Info("Hint: Use 'nmcli connection' to find out your config names.")
		config = promptForSecret("config")
	}

	if password == "" {
		password = promptForSecret("password")
	}

	if otpSecret == "" {
		otpSecret = promptForSecret("otpSecret")
	}

	go waitForDeath(config)

	sleepSeconds := 5
	clog.WithFields(log.Fields{"sleepSeconds": sleepSeconds}).Info("Starting then main loop.")
	for {
		active := nmcliConnectionActive(config)
		if active == false {
			// Check whether any network connection is active
			activeConns := nmcliGetActiveConnections()
			if len(activeConns) > 0 {
				clog.WithFields(log.Fields{"config": config}).Info("VPN connection isn't active. Starting.")

				passcode := GeneratePassCode(otpSecret)
				clog.WithFields(log.Fields{"passcode": passcode}).Info("Got a new pass code.")

				nmcliConnectionUpAsk(password, passcode, config)
			} else {
				clog.Info("No active connection found thus posponding VPN connection.")
			}
		}
		clog.WithFields(log.Fields{"config": config, "sleepSeconds": sleepSeconds}).Debug("Connection is active. Sleeping.")

		// Sleep for a minute
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}
}

func promptForSecret(secret string) string {
	service := "roly-poly-vpn"
	var secretValue string
	var err error

	secretValue, err = keyring.Get(service, secret)

	if err == nil && secretValue != "" {
		clog.WithFields(log.Fields{"secret": secret}).Info("Got secret value from keyring.")
		return secretValue
	}

	fmt.Printf("New '%v' value: ", secret)
	bytespw, _ := term.ReadPassword(int(syscall.Stdin))
	secretValue = string(bytespw)
	fmt.Print("\n")

	err = keyring.Set(service, secret, secretValue)

	if err != nil {
		clog.WithFields(log.Fields{"secret": secret, "error": err}).Fatal("Can't save password to keyring.")
	}

	clog.WithFields(log.Fields{"secret": secret}).Info("Secret saved to keyring.")
	return secretValue
}

func GeneratePassCode(secret string) string {
	passcode, err := totp.GenerateCodeCustom(secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})

	if err != nil {
		clog.Fatal("TOTP pass code generation failed.")
	}
	return passcode
}

func basher(command string, hide string) string {
	commandStr := command

	cmd, err := exec.Command("/bin/bash", "-c", command).Output()
	output := string(cmd)

	if hide != "" {
		commandStr = strings.Replace(commandStr, hide, "*****", -1)
	}

	clog.WithFields(log.Fields{"command": commandStr, "output": output}).Debug("Command output.")

	if err != nil {
		clog.WithFields(log.Fields{"command": commandStr, "error": err}).Fatal("Shell command failed.")
	}

	return output
}

func waitForDeath(config string) {
	clog.Info("Starting Wait For Death loop.")
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	for {
		time.Sleep(time.Duration(1) * time.Second)

		sig := <-cancelChan
		clog.WithFields(log.Fields{"signal": sig}).Info("Caught signal. Terminating.")

		nmcliConnectionDown(config)

		clog.WithFields(log.Fields{"signal": sig}).Info("We are good to go, see you next time!.")
		os.Exit(0)
	}
}
