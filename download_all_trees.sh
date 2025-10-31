#!/bin/bash

# Script to download passive tree data from multiple repositories
# Calls download_passives.sh for each repository with appropriate directories

set -e  # Exit on any error

echo "=========================================="
echo "Downloading passive tree data from repos"
echo "=========================================="
echo

# Check if jq is available for JSON minification
if ! command -v jq >/dev/null 2>&1; then
    echo "Error: jq is required for JSON minification but not installed."
    echo "Please install jq: sudo apt install jq"
    exit 1
fi

# Check if download_passives.sh exists
if [ ! -f "./download_passives.sh" ]; then
    echo "ERROR: download_passives.sh not found in current directory"
    echo "Make sure you're running this script from the same directory as download_passives.sh"
    exit 1
fi

# Make sure download_passives.sh is executable
chmod +x ./download_passives.sh
./download_passives.sh git@github.com:grindinggear/atlastree-export.git grindinggear/atlastree-export ./atlastree
./download_passives.sh git@github.com:grindinggear/skilltree-export.git grindinggear/skilltree-export ./skilltree


echo "Done! 🎉"