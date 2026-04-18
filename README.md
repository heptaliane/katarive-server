# Katarive Server

## Description
**Katarive Server** is the central backend engine for the Katarive audio narration service. It manages narration job lifecycles and provides a plugin-based architecture for flexible audio generation. 

The server acts as a unified endpoint, supporting native gRPC, gRPC-Web, and static asset hosting.

## Dependencies
The following external tools are required to build and manage the project (excludes libraries already managed in `go.mod`):
- **Go**: v1.25+
- **Buf CLI**: For Protobuf code generation
- **Make**: For running build tasks

## Core Features

### UI Customization
The server serves static files from the `web/` directory. You can customize the dashboard by placing your preferred static UI files (HTML, CSS, JS) into the `web/` directory. By default, these are served at the `/static/` path.

### Plugin System
Katarive supports an extensible plugin system for audio narration. You can add your preferred narrator or source plugins by placing the compiled plugin binaries into the `plugins/` directory. The server automatically loads plugins from this location at startup.

## Main Usage

### Run the Server
Starts the server on port `9421`:
```bash
make run
```

### Protocol Generation
Generate Go code from remote Protobuf definitions:
```bash
make pb-generate
```

### Running Tests
Execute the unit test suite:
```bash
make test
```

### Code Formatting
Format the Go source code:
```bash
make format
```
