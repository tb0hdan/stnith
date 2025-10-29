# Project Notes - stnith

## Project Overview
stnith is an advanced infrastructure security tool with client-server architecture support.

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
- **Engine architecture implemented**: Central coordination system for all components
- Multiple components implemented:
  - **Disk enumeration tool**: Cross-platform support (Linux, macOS, Windows), filters virtual filesystems, returns physical disk partitions with detailed information
  - **Client-server architecture**: Basic client and server packages with test coverage
  - **Destructor modules**: Disk wiping and poweroff capabilities with platform-specific implementations
  - **Security disablers**: MAC (Mandatory Access Control) handlers for different platforms
  - **Failsafes**: Process management and protection mechanisms
  - **Savers**: Data preservation modules including rsync and scriptdir functionality
  - **Utilities**: File operations, time utilities, and permissions management with cross-platform support

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
│   ├── engine/         # Core engine package
│   │   ├── destructors/    # System destructor modules
│   │   │   ├── disks/      # Disk wiping functionality
│   │   │   └── poweroff/   # System poweroff functionality
│   │   ├── disablers/      # Security disablers
│   │   │   └── mac/        # MAC (Mandatory Access Control) handlers
│   │   ├── failsafes/      # Protection mechanisms
│   │   │   └── process/    # Process management failsafes
│   │   ├── hardware/       # Hardware-related utilities
│   │   │   └── diskenum/   # Disk enumeration package
│   │   └── savers/         # Data preservation modules
│   │       ├── rsync/      # Rsync-based data saver
│   │       └── scriptdir/  # Script directory saver
│   ├── server/         # Server implementation
│   └── utils/          # Utility functions
│       ├── permissions/    # Permission management
│       ├── file.go         # File operations
│       └── time.go         # Time utilities
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
  - `engine/` - Central engine coordinating all operations
    - `destructors/` - Modules for destructive operations (disk wiping, poweroff)
    - `disablers/` - Security bypass modules (MAC handlers)
    - `failsafes/` - Protection and recovery mechanisms
    - `hardware/` - Hardware interaction utilities
    - `savers/` - Data preservation and backup modules
  - `utils/` - Common utility functions and cross-platform helpers

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
- Full cross-platform support (Linux, macOS, Windows)
- Modular architecture for easy extension
- Engine-based design centralizes component coordination
- Platform-specific implementations use build tags
- Test coverage included for critical components
