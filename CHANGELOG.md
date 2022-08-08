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

## [1.2.0] - 8/4/2022
### Fixed
-driven/laundry/csc_laundryview.go - standardized base api url from CSC so the switch to production endpoints can be made
### Added
-Endpoint courses/studentcourses was added to return classes and their locations for students for selected semester
-driven/courses/uiuc_courses.go
-Endpont /termsessions/listcurrent was added to return a list of currently selectable term sessions
-driven/terms/uiuc_termsessions.go
-driver/rest/termsessionapi.go
-core/model/termsessions.go


### Changed
-core/interfaces.go - GetStudentCourses added to Services interface
-core/interfaces.go - Defined student courses interface
-core/services.go - added getStudentCourses implementation to application
-core/interfaces.go - GetTermSessions added to Services interface
-core/interfaces.go - Defined TermSessions interface
-core/services.go - added getTermSessions implementation to application
-driver/rest/coursesapi.go
-core/model/courses.go
-drive/adapter.go - added routes for /courses/studentcourses and termsessions/listcurrent
-main.go - added term session adapter to application
-application.go - added a term session adapter
-services.go - implemented getStudentCourses and getTermSessions

## [1.2.1] - 8/4/2022
### Fixed
### Added

### Changed
-core/model/courses.go - added CourseReferenceNumber property to CourseSection object
-core/driven/courses/uiuc_courses.go - map data from campus json to CourseReferenceNumber property

## [1.2.2] - 8/8/2022
### Fixed
--uiuc_laundryview.go(322) - {"subscription-id": "uic-chicago", "key-type": "primaryKey" } to {"subscription-id": "uiuc", "key-type": "primaryKey" }
--uiuc_laundryview.go - changed code to handle new data sent by getsubscription vendor call
--uiuc_laundryview.go - added error handling for missing subscription key or request token
### Added
### Changed