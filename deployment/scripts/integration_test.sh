#!/bin/bash

OUTPUT_JSON_FILE="./testdata/reports/responses.json"

# Testdata files
TESTDATA_JSON_FILES=(
    "./testdata/json/loanBooks.json" # Single test file
)

# Base URL (move this to the test data or environment if it changes)
BASE_URL="localhost:3000"

# Run the main application
./e-lib-it &

# Capture the PID of the application
APP_PID=$!

# Wait for the application to start
sleep 3

# Function to parse JSON using jq
parse_json() {
    echo "$1" | jq -r "$2"
}

# Initialize the output JSON file with an empty array
echo "[]" > "$OUTPUT_JSON_FILE"

# Create response_body.txt file
touch response_body.txt

# Main Test Loop
for i in "${!TESTDATA_JSON_FILES[@]}"; do
    TESTDATA_JSON_FILE=${TESTDATA_JSON_FILES[$i]}

    # Get the current date and time in the desired format
    DATETIME=$(date +"%Y-%m-%dT%H:%M:%S%z")

    # Add DateTime and initialize array of tests for a test file
    jq --arg datetime "$DATETIME" --arg file "$TESTDATA_JSON_FILE" \
       '. += [{"DateTime": $datetime, "TestFile": $file, "Tests": []}]' \
       "$OUTPUT_JSON_FILE" > temp.json && mv temp.json "$OUTPUT_JSON_FILE"

    # Read and loop through all tests in the JSON file
    jq -c '.[]' "$TESTDATA_JSON_FILE" | while read -r test_case; do

        # Extract test case details
        test_name=$(echo "$test_case" | jq -r '.test_name')
        method=$(echo "$test_case" | jq -r '.method')
        url_path=$(echo "$test_case" | jq -r '.url_path')
        url_query_string=$(echo "$test_case" | jq -r '.url_query_string')
        expected_status_code=$(echo "$test_case" | jq -r '.expected_status_code')
        json_body_request=$(echo "$test_case" | jq -c '.json_body_request')

        # Construct the full URL
        full_url="$BASE_URL$url_path"

        # Append query string if it exists and method is GET
        if [[ ! -z "$url_query_string" ]]; then
            full_url="$full_url?$url_query_string"
        fi
        echo "Full URL: $full_url"

        # Execute the request based on the method
        if [[ "$method" == "GET" ]]; then

            response_status=$(curl -s \
                -o response_body.txt \
                -w "%{http_code}" \
                "$full_url")    # GET request

        elif [[ "$method" == "POST" ]]; then

            response_status=$(curl -s \
                -o response_body.txt \
                -w "%{http_code}" \
                -X POST \
                -H "Content-Type: application/json" \
                -d "$json_body_request" \
                "$full_url") # POST request

        else
            echo "Error: Invalid method '$method' in test case '$test_name'"
            continue  # Skip to the next test case
        fi

        # Read the response body
        response_body=$(cat response_body.txt)

        # Append the test result to the output JSON file
        jq --arg test_name "$test_name" \
           --arg method "$method" \
           --arg url "$full_url" \
           --arg expected_status_code "$expected_status_code" \
           --arg response_status "$response_status" \
           --arg response_body "$response_body" \
           '.[-1].Tests += [{
                "TestName": $test_name,
                "Method": $method,
                "URL": $url,
                "ExpectedStatusCode": $expected_status_code,
                "ResponseStatus": $response_status,
                "ResponseBody": $response_body
            }]' \
           "$OUTPUT_JSON_FILE" > temp.json && mv temp.json "$OUTPUT_JSON_FILE"
    done
done

# Cleanup temporary file
rm response_body.txt

sleep 1

# Initiate graceful shutdown by sending SIGTERM to the application process
kill -SIGTERM "$APP_PID"

# Wait for the application to shutdown gracefully
wait "$APP_PID"

echo "Completed running the integration_test.sh"
