#!/bin/bash

# Multi-architecture Docker build and push script for VSQL
# This script builds Docker images for multiple architectures and pushes them to DockerHub

set -e

# Configuration
DOCKER_REPO="satetsu888/vsql"
PLATFORMS="linux/amd64,linux/arm64,linux/arm/v7"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to check if docker buildx is available
check_buildx() {
    if ! docker buildx version &> /dev/null; then
        print_error "Docker buildx is not available. Please update Docker."
        exit 1
    fi
}

# Function to create or use buildx builder
setup_builder() {
    BUILDER_NAME="vsql-builder"
    
    # Check if builder already exists
    if docker buildx ls | grep -q "$BUILDER_NAME"; then
        print_info "Using existing buildx builder: $BUILDER_NAME"
    else
        print_info "Creating new buildx builder: $BUILDER_NAME"
        docker buildx create --name "$BUILDER_NAME" --use
    fi
    
    # Start the builder
    docker buildx inspect --bootstrap
}

# Function to build and push multi-arch image
build_and_push() {
    local tag=$1
    local push_flag=$2
    
    print_info "Building multi-architecture image: ${DOCKER_REPO}:${tag}"
    print_info "Platforms: ${PLATFORMS}"
    
    local build_args="--platform ${PLATFORMS} --tag ${DOCKER_REPO}:${tag}"
    
    if [ "$push_flag" == "true" ]; then
        build_args="${build_args} --push"
        print_info "Will push to DockerHub after build"
    else
        build_args="${build_args} --load"
        print_warning "Build only mode - will not push to DockerHub"
    fi
    
    # Build the image
    docker buildx build ${build_args} .
    
    if [ $? -eq 0 ]; then
        print_info "Successfully built ${DOCKER_REPO}:${tag}"
        if [ "$push_flag" == "true" ]; then
            print_info "Successfully pushed to DockerHub"
        fi
    else
        print_error "Build failed"
        exit 1
    fi
}

# Function to check Docker Hub login
check_docker_login() {
    # Try to pull a small image to check if we're logged in
    if ! docker pull hello-world:latest &>/dev/null; then
        print_warning "Not logged in to Docker Hub"
        print_info "Please run: docker login"
        return 1
    fi
    return 0
}

# Main script
main() {
    print_info "Starting multi-architecture Docker build for VSQL"
    
    # Parse command line arguments
    PUSH_TO_HUB=false
    TAG="latest"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --push)
                PUSH_TO_HUB=true
                shift
                ;;
            --tag)
                TAG="$2"
                shift 2
                ;;
            --help|-h)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --push          Push the image to DockerHub after building"
                echo "  --tag TAG       Specify the tag to use (default: latest)"
                echo "  --help, -h      Show this help message"
                echo ""
                echo "Examples:"
                echo "  $0                    # Build only, don't push"
                echo "  $0 --push             # Build and push with 'latest' tag"
                echo "  $0 --push --tag v1.0  # Build and push with 'v1.0' tag"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
    
    # Check prerequisites
    check_buildx
    
    if [ "$PUSH_TO_HUB" == "true" ]; then
        if ! check_docker_login; then
            print_error "Docker Hub login required for pushing images"
            exit 1
        fi
    fi
    
    # Setup buildx builder
    setup_builder
    
    # Build and optionally push the image
    build_and_push "$TAG" "$PUSH_TO_HUB"
    
    # Build and push latest tag if we're pushing a version tag
    if [ "$PUSH_TO_HUB" == "true" ] && [ "$TAG" != "latest" ]; then
        print_info "Also tagging as 'latest'"
        build_and_push "latest" "true"
    fi
    
    print_info "Build process completed successfully!"
}

# Run main function
main "$@"