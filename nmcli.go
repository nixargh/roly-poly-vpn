package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

func nmcliGetActiveConnections() []string {
	output := basher("nmcli -f NAME -t connection show --active", "")
	var connections []string

	if len(output) > 0 {
		connections = strings.Split(output, "\n")
	}

	clog.WithFields(log.Fields{"connections": connections}).Debug("Active connections found.")
	return connections
}

func nmcliConnectionActive(config string) bool {
	connections := nmcliGetActiveConnections()
	sort.Strings(connections)
	index := sort.SearchStrings(connections, config)

	if index == len(connections) {
		return false
	} else {
		return true
	}
}

func nmcliConnectionUpPasswd(password string, passcode string, config string) {
	clog.WithFields(log.Fields{"config": config}).Info("Starting VPN connection.")

	passwdFile := "/tmp/roly-poly-vpn.nmcli.passwd"
	fullPassword := fmt.Sprintf("vpn.secrets.password:%v%v", password, passcode)

	err := os.WriteFile(passwdFile, []byte(fullPassword), 0600)
	if err != nil {
		clog.WithFields(log.Fields{"file": passwdFile, "error": err}).Fatal("Can't create temporary passwd file for nmcli.")
	}

	cmd := fmt.Sprintf("nmcli connection up %v passwd-file %v", config, passwdFile)
	basher(cmd, password)

	os.Remove(passwdFile)
}

func nmcliConnectionUpAsk(password string, passcode string, config string) {
	clog.WithFields(log.Fields{"config": config}).Info("Starting VPN connection.")

	cmd := fmt.Sprintf("echo '%v%v' | nmcli connection up %v --ask", password, passcode, config)
	basher(cmd, password)
}

func nmcliConnectionDown(config string) {
	clog.WithFields(log.Fields{"config": config}).Info("Stopping VPN connection.")
	cmd := fmt.Sprintf("nmcli connection down %v", config)
	basher(cmd, "")
}
