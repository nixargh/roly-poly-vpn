# Roly-Poly VPN
## Why?
It is a daemon designed to ease 2FA VPN usage with one time passwords at GNU/Linux systems.

## How it works
At first start it takes your permanent password, TOTP secret and Network Manager config name and places them into your keyring.
Then you can start it as a daemon and it will bring up that VPN connection using password and TOTP generated passcode. Also it will deactivate VPN connection at termination.

## Systemd
I want to use it as systemd service and I prepared a unit file plus a piece of audit config, but both methods of providing password (--ask, passwd-file) don't work when `roly-poly-vpn` is started by systemd.
I'm going to find another way or fix some of these methods but it doesn't work as expected right now. So run it from your session somehow.
