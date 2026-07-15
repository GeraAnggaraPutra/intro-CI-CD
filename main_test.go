package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFileServerHandler(t *testing.T) {
	// Setup temporary static directory and dummy file
	tmpDir, err := os.MkdirTemp("", "static_test_*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dummyContent := "Hello NATA HR!"
	tmpFile := filepath.Join(tmpDir, "index.html")
	if err := os.WriteFile(tmpFile, []byte(dummyContent), 0644); err != nil {
		t.Fatalf("failed to create dummy file: %v", err)
	}

	// Initialize the handler with the temp directory
	handler := NewFileServerHandler(tmpDir)

	// Define Table-Driven Test Cases
	tests := []struct {
		name           string
		requestPath    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success serving index.html",
			requestPath:    "/",
			expectedStatus: http.StatusOK,
			expectedBody:   dummyContent,
		},
		{
			name:           "Status 404 on non-existent file",
			requestPath:    "/not-found.html",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "404 page not found\n",
		},
	}

	// Execute Test Cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tc.requestPath, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tc.expectedBody {
				t.Errorf("expected body %q, got %q", tc.expectedBody, rr.Body.String())
			}
		})
	}
}
