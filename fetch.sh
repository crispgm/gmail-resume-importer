# !/usr/bin/env bash

set -e

echo "Fetching resumes..."
./resume-import

echo "Moving to hiring folder..."
FILESCOUNT=$(ls -l ./attachments/backend/ | wc -l)
if [ "$FILESCOUNT" -gt "0" ]; then
    mkdir -p ~/Downloads/hiring/
    mv ./attachments/backend/* ~/Downloads/hiring/
fi

echo "Done."
