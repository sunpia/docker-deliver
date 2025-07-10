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