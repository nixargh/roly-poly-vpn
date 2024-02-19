package main

import (
	"bufio"
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

var version string = "1.3.2"

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
	flag.BoolVar(&showVersion, "version", false, "Show version")

	flag.Parse()

	// Setup logging
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
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
	clog.Info("Hint: Use 'nmcli connection' to find out your config names.")
	config = manageParameter("config", config, false)
	password = manageParameter("password", password, true)
	otpSecret = manageParameter("otpSecret", otpSecret, true)

	go waitForDeath(config)

	sleepSeconds := 5
	clog.WithFields(log.Fields{"sleepSeconds": sleepSeconds}).Info("Starting the main loop.")
	for {
		active := nmcliConnectionActive(config)
		if active == false {
			// Check whether any network connection is active
			activeConns := nmcliGetActiveConnections(true)
			if len(activeConns) > 0 {
				clog.WithFields(log.Fields{"config": config}).Info("VPN connection isn't active. Starting.")

				passcode := GeneratePassCode(otpSecret)
				clog.WithFields(log.Fields{"passcode": passcode}).Info("Got a new pass code.")

				// Update VPN config to store password only for current user
				nmcliConnectionUpdatePasswordFlags(config, 1)

				nmcliConnectionUpdatePassword(password, passcode, config)

				nmcliConnectionUp(config)

				/* Update VPN config to ask password every time.
				That should prevent NM reconections with an old password. */
				nmcliConnectionUpdatePasswordFlags(config, 2)
			} else {
				clog.Info("No active connection found thus posponding VPN connection.")
			}
		}
		clog.WithFields(log.Fields{"config": config, "sleepSeconds": sleepSeconds}).Debug("Connection is active. Sleeping.")

		// Sleep for a minute
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}
}

func manageParameter(parameter string, parameterValue string, hide bool) string {
	service := "roly-poly-vpn"
	var err error

	// If value is empty - read from keyring or ask
	if parameterValue == "" {
		parameterValue, err = keyring.Get(service, parameter)

		if err == nil && parameterValue != "" {
			clog.WithFields(log.Fields{"parameter": parameter}).Info("Got parameter value from keyring.")
			return parameterValue
		}

		fmt.Printf("New '%v' value: ", parameter)

		if hide {
			bytespw, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				log.Fatal(err)
				clog.WithFields(log.Fields{
					"parameter": parameter,
					"error":     err,
				}).Fatal("Reading hidden parameter value from cmd failed.")
			}
			parameterValue = string(bytespw)
		} else {
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			err := scanner.Err()
			if err != nil {
				log.Fatal(err)
				clog.WithFields(log.Fields{
					"parameter": parameter,
					"error":     err,
				}).Fatal("Reading parameter value from cmd failed.")
			}
			parameterValue = scanner.Text()
		}
		fmt.Print("\n")
	}

	// Save value gotten as flag or asked
	err = keyring.Set(service, parameter, parameterValue)

	if err != nil {
		clog.WithFields(log.Fields{
			"parameter": parameter,
			"error":     err,
		}).Fatal("Can't save password to keyring.")
	}

	clog.WithFields(log.Fields{"parameter": parameter}).Info("Parameter's value saved to keyring.")
	return parameterValue
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
