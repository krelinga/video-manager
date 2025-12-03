#! /usr/bin/bash

set -e

# Check if there are any uncommitted changes
if ! git diff-index --quiet HEAD --; then
	echo "Error: There are uncommitted changes in the repository. Please commit or stash them before running this script."
	exit 1
fi

go get "github.com/krelinga/video-manager-api/go/vmapi@latest"

# Check if there are any changes after updating dependencies
if ! git diff-index --quiet HEAD --; then
    go mod tidy
    echo "Protobuf dependencies have been updated. Committing changes..."
    git add .
    git commit -m "Update api dependencies."
else
    echo "No changes in api dependencies."
    exit 1
fi
