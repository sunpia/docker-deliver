# Docker Deliver Example

This example demonstrates how to use `docker-deliver` to package a multi-service Docker Compose project with shared base images and build contexts.

## Highlights

### Efficiency Comparison

The following table compares the total size of delivered images using different methods:

| Method                  | container0 | container1 | container2 | container-base | Total Size |
|-------------------------|------------|------------|------------|----------------|------------|
| **docker-deliver**      |  Included  |  Included  |  Included  |  Included      |  **3.1 GB**|
| Separate Image Delivery |  1.8 GB    |  3.0 GB    |  1.8 GB    |   1.8 GB       |  **8.4 GB**|
| Conda Pack Tarballs     |  1.8 GB    |  2.7 GB    |  1.8 GB    |   1.8 GB       |  **8.1 GB**|



### Key Benefits
**Consistent Runtime Environment:** By delivering the exact same Docker images, you can reliably reproduce the runtime environment on any host PC, ensuring your production and test environments are identical.
- **Efficient Storage & Transfer:** Shared layers are packaged only once, dramatically reducing total size and network usage.
- **Easy to Use:** Simple commands automate complex multi-service packaging and deployment.
- **Portability:** Single tarball contains everything needed for deployment on any Docker host.
- **Offline Delivery:** Enables deployment in environments without network access or third-party tools—only Docker is required on the destination system.

## Project Structure

```
example/
├── docker-compose.base.yaml       # Base service definition
├── docker-compose.extend.yaml     # Extended services with dependencies
├── container_base/                # Base container with shared dependencies
│   ├── Dockerfile
│   └── environment.yaml
├── container0/                    # Service 0 extending base
│   ├── Dockerfile
│   ├── environment.yaml
│   └── hello_world.py
├── container1/                    # Service 1 extending base
│   ├── Dockerfile
│   ├── environment.yaml
│   └── hello_world.py
├── container2/                    # Service 2 extending base
│   ├── Dockerfile
│   ├── environment.yaml
│   └── hello_world.py
└── benchmark/
    └── benchmark.sh               # Script to measure and compare image/tarball sizes
```

## Architecture Overview

This example uses Docker's **additional contexts** feature to create a shared base image that other services can extend from. This pattern is useful for:

- Sharing common dependencies across multiple services
- Reducing build times through layer caching
- Maintaining consistency across related containers
- Optimizing image sizes by reusing base layers

### Services

1. **container-base**: Base service with Miniconda3 and common system packages
2. **container0-2**: Application services that extend the base with specific configurations

## Usage

### Building and Packaging

Run the following command from the project root directory:

```bash
docker-deliver save \
  -f example/docker-compose.base.yaml \
  -f example/docker-compose.extend.yaml \
  --tag latest \
  -o tmp
```

### Output Structure

After running the command, the `tmp/` directory will contain:

```
tmp/
├── images.tar                      # All Docker images packaged as tar archive
└── docker-compose.generated.yaml   # Modified compose file for deployment
```

### images.tar

This file contains **all Docker image layers** from your compose project, including:

- Base images (like `continuumio/miniconda3:25.3.1-1`)
- Built images from your Dockerfiles
- All intermediate layers and dependencies

The tar file can be loaded on any Docker host using:
```bash
docker load < images.tar
```

### docker-compose.generated.yaml

This is a **deployment-ready** compose file with the following modifications:

1. **Removed build contexts**: All `build:` sections are removed since images are pre-built
2. **Added missing image tags**: Services that only had build configs now have proper image names
3. **Applied custom tags**: Images are tagged with your specified tag (`1234456`)

#### Example transformation:

**Original** (from extend file):
```yaml
services:
  container0:
    build:
      context: ./container0
      dockerfile: Dockerfile
      additional_contexts:
        container_base: "service:container-base"
```

**Generated**:
```yaml
services:
  container0:
    image: example-container0:latest
```

## Deployment

To deploy the packaged project on another system:

1. **Transfer files**:
   ```bash
   scp tmp/images.tar tmp/docker-compose.generated.yaml user@target-host:/deployment/
   ```

2. **Load images**:
   ```bash
   docker load < images.tar
   ```

3. **Deploy services**:
   ```bash
   docker compose -f docker-compose.generated.yaml up
   ```

## Key Features Demonstrated

### Multi-file Compose Configuration
Using multiple compose files allows for:
- Separation of concerns (base vs. extended services)
- Environment-specific overrides
- Modular configuration management

### Additional Build Contexts
The `additional_contexts` feature enables:
- Sharing built images between services
- Complex dependency relationships
- Efficient layer reuse

### Image Tagging Strategy
Custom tags help with:
- Version management
- Deployment tracking
- Environment identification

## Benchmark

To measure and compare the size of delivered images or Conda pack tarballs, you can use the provided benchmark script.

Run the following command from the project root:

```bash
bash example/benchmark/benchmark.sh
```

This script will build the images or Conda pack files as needed and print their sizes in your terminal, allowing you to directly compare the efficiency of different delivery methods.

## Troubleshooting

### Common Issues

1. **Missing Base Image**:
   Ensure `container-base` is built before dependent services:
   ```bash
   docker-compose -f example/docker-compose.base.yaml build
   ```

2. **Build Context Errors**:
   Run from the correct working directory (project root):
   ```bash
   cd /path/to/docker-deliver
   docker-deliver save ...
   ```

This example showcases the power of `docker-deliver` for creating portable, deployment-ready Docker Compose packages with complex build dependencies.
