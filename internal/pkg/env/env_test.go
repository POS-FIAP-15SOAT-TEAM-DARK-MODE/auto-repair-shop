package env

import (
	"os"
	"testing"
)

func TestGetString(t *testing.T) {
	// Test with existing environment variable
	_ = os.Setenv("TEST_KEY", "test_value")
	result := GetString("TEST_KEY", "default_value")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}

	// Test with non-existing environment variable
	result = GetString("NON_EXISTING_KEY", "default_value")
	if result != "default_value" {
		t.Errorf("Expected 'default_value', got '%s'", result)
	}

	// Test with empty environment variable
	_ = os.Setenv("EMPTY_KEY", "")
	result = GetString("EMPTY_KEY", "default_value")
	if result != "default_value" {
		t.Errorf("Expected 'default_value', got '%s'", result)
	}
}

func TestGetInt(t *testing.T) {
	// Test with existing environment variable
	_ = os.Setenv("INT_KEY", "1234")
	result := GetInt("INT_KEY", 8080)
	if result != 1234 {
		t.Errorf("Expected 1234, got %d", result)
	}

	// Test with non-existing environment variable
	result = GetInt("NON_EXISTING_INT_KEY", 8080)
	if result != 8080 {
		t.Errorf("Expected 8080, got %d", result)
	}

	// Test with invalid integer value
	_ = os.Setenv("INVALID_INT_KEY", "not_a_number")
	result = GetInt("INVALID_INT_KEY", 8080)
	if result != 8080 {
		t.Errorf("Expected 8080, got %d", result)
	}
}
