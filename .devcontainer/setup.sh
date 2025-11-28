#! /usr/bin/bash

set -e

readonly migrate_version="v4.19.0"

echo "Downloading migrate tool..."
mkdir -p .tools
cd .tools
mkdir -p migrate
cd migrate
curl -L "https://github.com/golang-migrate/migrate/releases/download/${migrate_version}/migrate.linux-$(go env GOARCH).tar.gz" | tar xvz
