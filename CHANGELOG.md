# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Prepare the project to become open source [#2](https://github.com/rokwire/gateway-building-block/issues/2)

## [0.1.0] - 2021-09-03
### Added
- Initial implementation

## [1.1.0] - 7/21/2022
### Fixed
### Added
-Endpoint courses/giescourses was added to return current semester classes for gies students
-driven/courses/uiuc_gies_courses.go
-driver/rest/coursesapi.go
-core/model/courses.go
### Changed
-core/interfaces.go - GetGiesCourses added to Services interface
-core/interfaces.go - Defined GiesCourses interface
-core/services.go - added getGiesCourses implementation to application
