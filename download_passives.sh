#!/bin/bash

# Script to download data.json for each git tag from a remote repository
# Creates files in specified directory for each tag
# Usage: ./download_passives.sh [repository_url] [owner/repo] [save_directory]
#
# Examples:
#   ./download_passives.sh git@github.com:grindinggear/skilltree-export.git grindinggear/skilltree-export ./passives
#   ./download_passives.sh https://github.com/grindinggear/skilltree-export.git grindinggear/skilltree-export /tmp/data
#   ./download_passives.sh "" "" ./my-passives  # Use defaults for repo but custom directory

# Default values (can be overridden by command line arguments)
REPO_URL="${1:-git@github.com:grindinggear/skilltree-export.git}"
GITHUB_REPO="${2:-grindinggear/skilltree-export}"
SAVE_DIR="${3:-./passives}"

echo "Repository URL: $REPO_URL"
echo "GitHub repo: $GITHUB_REPO"
echo "Save directory: $SAVE_DIR"
echo

# Check if jq is available for JSON minification
if ! command -v jq >/dev/null 2>&1; then
    echo "Error: jq is required for JSON minification but not installed."
    echo "Please install jq: sudo apt install jq"
    exit 1
fi

# Create the save directory if it doesn't exist
mkdir -p "$SAVE_DIR"

# Get all tags and sort them from remote repository
tags=$(git ls-remote --tags --refs "$REPO_URL" | sed 's/.*refs\/tags\///' | sort -V)

# Counter for progress
total_tags=$(echo "$tags" | wc -l)
current=0
skipped=0
downloaded=0
not_found=0


# Loop through each tag
for tag in $tags; do
    current=$((current + 1))    
    # Remove "-atlas" from the tag string, then remove ".0" from the end
    clean_tag="${tag/-atlas/}"
    
    # Check if the clean tag ends with "0"
    if [[ ! "$clean_tag" =~ 0$ ]]; then
        continue
    fi
        clean_tag="${clean_tag%.0}"
    
    # Check if we already have a file for this clean tag with any extension
    existing_file=""
    if [ -f "$SAVE_DIR/$clean_tag.json" ]; then
        existing_file="$SAVE_DIR/$clean_tag.json"
    elif [ -f "$SAVE_DIR/$clean_tag.txt" ]; then
        existing_file="$SAVE_DIR/$clean_tag.txt"
    elif [ -f "$SAVE_DIR/$clean_tag.data" ]; then
        existing_file="$SAVE_DIR/$clean_tag.data"
    fi
    
    if [ -n "$existing_file" ]; then
        continue
    fi
    
    # Get the commit hash for this tag from remote
    commit_hash=$(git ls-remote --tags "$REPO_URL" "refs/tags/$tag" | cut -f1)
    
    if [ -z "$commit_hash" ]; then
        echo "  ERROR: Could not get commit hash for tag $tag"
        continue
    fi
    

    
    # Construct the URL for the source archive (using original tag)
    zip_url="https://github.com/$GITHUB_REPO/archive/refs/tags/$tag.zip"
    temp_zip="/tmp/${clean_tag}_source.zip"
    temp_dir="/tmp/${clean_tag}_extract"
    

    
    # Download the source zip (follow redirects)
    echo "[$current/$total_tags] Downloading tag: $tag -> $clean_tag"
    if curl -s -f -L -o "$temp_zip" "$zip_url"; then
        if ! file "$temp_zip" | grep -q "Zip archive"; then
            rm -f "$temp_zip"
            continue
        fi
        
        # Create temporary extraction directory
        rm -rf "$temp_dir"
        mkdir -p "$temp_dir"
        
        # Extract the zip file
        if unzip -q "$temp_zip" -d "$temp_dir" 2>/dev/null; then
            extracted_folder=$(find "$temp_dir" -maxdepth 1 -type d ! -path "$temp_dir" | head -1)
            
            if [ -d "$extracted_folder" ]; then
                # Look for data file in various formats
                data_file=""
                if [ -f "$extracted_folder/data.json" ]; then
                    data_file="$extracted_folder/data.json"
                elif [ -f "$extracted_folder/data.txt" ]; then
                    data_file="$extracted_folder/data.txt"
                elif [ -f "$extracted_folder/data" ]; then
                    data_file="$extracted_folder/data"
                fi
                
                if [ -n "$data_file" ]; then
                    source_filename=$(basename "$data_file")
                    if [[ "$source_filename" == "data.json" ]]; then
                        output_file="$SAVE_DIR/$clean_tag.json"
                        jq -c . "$data_file" > "$output_file"
                    elif [[ "$source_filename" == "data.txt" ]]; then
                        output_file="$SAVE_DIR/$clean_tag.txt"
                        cp "$data_file" "$output_file"
                    else
                        output_file="$SAVE_DIR/$clean_tag.data"
                        cp "$data_file" "$output_file"
                    fi
                    downloaded=$((downloaded + 1))
                fi
            fi
        fi
        
        # Clean up temporary files
        rm -f "$temp_zip"
        rm -rf "$temp_dir"
    fi
done

echo "Download complete!"
