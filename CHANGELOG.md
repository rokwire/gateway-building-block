# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

[2.4.5] - 2024-04-12
### Fixed
- Fix delete event context deadline  [#73] (https://github.com/rokwire/gateway-building-block/issues/73)

[2.4.4] - 2024-04-11
### Fixed
- Improve Get Legacy Events API [#70](https://github.com/rokwire/gateway-building-block/issues/70)

[2.4.3] - 2024-04-08
### Fixed
- Fix the Webtools events import [#68](https://github.com/rokwire/gateway-building-block/issues/68)

[2.4.2] - 2024-04-01
### Fixed
- Fix add to webtools blacklist [#65](https://github.com/rokwire/gateway-building-block/issues/65)

[2.4.1] - 2024-04-01
### Fixed
- Fix webtools blacklist APIs [#62](https://github.com/rokwire/gateway-building-block/issues/62)

[2.4.0] - 2024-03-29
### Added
- Ability to block/blacklist specific Webtools events [#57](https://github.com/rokwire/gateway-building-block/issues/57)

[2.3.2] - 2024-03-27
- Increase webtools transaction timeout

[2.3.1] - 2024-03-21
### Fixed
- Handle cost, tags, target and location on the Webtools import [#54](https://github.com/rokwire/gateway-building-block/issues/54)

[2.3.0] - 2024-03-18
### Added
- Delete tps events API [#52](https://github.com/rokwire/gateway-building-block/issues/52)
- Create tps events API [#47](https://github.com/rokwire/gateway-building-block/issues/47)

## [2.2.1] - 2024-03-06
### Fixed
- Fix daily timer [#49](https://github.com/rokwire/gateway-building-block/issues/49)

## [2.2.0] - 2024-02-08
### Added
- WebTools events handling [#39](https://github.com/rokwire/gateway-building-block/issues/39)

[2.1.0] - 2024-02-07
- added successteam end point
- added successteam/pcp end point
- added successteam/adivsors end point

## [2.0.14] - 2023-12-06
### Fixed
- fixed typo in adapter.go from /laundry/reqeustservice to /laundry/requestservice
- changed allowed method on /laundry/requestservice to POST from GET

## [2.0.12]
### Added
- changed datatype of lat/long build coordinates to long 64
- changed auth token type expected by wayfining endpoints from client.auth to client.standard

## [2.0.7] - 2023-05-05
### Fixed
- Fix permissions [#26](https://github.com/rokwire/gateway-building-block/issues/26)

## [2.0.6] - 2023-05-03
### Fixed
- Fix versioning issues

## [2.0.5] - 2023-05-03
### Changed
- Added host information to create and update results

## [2.0.4] - 2023-05-02
### Changed
- Convert time slot and advisor times to UTC
- filter units by college code (based on provider_id)

## [2.0.3] - 2023-05-01
### Added
- appointments end points and interfaces

### Changed
- updated old code to new building block template model


## [1.2.6] - 2023-03-09
### Fixed
- Security vulnerability in golang.org/x/text/language
- Security vulnerability in golang.org/x/crypto
- Security vulnerability in golang/org/x/net

## [1.2.4] - 2023-02-21 
### Fixed
- Security vulnerability in Go prior to 1.19.1. Switched project to 1.20.1 and docker image to golang:1.20.1-buster

## [1.2.3] - 2023-02-15
### Fixed
- uiuc_termsessions.go  - fixed wrong term session id being returned for future fall term #8
### Added
### Changed
- core/interfaces.go - GetGiesCourses added to Services interface
- core/interfaces.go - Defined GiesCourses interface
- core/services.go - added getGiesCourses implementation to application

## [0.1.0] - 2021-09-03
### Added
- Initial implementation





