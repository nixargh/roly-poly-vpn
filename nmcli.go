package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

func nmcliGetActiveConnections(excludeBridge bool) []string {
	output := basher("nmcli -f NAME,TYPE -t connection show --active", "")
	var connections, filteredConnections []string

	if len(output) > 0 {
		connections = strings.Split(output, "\n")
	}
	clog.WithFields(log.Fields{"raw_connections": connections}).Debug("All active connections found.")

	for i := 0; i < len(connections)-1; i++ {
		splitCon := strings.Split(connections[i], ":")
		name := splitCon[0]
		cType := splitCon[1]
		clog.WithFields(log.Fields{"name": name, "type": cType}).Debug("Connections name and type.")

		if cType == "bridge" && excludeBridge == true {
			continue
		}
		filteredConnections = append(filteredConnections, name)
	}

	clog.WithFields(log.Fields{"connections": filteredConnections, "excludeBridge": excludeBridge}).Debug("Filtered active connections found.")
	return filteredConnections
}

func nmcliConnectionActive(config string) bool {
	connections := nmcliGetActiveConnections(false)
	index := slices.Index(connections, config)

	if index == -1 {
		return false
	} else {
		return true
	}
}

func nmcliConnectionUpPasswd(password string, passcode string, config string) {
	clog.WithFields(log.Fields{"config": config}).Info("Starting VPN connection.")

	passwdFile := "/tmp/roly-poly-vpn.nmcli.passwd"
	fullPassword := fmt.Sprintf("vpn.secrets.password:\"%v%v\"", password, passcode)

	err := os.WriteFile(passwdFile, []byte(fullPassword), 0600)
	if err != nil {
		clog.WithFields(log.Fields{"file": passwdFile, "error": err}).Fatal("Can't create temporary passwd file for nmcli.")
	}

	cmd := fmt.Sprintf("nmcli connection up %v passwd-file %v", config, passwdFile)
	basher(cmd, password)
	clog.WithFields(log.Fields{"config": config}).Info("VPN is connected.")

	os.Remove(passwdFile)
}

func nmcliConnectionUpAsk(password string, passcode string, config string) {
	var cmd string

	clog.WithFields(log.Fields{"config": config}).Info("Starting VPN connection.")

	// Update VPN config to ask password every time
	cmd = fmt.Sprintf("nmcli connection mod %v vpn.secrets 'password-flags=2'", config)
	basher(cmd, "")

	// Answer to password request interactively
	fullpass := fmt.Sprintf("\"%v%v\"", password, passcode)
	cmd = fmt.Sprintf("nmcli connection mod %v vpn.secrets password=%v", config, fullpass)
	basher(cmd, fullpass)

	clog.WithFields(log.Fields{"config": config}).Info("VPN is connected.")
}

func nmcliConnectionUp(config string) {
	clog.WithFields(log.Fields{"config": config}).Info("Starting VPN connection.")

	cmd := fmt.Sprintf("nmcli connection up %v", config)
	basher(cmd, "")
	clog.WithFields(log.Fields{"config": config}).Info("VPN is connected.")
}

func nmcliConnectionUpdatePasswordFlags(config string, value int) {
	var cmd string

	clog.WithFields(log.Fields{
		"config":         config,
		"password-flags": value,
	}).Debug("Updating VPN connection with a new password-flags.")

	cmd = fmt.Sprintf("nmcli connection mod %v +vpn.data 'password-flags=%d'", config, value)
	basher(cmd, "")

	clog.WithFields(log.Fields{"config": config}).Debug("VPN password-flags is updated.")
}

func nmcliConnectionUpdatePassword(password string, passcode string, config string) {
	var cmd string

	clog.WithFields(log.Fields{"config": config}).Info("Updating VPN connection with a new password.")

	// Update VPN config with a newly generated password
	fullpass := fmt.Sprintf("\"%v%v\"", password, passcode)
	cmd = fmt.Sprintf("nmcli connection mod %v vpn.secrets password=%v", config, fullpass)
	basher(cmd, fullpass)

	clog.WithFields(log.Fields{"config": config}).Info("VPN config is updated.")
}

func nmcliConnectionDown(config string) {
	clog.WithFields(log.Fields{"config": config}).Info("Stopping VPN connection.")
	cmd := fmt.Sprintf("nmcli connection down %v", config)
	basher(cmd, "")
}
