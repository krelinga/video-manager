#! /usr/bin/sh

set -e

apt-get update
apt-get install -y \
    ffmpeg \
    handbrake-cli
apt-get clean
rm -rf /var/lib/apt/lists/*
