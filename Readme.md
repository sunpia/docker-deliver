# Docker Deliver

A command-line tool for packaging and delivering Docker Compose projects. Docker Deliver enables you to build, save, and package Docker Compose projects into portable archives for easy deployment—including fully offline delivery to environments without internet access.

## Features

- 🛠️ Seamless integration with Docker Compose projects
- 📦 Packages all Docker images into a single tar archive for easy transfer
- 🗂️ Shares Docker layers between images to minimize disk usage
- 🌐 Supports fully offline deployment—no network required on the destination host
- 🚫 No requirement for third-party software on the destination host—only Docker and Docker Compose needed

## Table of Contents

- [Installation](#installation)
- [Building from Source](#building-from-source)
- [Usage](#usage)
- [MCP (Model Context Protocol) Server](#mcp-model-context-protocol-server)
- [Examples](#examples--benchmark)
- [Configuration](#configuration)
- [License](#license)

## Installation

### Prerequisites

Before installing Docker Deliver, ensure you have the following installed:

- **Docker Engine** (version 20.10 or later)
- **Docker Compose** (version 2.0 or later)
- **Go** (version 1.21 or later) — only needed for building from source

### Install with Go

Install the latest version directly using Go:

```bash
go install github.com/sunpia/docker-deliver/cmd/docker-deliver@latest
```

Ensure your `GOPATH/bin` is in your `PATH` to run `docker-deliver` from anywhere.

## Building from Source

### Clone the Repository

```bash
git clone https://github.com/sunpia/docker-deliver.git
cd docker-deliver
```

### Build

```bash
# Build the binary to ./dist/docker-deliver
make build

# Or install to GOPATH/bin
make install
```

## Usage

### Basic Command Structure

```bash
docker-deliver save [flags]
```

### Required Flags

- `-f, --file`: Path to docker-compose file(s) (required)
- `-o, --output`: Output directory for generated files (required)

### Optional Flags

- `-w, --workdir`: Working directory (default: current directory)
- `-t, --tag`: Default tag for images (default: "latest")
- `-l, --loglevel`: Log level - debug, info, warn, error (default: "info")

## MCP (Model Context Protocol) Server

Docker Deliver includes a built-in MCP server that exposes its functionality as tools for AI assistants and other MCP-compatible clients.

### Starting the MCP Server

```bash
# Start MCP server with stdio transport (for AI assistants)
docker-deliver mcp

# Start MCP server with HTTP transport
docker-deliver mcp --http :8080
```

### Available MCP Tools

The MCP server exposes the following tools:

#### `deliver_compose_project`

Delivers a Docker Compose project and its images to a folder, enabling offline deployment.

**Parameters:**
- `docker_compose_path` (array): Paths to docker-compose file(s)
- `work_dir` (string): Working directory
- `output_dir` (string): Output directory for generated files
- `tag` (string): Default tag for images
- `loglevel` (string): Log level (debug, info, warn, error)

**Example usage in MCP client:**
```json
{
  "tool": "deliver_compose_project",
  "arguments": {
    "docker_compose_path": ["./docker-compose.yml"],
    "work_dir": "/path/to/project",
    "output_dir": "/path/to/output",
    "tag": "latest",
    "loglevel": "info"
  }
}
```

### Integration with AI Assistants

The MCP server allows AI assistants to directly use Docker Deliver's functionality. Configure your AI assistant to connect to the MCP server and use the `deliver_compose_project` tool to package and prepare Docker Compose projects for deployment.

### Server Transports

- **Stdio Transport**: Default mode for AI assistant integration
- **HTTP Transport**: Use `--http` flag to specify HTTP address for web-based integrations

## Examples & Benchmark

For more detailed usage examples and benchmarks, see [example/Readme.md](example/Readme.md) in this repository.

## Configuration

### Output Structure

After running docker-deliver, your output directory will contain:

```
output/
├── images.tar                      # Saved Docker images
└── docker-compose.generated.yaml   # Generated compose file
```

### Compose File Requirements

Your docker-compose.yml should specify either:

1. **Image names** for services using pre-built images:
    ```yaml
    services:
      web:
         image: nginx:latest
    ```

2. **Build configurations** for services built from Dockerfiles:
    ```yaml
    services:
      app:
         build:
            context: ./app
            dockerfile: Dockerfile
    ```

### Development Setup

```bash
# Clone your fork
git clone https://github.com/yourusername/docker-deliver.git
cd docker-deliver

# Install dependencies
go mod download

# Build and test locally
make build
./dist/docker-deliver save --help
```

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.