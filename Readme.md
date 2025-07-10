# Docker Deliver

A command-line tool for packaging and delivering Docker Compose projects. Docker Deliver allows you to build, save, and package Docker Compose projects into portable archives for easy deployment.

## Features

- 🚀 Build Docker Compose projects
- 📦 Save All Docker images in Compose projects to tar archives.
- 📄 Generate portable compose files
- 🏷️ Custom image tagging support
- 📝 Configurable logging levels
- 🔧 Flexible output directory management

## Table of Contents

- [Installation](#installation)
- [Building from Source](#building-from-source)
- [Usage](#usage)
- [Examples](#examples)
- [Configuration](#configuration)
- [Requirements](#requirements)
- [Contributing](#contributing)
- [License](#license)

## Installation

### Prerequisites

Before installing Docker Deliver, ensure you have the following installed:

- **Docker Engine** (version 20.10 or later)
- **Docker Compose** (version 2.0 or later)
- **Go** (version 1.21 or later) - only needed for building from source

### Install with Go

You can install the latest version directly using Go:

```bash
go install github.com/sunpia/docker-deliver/cmd/docker-deliver@latest
```

Make sure that your `GOPATH/bin` is in your `PATH` to run `docker-deliver` from anywhere.

## Building from Source

### Clone the Repository

```bash
git clone https://github.com/sunpia/docker-deliver.git
cd docker-deliver
```

### Build with Make

```bash
# Build the binary
make build

# The binary will be created at ./dist/docker-deliver

# Or install to GOPATH/bin
make install
```

### Build with Go

```bash
# Build directly with Go
go build -o docker-deliver ./cmd/docker-deliver

# Or install to GOPATH/bin
go install ./cmd/docker-deliver
```

### Development Dependencies

```bash
# Download Go dependencies
go mod download

# Verify dependencies
go mod verify
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

### Help

```bash
# Show help
docker-deliver --help

# Show help for save command
docker-deliver save --help
```

## Examples && Benchmark
For more detailed usage examples and benchmarks, please visit [example/Readme.md](example/Readme.md) in this repository.


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

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/yourusername/docker-deliver.git
cd docker-deliver

# Install dependencies
go mod download

# Run tests
go test ./...

# Build and test locally
make build
./dist/docker-deliver save --help
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra CLI](https://github.com/spf13/cobra)
- Uses [Docker Compose Go](https://github.com/compose-spec/compose-go) library
- Inspired by the need for portable Docker deployments

---

For more information, visit the [project repository](https://github.com/sunpia/docker-deliver) or open an issue for support.