#!/bin/bash
# GOAT Release Script
# This script automates the process of creating a new release

set -e

# Check if version is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 1.0.0"
  exit 1
fi

VERSION=$1
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

echo "Preparing release v$VERSION on branch $CURRENT_BRANCH..."

# Ensure working directory is clean
if [ -n "$(git status --porcelain)" ]; then
  echo "Error: Working directory is not clean. Please commit or stash changes."
  exit 1
fi

# Run tests
echo "Running tests..."
make check

# Build release binaries
echo "Building release binaries..."
VERSION=$VERSION make release

# Create release packages
echo "Creating release packages..."
VERSION=$VERSION make package-release

# Create git tag
echo "Creating git tag v$VERSION..."
git tag -a "v$VERSION" -m "Release v$VERSION"

echo ""
echo "Release v$VERSION prepared successfully!"
echo ""
echo "Next steps:"
echo "1. Push the tag: git push origin v$VERSION"
echo "2. Create a GitHub release with the tag"
echo "3. Upload the release packages from bin/packages/"
echo ""
echo "Release binaries are in bin/release/"
echo "Release packages are in bin/packages/"
