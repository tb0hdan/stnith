# Project Notes - stnith

## Project Overview
stnith is a Go project currently in initial development phase.

## Goals
- Provide cross-platform system utilities
- Create tools for disk and system information gathering

## Current Implementation Status
- Initial project structure set up
- Basic build configuration with Makefile
- Linting configuration with .golangci.yml
- Git repository initialized
- Disk enumeration tool implemented:
  - Cross-platform support (Linux and macOS)
  - Filters out virtual filesystems (tmpfs, loop devices, etc.)
  - Returns only physical disk partitions
  - Provides detailed partition information (size, usage, mount points, labels)

## Architecture
- cmd/ - Command-line applications
- pkg/ - Reusable packages

## Development Setup
- Go module initialized (go.mod)
- Makefile configured with build, test, and lint commands
- Golangci-lint configured for code quality

## Next Steps
- Define project goals and requirements
- Implement core functionality
- Add tests
- Document API and usage

## Notes
- Project uses Make for build automation
- Follows standard Go project layout