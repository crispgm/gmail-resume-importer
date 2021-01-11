# !/usr/bin/env bash

set -e

echo "Fetching resumes..."
go run main.go

echo "Moving to hiring folder..."
mkdir -p ~/Downloads/hiring/
mv ./attachments/backend/* ~/Downloads/hiring/

echo "Done."
