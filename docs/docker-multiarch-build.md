# Multi-Architecture Docker Build Guide

This guide explains how to build and push multi-architecture Docker images for VSQL.

## Prerequisites

1. Docker Desktop or Docker Engine with buildx support (Docker 19.03+)
2. Docker Hub account (for pushing images)
3. Login to Docker Hub: `docker login`

## Quick Start

### Build without pushing (local testing)
```bash
./build-and-push.sh
```

### Build and push to Docker Hub
```bash
./build-and-push.sh --push
```

### Build and push with custom tag
```bash
./build-and-push.sh --push --tag v1.0.0
```

## Supported Architectures

The script builds images for the following platforms:
- `linux/amd64` (Intel/AMD 64-bit)
- `linux/arm64` (ARM 64-bit, including Apple Silicon)
- `linux/arm/v7` (ARM 32-bit v7)

## How It Works

1. **Buildx Setup**: The script creates a dedicated buildx builder instance named `vsql-builder`
2. **Multi-arch Build**: Uses Docker buildx to build for all target platforms simultaneously
3. **Push to Registry**: When `--push` is specified, the manifest and all platform-specific images are pushed to Docker Hub

## Manual Build Commands

If you prefer to run the commands manually:

```bash
# Create buildx builder
docker buildx create --name vsql-builder --use

# Build and push multi-arch image
docker buildx build \
  --platform linux/amd64,linux/arm64,linux/arm/v7 \
  --tag satetsu888/vsql:latest \
  --push \
  .

# Build for local testing (single architecture)
docker buildx build \
  --platform linux/amd64 \
  --tag satetsu888/vsql:latest \
  --load \
  .
```

## Verifying Multi-Architecture Support

After pushing, you can verify the multi-architecture manifest:

```bash
docker manifest inspect satetsu888/vsql:latest
```

This will show all the available architectures for the image.

## Troubleshooting

### "docker buildx not found"
Update Docker Desktop to the latest version or install buildx manually.

### "failed to solve: error getting credentials"
Run `docker login` to authenticate with Docker Hub.

### Build fails on specific architecture
Check the Dockerfile for architecture-specific issues. The Go compiler and pg_query_go should support all listed architectures.

## CI/CD Integration

For automated builds in CI/CD pipelines, you can use the script with appropriate flags:

```bash
# In GitHub Actions or similar
./build-and-push.sh --push --tag ${GITHUB_REF#refs/tags/}
```