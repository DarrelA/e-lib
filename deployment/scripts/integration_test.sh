#!/bin/bash

# Testdata files
TESTDATA_JSON_FILES=(
    "./testdata/json/test.booksRequests.json" # Single test file
)

# Base URL (move this to the test data or environment if it changes)
BASE_URL="localhost:3000"

# Run the main application
./e-lib-it &

# Capture the PID of the application
APP_PID=$!

# Wait for the application to start
sleep 3

# Main Test Loop
for i in "${!TESTDATA_JSON_FILES[@]}"; do
    TESTDATA_JSON_FILE=${TESTDATA_JSON_FILES[$i]}

    # Read and loop through all tests in the JSON file
    jq -c '.[]' "$TESTDATA_JSON_FILE" | while read -r test_case; do

        # Extract test case details
        method=$(echo "$test_case" | jq -r '.method')
        url_path=$(echo "$test_case" | jq -r '.url_path')
        url_query_string=$(echo "$test_case" | jq -r '.url_query_string')
        json_body_request=$(echo "$test_case" | jq -c '.json_body_request')

        # Construct the full URL
        full_url="$BASE_URL$url_path"

        # Append query string if it exists and method is GET
        if [[ ! -z "$url_query_string" ]]; then
            full_url="$full_url?$url_query_string"
        fi


        # Execute the request based on the method
        if [[ "$method" == "GET" ]]; then

            curl -s "$full_url" > /dev/null   # GET Request -- Discard Output

        elif [[ "$method" == "POST" ]]; then

            curl -s -X POST -H "Content-Type: application/json" -d "$json_body_request" "$full_url" > /dev/null  # POST Request -- Discard Output
        fi
    done
done

sleep 1

# Initiate graceful shutdown by sending SIGTERM to the application process
kill -SIGTERM "$APP_PID"

# Wait for the application to shutdown gracefully
wait "$APP_PID"

echo "Completed running the integration_test.sh"
