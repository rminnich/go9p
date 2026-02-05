#!/bin/bash
set -ex

copyright_header='// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.'

echo "Starting copyright check..."

update_copyright() {
    local file="$1"
    local temp_file
    temp_file=$(mktemp)
    echo "${copyright_header}" >"${temp_file}"
    echo "" >>"${temp_file}" # Add an empty line after the header
    # Remove existing copyright header (// Copyright ... license file.) and leading blank lines
    sed '/^\/\/ Copyright.*Go9p/,/^\/\/ .*LICENSE/d' "$file" | sed '/./,$!d' >>"${temp_file}"
    mv "${temp_file}" "${file}"
}

# Get the list of staged .go files
staged_files=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

# Check if there are any staged .go files
if [[ -z "$staged_files" ]]; then
    echo "No .go files staged for commit. Exiting."
    exit 0
fi

for file in $staged_files; do
    echo "Checking file: $file"
    if grep -qF "Copyright" "$file" && grep -qF "Go9p Authors" "$file"; then
        echo "Copyright header present in $file"
    else
        echo "Updating copyright header in $file"
        update_copyright "$file"
        echo "Copyright header updated in $file"
    fi
done

echo "Copyright check completed."
