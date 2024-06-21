#!/bin/sh
# scripts/completions.sh
set -e

# exit if no arguments are provided
if [ "$#" -ne 1 ]; then
    echo "Generates shell completions for icpctl"
    echo "Usage: $0 <target_dir>"
    exit 1
fi

# $TARGET_DIR is set from first argument
TARGET_DIR=$1/completions

rm -rf $TARGET_DIR
mkdir -p $TARGET_DIR
for sh in bash zsh fish; do
	go run main.go completion "$sh" >"$TARGET_DIR/icpctl.$sh"
	# set static accessed and modified date to files
	touch -a -m -t 202401010000.00 "$TARGET_DIR/icpctl.$sh"
done
