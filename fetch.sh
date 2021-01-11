# !/usr/bin/env bash

set -e

echo "Fetching resumes..."
./resume-import main.go

echo "Moving to hiring folder..."
mkdir -p ~/Downloads/hiring/
mv ./attachments/backend/* ~/Downloads/hiring/

echo "Done."
