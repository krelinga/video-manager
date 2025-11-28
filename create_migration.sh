#! /usr/bin/bash

# Use this tool to create migration files, see https://github.com/golang-migrate/migrate/blob/master/GETTING_STARTED.md .

set -e

.tools/migrate/migrate create -ext sql -dir migrations -seq "$@"
