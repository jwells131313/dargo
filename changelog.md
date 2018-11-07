# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Changed

## [1.0.0] - 2018-11-07
### Changed
- Fixed issue with locator shutdown not removing name

## [0.5.0] - 2018-09-30
### Changed
- BACKWARD BREAK:  Filter interface, two methods added for lookup optimization
- Fixes bug where providers were not recoginized in vendor directory
- Added Custom Injection Handler along with doc and example
- Changed to use go modules
- optional injection

## [0.4.0] - 2018-09-19
### Changed
- Add ValidationService example to README
- Add other improvements to README
- Add API Inject to ServiceLocator
- Some code refactoring
- Binder.Bind method now takes a pointer to a struct
- ConfigurationListener service
- ImmediateScope implemented

## [0.3.0] - 2018-09-12
### Added
- ServiceLocator for looking up services
- Binder for adding services to ServiceLocator
- Injection of services into other services
- Singleton and PerLookup scopes
- Context scope
- Documentation and examples
- Error service for catching certain errors in Dargo processing
- Validation service for security and other applications
