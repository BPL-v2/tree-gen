#!/bin/bash

# Script to download passive tree data from multiple repositories
# Calls download_passives.sh for each repository with appropriate directories

set -e  # Exit on any error

echo "=========================================="
echo "Downloading passive tree data from repos"
echo "=========================================="
echo

# Check if download_passives.sh exists
if [ ! -f "./download_passives.sh" ]; then
    echo "ERROR: download_passives.sh not found in current directory"
    echo "Make sure you're running this script from the same directory as download_passives.sh"
    exit 1
fi

# Make sure download_passives.sh is executable
chmod +x ./download_passives.sh

echo "🌳 Downloading from atlastree-export repository..."
echo "Repository: git@github.com:grindinggear/atlastree-export.git"
echo "Save directory: ./atlastree/"
echo
./download_passives.sh git@github.com:grindinggear/atlastree-export.git grindinggear/atlastree-export ./atlastree

echo
echo "=========================================="
echo

echo "🌲 Downloading from skilltree-export repository..."
echo "Repository: git@github.com:grindinggear/skilltree-export.git"
echo "Save directory: ./skilltree/"
echo
./download_passives.sh git@github.com:grindinggear/skilltree-export.git grindinggear/skilltree-export ./skilltree

echo
echo "=========================================="
echo "✅ All downloads complete!"
echo "=========================================="
echo
echo "Summary of directories:"
echo
echo "📁 atlastree/ - Contains data from atlastree-export repository"
if [ -d "./atlastree" ]; then
    atlas_count=$(find ./atlastree -name "*.json" 2>/dev/null | wc -l)
    echo "   Files: $atlas_count JSON files"
else
    echo "   Directory not found or empty"
fi

# echo
# echo "📁 skilltree/ - Contains data from skilltree-export repository"
# if [ -d "./skilltree" ]; then
#     skill_count=$(find ./skilltree -name "*.json" 2>/dev/null | wc -l)
#     echo "   Files: $skill_count JSON files"
# else
#     echo "   Directory not found or empty"
# fi

# echo
echo "Done! 🎉"