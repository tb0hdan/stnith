# Project Notes - stnith

## Project Overview
stnith is a Go project providing system utilities and destructive operations tools with client-server architecture support.

## Goals
- Provide cross-platform system utilities
- Create tools for disk and system information gathering
- Implement client-server architecture for remote operations
- Support various system destructors (disk wiping, poweroff)
- Handle security disablers (AppArmor, SELinux)

## Current Implementation Status
- Initial project structure set up with standard Go layout
- Basic build configuration with Makefile
- Linting configuration with .golangci.yml
- Git repository initialized
- Multiple components implemented:
  - **Disk enumeration tool**: Cross-platform support (Linux and macOS), filters virtual filesystems, returns physical disk partitions with detailed information
  - **Client-server architecture**: Basic client and server packages
  - **Destructor modules**: Disk wiping and poweroff capabilities
  - **Security disablers**: AppArmor and SELinux handlers
  - **Utilities**: File and time helper functions

## Directory Structure
```
.
├── build/              # Build output directory
├── cmd/                # Command-line applications
│   ├── diskenum/       # Disk enumeration CLI tool
│   └── stnith/         # Main stnith application
├── docs/               # Documentation
├── pkg/                # Reusable packages
│   ├── client/         # Client implementation
│   ├── destructors/    # System destructor modules
│   │   ├── disks/      # Disk wiping functionality
│   │   └── poweroff/   # System poweroff functionality
│   ├── disablers/      # Security disablers
│   │   ├── apparmor/   # AppArmor disabler
│   │   └── selinux/    # SELinux disabler
│   ├── hardware/       # Hardware-related utilities
│   │   └── diskenum/   # Disk enumeration package
│   ├── server/         # Server implementation
│   └── utils/          # Utility functions
│       ├── file.go     # File operations
│       └── time.go     # Time utilities
```

## Key Files
- `Makefile` - Build, test, and lint commands
- `.golangci.yml` - Linting configuration
- `go.mod` - Go module configuration
- `CLAUDE.md` - Claude AI instructions for the project
- `README.md` - Project readme

## Architecture
- **cmd/** - Contains entry points for CLI applications
  - `diskenum/` - Standalone disk enumeration tool
  - `stnith/` - Main application with full functionality
- **pkg/** - Core library packages
  - `client/` - Client-side communication logic
  - `server/` - Server-side request handling
  - `destructors/` - Modules for destructive operations
  - `disablers/` - Security bypass modules
  - `hardware/` - Hardware interaction utilities
  - `utils/` - Common utility functions

## Development Setup
- Go module initialized (go.mod)
- Makefile configured with build, test, and lint commands
- Golangci-lint configured for code quality
- Build outputs go to `build/` directory

## Build Commands
- `make build` - Build all binaries
- `make test` - Run tests
- `make lint` - Run linting checks
- `make clean` - Clean build artifacts

## Next Steps
- Add comprehensive tests for all packages
- Implement authentication for client-server communication
- Add configuration file support
- Improve error handling and logging
- Document API and usage examples
- Add more destructor modules
- Implement progress reporting for long operations

## Notes
- Project uses Make for build automation
- Follows standard Go project layout
- Cross-platform support (Linux and macOS)
- Modular architecture for easy extension