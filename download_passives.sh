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

# Create the save directory if it doesn't exist
mkdir -p "$SAVE_DIR"

# Get all tags and sort them from remote repository
echo "Getting list of tags from remote repository..."
tags=$(git ls-remote --tags --refs "$REPO_URL" | sed 's/.*refs\/tags\///' | sort -V)

# Counter for progress
total_tags=$(echo "$tags" | wc -l)
current=0
skipped=0
downloaded=0
not_found=0

echo "Found $total_tags tags. Checking for existing files and downloading missing ones..."
echo

# Loop through each tag
for tag in $tags; do
    current=$((current + 1))
    echo "[$current/$total_tags] Processing tag: $tag"
    
    # Check if we already have a file for this tag with any extension
    existing_file=""
    if [ -f "$SAVE_DIR/$tag.json" ]; then
        existing_file="$SAVE_DIR/$tag.json"
    elif [ -f "$SAVE_DIR/$tag.txt" ]; then
        existing_file="$SAVE_DIR/$tag.txt"
    elif [ -f "$SAVE_DIR/$tag.data" ]; then
        existing_file="$SAVE_DIR/$tag.data"
    fi
    
    if [ -n "$existing_file" ]; then
        echo "  ⏭️  File already exists: $existing_file"
        skipped=$((skipped + 1))
        echo
        continue
    fi
    
    # Get the commit hash for this tag from remote
    commit_hash=$(git ls-remote --tags "$REPO_URL" "refs/tags/$tag" | cut -f1)
    
    if [ -z "$commit_hash" ]; then
        echo "  ERROR: Could not get commit hash for tag $tag"
        continue
    fi
    

    
    # Construct the URL for the source archive
    zip_url="https://github.com/$GITHUB_REPO/archive/refs/tags/$tag.zip"
    temp_zip="/tmp/${tag}_source.zip"
    temp_dir="/tmp/${tag}_extract"
    

    
    # Download the source zip (follow redirects)
    if curl -s -f -L -o "$temp_zip" "$zip_url"; then
        echo "  ✓ Downloaded source archive"
        
        # Validate that we got a proper zip file
        if ! file "$temp_zip" | grep -q "Zip archive"; then
            echo "  ⚠️  Downloaded file is not a valid zip archive (tag may not exist)"
            not_found=$((not_found + 1))
            rm -f "$temp_zip"
            echo
            continue
        fi
        
        # Create temporary extraction directory
        rm -rf "$temp_dir"
        mkdir -p "$temp_dir"
        
        # Extract the zip file
        if unzip -q "$temp_zip" -d "$temp_dir" 2>/dev/null; then
            echo "  ✓ Extracted source archive"
            
            # Find the data.json file in the extracted directory
            
            # Find the actual extracted folder (there should be exactly one directory)
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
                    # Determine the correct extension based on the source file
                    source_filename=$(basename "$data_file")
                    if [[ "$source_filename" == "data.json" ]]; then
                        output_file="$SAVE_DIR/$tag.json"
                    elif [[ "$source_filename" == "data.txt" ]]; then
                        output_file="$SAVE_DIR/$tag.txt"
                    else
                        output_file="$SAVE_DIR/$tag.data"
                    fi
                    
                    # Check if file already exists
                    if [ -f "$output_file" ]; then
                        echo "  ⏭️  File already exists: $output_file"
                        skipped=$((skipped + 1))
                    else
                        # Copy the data file to our target location with correct extension
                        cp "$data_file" "$output_file"
                        echo "  ✓ Successfully extracted $(basename "$output_file") (from $source_filename)"
                        downloaded=$((downloaded + 1))
                    fi
                else
                    echo "  ⚠️  No data file found (checked data.json, data.txt, data) in source archive for tag $tag"
                    not_found=$((not_found + 1))
                fi
            else
                echo "  ⚠️  Extracted folder not found: $extracted_folder"
                not_found=$((not_found + 1))
            fi
        else
            echo "  ✗ Failed to extract source archive for tag $tag (corrupted zip or tag doesn't exist)"
            not_found=$((not_found + 1))
        fi
        
        # Clean up temporary files
        rm -f "$temp_zip"
        rm -rf "$temp_dir"
    else
        echo "  ✗ Failed to download source archive for tag $tag"
        not_found=$((not_found + 1))
    fi
    
    echo
done

echo "Download complete!"
echo
echo "Summary:"
echo "  Total tags processed: $total_tags"
echo "  Files downloaded: $downloaded"
echo "  Files skipped (already existed): $skipped"
echo "  Files not found (404): $not_found"
echo "  Save directory: $SAVE_DIR"
echo
echo "Files in save directory:"
ls -la "$SAVE_DIR"/ | grep -E "\.json$" | wc -l | xargs echo "  Total JSON files:"