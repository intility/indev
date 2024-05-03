#!/bin/sh
# scripts/completions.sh
set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish; do
	go run main.go completion "$sh" >"completions/icpctl.$sh"
	# set static accessed and modified date to files
	touch -a -m -t 202401010000.00 "completions/icpctl.$sh"
done