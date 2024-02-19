# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [1.3.2] - 2024-01-24
### Fixed
- `nmcli.go` issue when password contains a single quote sign.

## [1.3.1] - 2024-01-24
### Fixed
- `nmcli.go` fix the way NM VPN config is being updated.
- `main.go` use full timestamp even when TTY is attached.

## [1.3.0] - 2024-01-21
### Added
- `nmcli.go` new function **nmcliConnectionUpdatePassword** that updates NM VPN config with a generated password for only current user.
- `nmcli.go` new function **nmcliConnectionUp** that only brings up a NM VPN connection with previously set password.

### Changed
- `main.go` use **nmcliConnectionUpdatePassword** + **nmcliConnectionUp** to bring up VPN connection.
- `systemd/roly-poly-vpn.service` update to run as per user service.
- `README.md` update **Installation** section.

## [1.2.0] - 2024-01-18
### Changed
- `main.go` rewrite secret at keyring if set by flag.
- `README.md` describe installation process.

## [1.1.1] - 2023-07-27
### Fixed
- `nmcli.go` **nmcliConnectionActive** search of element in slice.

## [1.1.0] - 2023-07-26
### Added
- Filter out **bridge** type of NetworkManager connection during check of active connection before VPN activation.

## [1.0.1] - 2023-07-07
### Added
- First release.
