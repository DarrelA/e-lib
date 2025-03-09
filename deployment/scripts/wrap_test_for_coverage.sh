#!/bin/bash

COVDATAFILES_DIR="./testdata/reports/covdatafiles"
OUTPUT_HTML="./testdata/reports/it_coverage.html"

# Setup
rm "$OUTPUT_HTML"
rm -rf "$COVDATAFILES_DIR"
mkdir "$COVDATAFILES_DIR"

# Run test
GOCOVERDIR="$COVDATAFILES_DIR" ./deployment/scripts/integration_test.sh

# Generate coverage data percentage in html
go tool covdata percent -i="$COVDATAFILES_DIR" -o "$COVDATAFILES_DIR/coverage.out"

echo "Completed running the wrap_test_for_coverage.sh"