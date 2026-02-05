#!/bin/bash
set -e

# Check if govulncheck is installed
if ! command -v govulncheck &>/dev/null; then
    echo "govulncheck is not installed. Installing..."
    if ! go install golang.org/x/vuln/cmd/govulncheck@latest; then
        echo "Warning: Failed to install govulncheck, skipping vulnerability scan"
        exit 0
    fi
    echo "govulncheck installed successfully"
fi

# Verify govulncheck is now available
if ! command -v govulncheck &>/dev/null; then
    echo "Warning: govulncheck not found in PATH after installation, skipping scan"
    exit 0
fi

# Run govulncheck vulnerability scan
echo "Running govulncheck vulnerability scan..."
if ! output=$(govulncheck ./... 2>&1); then
    echo ""
    echo "govulncheck found vulnerabilities in dependencies!"
    echo "$output"
    echo ""
    echo "Please fix the vulnerabilities before committing."
    echo ""
    echo "To update vulnerable dependencies, run:"
    echo "  go get -u <package>@<fixed-version>"
    echo "  go mod tidy"
    echo ""
    echo "For more information, visit: https://go.dev/security/vuln"
    exit 1
fi

echo "No vulnerabilities found by govulncheck"
