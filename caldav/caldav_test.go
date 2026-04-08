package caldav

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
)

// TestNewHandlerNilKeys tests that NewHandler panics with nil keys
func TestNewHandlerNilKeys(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewHandler() should panic when no private keys provided")
		}
	}()

	NewHandler(nil, nil, nil)
}

// TestNewHandlerValidKeys tests handler creation with valid keys
func TestNewHandlerValidKeys(t *testing.T) {
	// Create a test key
	entity, err := openpgp.NewEntity("Test User", "test", "test@example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create test key: %v", err)
	}

	handler := NewHandler(nil, openpgp.EntityList{entity}, nil)
	if handler == nil {
		t.Error("NewHandler() returned nil")
	}
}

// TestHandlerReturns501 tests that the handler returns 501 Not Implemented
func TestHandlerReturns501(t *testing.T) {
	entity, err := openpgp.NewEntity("Test User", "test", "test@example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create test key: %v", err)
	}

	handler := NewHandler(nil, openpgp.EntityList{entity}, nil)
	if handler == nil {
		t.Fatal("NewHandler() returned nil")
	}

	// Create test request
	req := httptest.NewRequest("PROPFIND", "/calendars/default", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 501 Not Implemented
	if w.Code != http.StatusNotImplemented {
		t.Errorf("Expected status 501, got %d", w.Code)
	}

	// Should mention implementation in progress
	if body := w.Body.String(); body == "" {
		t.Error("Expected non-empty response body")
	}
}
