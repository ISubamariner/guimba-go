package helpers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// AssertStatus checks that the response status code matches expected.
func AssertStatus(t *testing.T, rr *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if rr.Code != expected {
		t.Errorf("expected status %d, got %d (body: %s)", expected, rr.Code, rr.Body.String())
	}
}

// AssertJSONKey checks that a JSON response body contains a key with an expected value.
func AssertJSONKey(t *testing.T, rr *httptest.ResponseRecorder, key string, expected interface{}) {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	val, ok := body[key]
	if !ok {
		t.Errorf("expected key %q in response body, but not found", key)
		return
	}
	if val != expected {
		t.Errorf("expected %q = %v, got %v", key, expected, val)
	}
}
