日本語で会話して

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a minimal Go project called "insight" with a basic structure for CLI and migration tools. The project is in its early stages with placeholder implementations.

## Architecture

The project follows a standard Go project layout:

- `cmd/` - Contains application entry points
  - `cli/` - CLI tool entry point (main.go)
  - `migrate/` - Database migration tool entry point (main.go)
- `src/` - Source code directory
  - `models/` - Data models and structures (currently empty)

## Development Commands

### Building the project
```bash
go build ./cmd/cli         # Build CLI tool
go build ./cmd/migrate     # Build migration tool
```

### Running applications
```bash
go run ./cmd/cli           # Run CLI tool
go run ./cmd/migrate       # Run migration tool
```

### Go module management
```bash
go mod tidy               # Clean up dependencies
go mod download           # Download dependencies
```

### Code formatting
```bash
go fmt ./...              # Format all Go files
```

### Testing
```bash
go test ./...             # Run all tests
go test -v ./...          # Run tests with verbose output
```

## Current State

The project is in a very early stage with:
- Basic Go module setup (Go 1.24.5)
- Placeholder main functions that only print status messages
- Empty models package
- No external dependencies
- No database integration yet
- No actual business logic implemented

Both CLI and migration tools currently only output placeholder messages and need implementation.
