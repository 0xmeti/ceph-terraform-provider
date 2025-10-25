package provider

import (
    "testing"
)

// TestNewCephClient tests the client creation
func TestNewCephClient(t *testing.T) {
    client := NewCephClient("https://localhost:8443", "admin", "password")
    
    if client == nil {
        t.Fatal("Expected client to be created, got nil")
    }
    
    if client.Endpoint != "https://localhost:8443" {
        t.Errorf("Expected endpoint 'https://localhost:8443', got '%s'", client.Endpoint)
    }
    
    if client.Username != "admin" {
        t.Errorf("Expected username 'admin', got '%s'", client.Username)
    }
}
