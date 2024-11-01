#!/bin/bash

# Usage: ./rename-project.sh <old_name> <new_name>

# Check for correct number of arguments
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <old_name> <new_name>"
    exit 1
fi

OLD_NAME="$1"
NEW_NAME="$2"

OLD_BASE_NAME=$(basename $OLD_NAME)
NEW_BASE_NAME=$(basename $NEW_NAME)

# Array to hold matching file paths
declare -a file_paths

# Find files with the specified extensions and exclude specified paths. Store them in the array.
while IFS= read -r -d $'\0' file; do
    file_paths+=("$file")
done < <(find . -type d \( -path ./bin -o -path ./vendor -o -path ./dist \) -prune -o -type f \( -name "*.go" -o -name "*.mod" -o -name "*.md" -o -name "*.yaml" -o -name "*.sh" \) -print0)

# Iterate over the files to rename occurrences
for file in "${file_paths[@]}"; do
    echo "Renaming occurrences in $file"
    # Uncomment the following line to perform the actual replacement when you're ready
    sed -i '' "s|$OLD_NAME|$NEW_NAME|g" "$file"
    sed -i '' "s|$OLD_BASE_NAME|$NEW_BASE_NAME|g" "$file"
done

# Rename BINARY_NAME=indev to BINARY_NAME=NEW_BIN_NAME where NEW_BIN_NAME is the last segment from NEW_NAME delimited by "/"
sed -i '' "s|$OLD_BASE_NAME|$NEW_BASE_NAME|g" Makefile
sed -i '' "s|$OLD_BASE_NAME|$NEW_BASE_NAME|g" .gitlab-ci.yml

echo "Rename completed."