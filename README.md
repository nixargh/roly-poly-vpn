# Roly-Poly VPN
## Why?
It is a daemon designed to ease 2FA VPN usage with one time passwords at GNU/Linux systems.

## How it works
At first start it takes your permanent password, TOTP secret and Network Manager config name and places them into your keyring.
Then you can start it as a daemon and it will bring up that VPN connection using password and TOTP generated passcode. Also it will deactivate VPN connection at termination.

## Systemd
I want to use it as systemd service and I prepared a unit file plus a piece of audit config, but both methods of providing password (--ask, passwd-file) don't work when `roly-poly-vpn` is started by systemd.
I'm going to find another way or fix some of these methods but it doesn't work as expected right now. So run it from your session somehow.

## Installation
- Import your OpenVPN configuration to NetworkManager configuration.
- Set your login to the NM VPN config and set to "Ask password every time".
- Download from binary from [release page](https://github.com/nixargh/tired/releases).
- Set execution bit for binary: ```chmod +x ./roly-poly-vpn```
- Move somewhere to your **PATH**. At Ubuntu I prefer `~/.local/bin/` directory: ```mv ./roly-poly-vpn ~/.local/bin/```
- Run it and answer questions about NetworkManager VPN config name, ypur LDAP password and OTP secret.
If you make a mistake and want to change the value just run **roly-poly-vpn** with flag setting this secret and it will overwritten at your keyring. Or as alternative **seahorse** utility, which is a GUI keyring manager, could be used.
