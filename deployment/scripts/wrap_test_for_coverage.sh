#!/bin/bash
set -e

COVDATAFILES_DIR="./testdata/reports/covdatafiles"
OUTPUT_HTML="./testdata/reports/it_coverage.html"

# Setup coverage directory
rm -rf "$COVDATAFILES_DIR"
mkdir -p "$COVDATAFILES_DIR"

# Run integration tests
GOCOVERDIR="$COVDATAFILES_DIR" ./deployment/scripts/integration_test.sh

# Process coverage data
go tool covdata percent -i="$COVDATAFILES_DIR" -o "$COVDATAFILES_DIR/coverage.out"
go tool cover -html="$COVDATAFILES_DIR/coverage.out" -o "$OUTPUT_HTML"

echo "Coverage report generated at $OUTPUT_HTML"
