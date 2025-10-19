#!/bin/bash

# Build script for Easy Sync
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# App information
APP_NAME="easy-sync"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo -e "${GREEN}Building ${APP_NAME} version ${VERSION}${NC}"

# Create build directory
mkdir -p build

# Build flags
LDFLAGS="-ldflags \"-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}\""

# Function to build for a specific platform
build_platform() {
    local os=$1
    local arch=$2
    local ext=$3

    echo -e "${YELLOW}Building for ${os}/${arch}...${NC}"

    GOOS=$os GOARCH=$arch go build $LDFLAGS -o build/${APP_NAME}-${os}-${arch}${ext} ./cmd/server

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Build successful for ${os}/${arch}${NC}"
    else
        echo -e "${RED}✗ Build failed for ${os}/${arch}${NC}"
        exit 1
    fi
}

# Build for current platform first
echo -e "${YELLOW}Building for current platform...${NC}"
go build $LDFLAGS -o build/${APP_NAME} ./cmd/server

# Build for multiple platforms if requested
if [ "$1" = "all" ]; then
    echo -e "${YELLOW}Building for all platforms...${NC}"

    # Linux
    build_platform "linux" "amd64" ""
    build_platform "linux" "arm64" ""

    # Windows
    build_platform "windows" "amd64" ".exe"

    # macOS
    build_platform "darwin" "amd64" ""
    build_platform "darwin" "arm64" ""

    echo -e "${GREEN}All builds completed successfully!${NC}"
    echo -e "${YELLOW}Binaries are available in the 'build' directory${NC}"
else
    echo -e "${GREEN}Build completed successfully!${NC}"
    echo -e "${YELLOW}Binary: build/${APP_NAME}${NC}"
fi

# Show build information
echo -e "${YELLOW}Build Information:${NC}"
echo -e "  Version: ${VERSION}"
echo -e "  Build Time: ${BUILD_TIME}"
echo -e "  Go Version: $(go version)"

# Make binary executable
chmod +x build/${APP_NAME} 2>/dev/null || true

echo -e "${GREEN}Done!${NC}"