package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// TestFormatJSONHandler covers various scenarios for the formatJSONHandler.
func TestFormatJSONHandler(t *testing.T) {
	// Define test cases using a table-driven approach
	testCases := []struct {
		name               string // Name of the test case
		method             string // HTTP method (POST, GET, etc.)
		inputBody          string // JSON string input in the request body
		expectedStatusCode int    // Expected HTTP status code
		expectedBody       string // Expected pretty-printed JSON string (or error message)
		expectJSONResponse bool   // Whether the response body should be valid JSON
	}{
		{
			name:               "Valid JSON - Simple",
			method:             http.MethodPost,
			inputBody:          `{"key":"value", "number": 123}`,
			expectedStatusCode: http.StatusOK,
			expectedBody: `{
							  "key": "value",
							  "number": 123
							}`,
			expectJSONResponse: true,
		},
		{
			name:               "Valid JSON - Nested",
			method:             http.MethodPost,
			inputBody:          `{"a": [1, {"b": true}], "c": null}`,
			expectedStatusCode: http.StatusOK,
			expectedBody: `{
							  "a": [
								1,
								{
								  "b": true
								}
							  ],
							  "c": null
                           }`,
			expectJSONResponse: true,
		},
		{
			name:   "Valid JSON - Already Formatted",
			method: http.MethodPost,
			inputBody: `{
  "already": "formatted"
}`,
			expectedStatusCode: http.StatusOK,
			expectedBody: `{
  "already": "formatted"
}`,
			expectJSONResponse: true,
		},
		{
			name:               "Invalid JSON - Missing Quote",
			method:             http.MethodPost,
			inputBody:          `{"key: "value"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid JSON provided: invalid character 'v' after object key",
			expectJSONResponse: false,
		},
		{
			name:               "Invalid JSON - Trailing Comma",
			method:             http.MethodPost,
			inputBody:          `{"key": "value",}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid JSON provided: invalid character '}' looking for beginning of object key string",
			expectJSONResponse: false,
		},
		//{
		//	name:               "Empty Body",
		//	method:             http.MethodPost,
		//	inputBody:          "",
		//	expectedStatusCode: http.StatusBadRequest,
		//	expectedBody:       "Empty request body",
		//	expectJSONResponse: false,
		//},
		{
			name:               "Incorrect Method - GET",
			method:             http.MethodGet,
			inputBody:          `{"key":"value"}`, // Body ignored for GET test
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedBody:       "Invalid request method. Only POST is allowed.",
			expectJSONResponse: false,
		},
		{
			name:               "Incorrect Method - PUT",
			method:             http.MethodPut,
			inputBody:          `{"key":"value"}`, // Body ignored for PUT test
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedBody:       "Invalid request method. Only POST is allowed.",
			expectJSONResponse: false,
		},
	}

	// Iterate over test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a request
			var requestBody io.Reader
			if tc.inputBody != "" {
				requestBody = strings.NewReader(tc.inputBody)
			}
			req, err := http.NewRequest(tc.method, "/formatjson", requestBody)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}
			// Set content type for POST requests with bodies
			if tc.method == http.MethodPost && tc.inputBody != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Create an HTTP handler from our function
			handler := http.HandlerFunc(formatJSONHandler)

			// Serve the HTTP request to our handler
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tc.expectedStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tc.expectedStatusCode)
			}

			// Check the response body
			// Trim whitespace for non-JSON error messages for robustness
			responseBody := rr.Body.String()
			expectedBody := tc.expectedBody
			if !tc.expectJSONResponse {
				responseBody = strings.TrimSpace(responseBody)
				expectedBody = strings.TrimSpace(expectedBody)
			}

			// For successful JSON, compare the actual structure for robustness
			if tc.expectJSONResponse && tc.expectedStatusCode == http.StatusOK {
				var gotJSON, wantJSON interface{}

				errGot := json.Unmarshal(rr.Body.Bytes(), &gotJSON)
				if errGot != nil {
					t.Fatalf("Could not unmarshal actual response body: %v\nBody: %s", errGot, rr.Body.String())
				}

				errWant := json.Unmarshal([]byte(tc.expectedBody), &wantJSON)
				if errWant != nil {
					t.Fatalf("Could not unmarshal expected response body: %v\nBody: %s", errWant, tc.expectedBody)
				}

				if !reflect.DeepEqual(gotJSON, wantJSON) {
					// Use MarshalIndent for clearer diffs in test output
					gotFormatted, _ := json.MarshalIndent(gotJSON, "", "  ")
					wantFormatted, _ := json.MarshalIndent(wantJSON, "", "  ")
					t.Errorf("handler returned unexpected body structure:\ngot:\n%s\nwant:\n%s",
						string(gotFormatted), string(wantFormatted))
				}
			} else if !tc.expectJSONResponse {
				// For errors, check if the expected message is contained within the response
				// This is slightly less brittle than exact match for error messages
				if !strings.Contains(responseBody, expectedBody) {
					t.Errorf("handler returned unexpected body:\ngot: %q\nwant (substring): %q",
						responseBody, expectedBody)
				}
			}

			// Check Content-Type header for successful JSON responses
			if tc.expectJSONResponse && tc.expectedStatusCode == http.StatusOK {
				expectedContentType := "application/json"
				if ctype := rr.Header().Get("Content-Type"); ctype != expectedContentType {
					t.Errorf("handler returned wrong Content-Type: got %q want %q",
						ctype, expectedContentType)
				}
			}
		})
	}
}

// Helper function to compare JSON bodies ignoring formatting and key order (alternative approach)
// Not strictly needed if using reflect.DeepEqual as above, but good to know.
func compareJSON(t *testing.T, actual []byte, expected string) bool {
	t.Helper() // Marks this as a test helper function

	var actualJSON, expectedJSON interface{}

	err := json.Unmarshal(actual, &actualJSON)
	if err != nil {
		t.Errorf("Failed to unmarshal actual JSON: %v\nBody: %s", err, string(actual))
		return false
	}

	err = json.Unmarshal([]byte(expected), &expectedJSON)
	if err != nil {
		t.Errorf("Failed to unmarshal expected JSON: %v\nBody: %s", err, expected)
		return false
	}

	return reflect.DeepEqual(actualJSON, expectedJSON)
}

// Example usage of compareJSON (if not using direct DeepEqual in main test)
// if tc.expectJSONResponse && tc.expectedStatusCode == http.StatusOK {
//     if !compareJSON(t, rr.Body.Bytes(), tc.expectedBody) {
//         t.Errorf("handler returned unexpected body structure:\ngot:\n%s\nwant:\n%s",
//             rr.Body.String(), tc.expectedBody)
//     }
// }
