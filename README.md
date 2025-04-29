# JSON Validator/Formatter API

## Description

This is a simple Go web service that provides an HTTP endpoint to validate and format JSON data. It accepts potentially messy or unformatted JSON via a POST request and returns either a well-formatted (pretty-printed) JSON string or an error message if the input JSON is invalid.

## Technology Stack

*   **Language:** Go (v1.24+)
*   **Core Libraries:** Go Standard Library (`net/http`, `encoding/json`)
*   **Containerization:** Docker, Docker Compose
*   **CI:** GitHub Actions

## Prerequisites

*   **Go:** Version 1.24 or later (required for running natively)
*   **Docker:** Latest version (required for running via Docker)
*   **Docker Compose:** Latest version (required for running via Docker Compose)
*   **curl** or similar HTTP client (for testing the endpoint)

## Running Locally

There are two primary ways to run the application locally:

**1. Using Docker Compose (Recommended)**

This method builds the Docker image and runs the container. It's the easiest way to get started and mirrors a containerized deployment approach.

```bash
# Navigate to the project's root directory (where docker-compose.yml is)
cd json-validator-formatter

# Build the image (if not already built) and start the service in detached mode (-d)
docker-compose up --build -d

# To view logs:
docker-compose logs -f

# To stop the service:
docker-compose down
```
**2. Using Go natively**

This method requires a local Go installation.
```bash
# Navigate to the project's root directory
cd json-validator-formatter

# Run the main application file
go run ./cmd/server/main.go
```
The service will be available at http://localhost:8080.

## Usage

Send a POST request to the /formatjson endpoint with your JSON data in the request body.

**Example Request (Valid JSON):**
```http request
curl -X POST -H "Content-Type: application/json" --data '{"name": "example",
    "data": [1, 2, {"nested": true}], "valid":true}' http://localhost:8080/formatjson
```

**Expected Successful Response (Status 200 OK):**
```json
{
  "data": [
    1,
    2,
    {
      "nested": true
    }
  ],
  "name": "example",
  "valid": true
}
```
**Example Request (Invalid JSON - missing closing brace):**
```http request
curl -X POST -H "Content-Type: application/json" --data '{"name": "example",
    "value": 123' http://localhost:8080/formatjson
```

**Expected Error Response (Status 400 Bad Request):**  
`Invalid JSON provided: unexpected end of JSON input`

**Example Request (Empty Body):**
```http request
curl -X POST http://localhost:8080/formatjson
```

**Expected Error Response (Status 400 Bad Request):**  
`Empty request body`

