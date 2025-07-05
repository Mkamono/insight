# CLAUDE.md

日本語で会話して。

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is "insight" - an AI-powered knowledge documentation and linking system. The project aims to automatically create, update, and tag documents from fragments of information using AI.

### Architecture Vision

The project is designed with a multi-interface architecture:
- **core**: TypeScript core functionality library
- **cli**: Command-line interface
- **web**: Next.js web server with frontend
- **ai**: AI-powered document creation and management (may be part of core)

### Database Schema

The system uses SQLite with the following core entities:
- **Fragment**: Content pieces with optional URL/image, processing status, and hierarchical relationships
- **Document**: Titled documents with content, summary, linked to fragments and tags
- **Tag**: Named labels for categorizing documents
- **Question**: Stored questions for the system

Key relationships:
- Fragments can have parent-child relationships
- Documents and Fragments have many-to-many relationships
- Documents and Tags have many-to-many relationships

## Development Commands

### Build
```bash
mise run build
```
This compiles the Go application to `./insight` binary.

### Environment Setup
The project uses [mise](https://mise.jdx.dev/) for tool management. Required tools are automatically installed:
- Node.js (latest)
- Google Gemini CLI
- SQLite
- Task runner

### Environment Variables
- `GEMINI_API_KEY`: Set in mise.toml for AI functionality

## Current State

The repository is in early development with most Go source files deleted but the build configuration remains. The project appears to be transitioning from Go to TypeScript implementation as described in AI_INSTRUCTION.md.

Key files:
- `mise.toml`: Tool and environment configuration
- `AI_INSTRUCTION.md`: Project specifications in Japanese
- Go source files are deleted but SQL schema files were present

## Important Notes

- The database schema is not fixed and may evolve
- Focus on CRUD operations first before AI functionality
- The project supports multiple interfaces (CLI, web, core library)
- AI functionality uses Google Gemini API
