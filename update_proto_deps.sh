#! /usr/bin/bash

set -e

# Check if there are any uncommitted changes
if ! git diff-index --quiet HEAD --; then
	echo "Error: There are uncommitted changes in the repository. Please commit or stash them before running this script."
	exit 1
fi

go get "buf.build/gen/go/krelinga/proto/connectrpc/go@latest"
go get "buf.build/gen/go/krelinga/proto/protocolbuffers/go@latest"

# Check if there are any changes after updating dependencies
if ! git diff-index --quiet HEAD --; then
    git add .
    git commit -m "Update protobuf dependencies."
fi