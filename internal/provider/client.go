package provider

import (
    "bytes"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// CephClient handles communication with Ceph API
type CephClient struct {
    Endpoint   string
    Username   string
    Password   string
    Token      string
    HTTPClient *http.Client
}

// AuthRequest is the login request structure
type AuthRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

// AuthResponse contains the authentication token
type AuthResponse struct {
    Token string `json:"token"`
}

// PoolCreateRequest is the structure for creating a pool
type PoolCreateRequest struct {
    Pool        string `json:"pool"`
    PoolType    string `json:"pool_type"`
    PgNum       int    `json:"pg_num,omitempty"`
    PgpNum      int    `json:"pgp_num,omitempty"`
    Size        int    `json:"size,omitempty"`
    Application string `json:"application,omitempty"`
}

// NewCephClient creates a new Ceph API client
func NewCephClient(endpoint, username, password string) *CephClient {
    return &CephClient{
        Endpoint: endpoint,
        Username: username,
        Password: password,
        HTTPClient: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{
                    InsecureSkipVerify: true, // For self-signed certificates
                },
            },
        },
    }
}

// Authenticate logs in to Ceph and gets a token
func (c *CephClient) Authenticate() error {
    authReq := AuthRequest{
        Username: c.Username,
        Password: c.Password,
    }
    
    body, _ := json.Marshal(authReq)
    req, err := http.NewRequest("POST", c.Endpoint+"/api/auth", bytes.NewBuffer(body))
    if err != nil {
        return fmt.Errorf("failed to create auth request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return fmt.Errorf("authentication request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(bodyBytes))
    }
    
    var authResp AuthResponse
    if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
        return fmt.Errorf("failed to decode auth response: %w", err)
    }
    
    c.Token = authResp.Token
    return nil
}

// CreatePool creates a new Ceph pool
func (c *CephClient) CreatePool(poolReq PoolCreateRequest) error {
    if c.Token == "" {
        if err := c.Authenticate(); err != nil {
            return err
        }
    }
    
    body, _ := json.Marshal(poolReq)
    req, err := http.NewRequest("POST", c.Endpoint+"/api/pool", bytes.NewBuffer(body))
    if err != nil {
        return fmt.Errorf("failed to create pool request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.Token)
    
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return fmt.Errorf("create pool request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("failed to create pool with status %d: %s", resp.StatusCode, string(bodyBytes))
    }
    
    return nil
}

// GetPool retrieves information about a pool
func (c *CephClient) GetPool(poolName string) (map[string]interface{}, error) {
    if c.Token == "" {
        if err := c.Authenticate(); err != nil {
            return nil, err
        }
    }
    
    req, err := http.NewRequest("GET", c.Endpoint+"/api/pool/"+poolName, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create get pool request: %w", err)
    }
    
    req.Header.Set("Authorization", "Bearer "+c.Token)
    
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("get pool request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("pool not found")
    }
    
    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("failed to get pool with status %d: %s", resp.StatusCode, string(bodyBytes))
    }
    
    var pool map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
        return nil, fmt.Errorf("failed to decode pool response: %w", err)
    }
    
    return pool, nil
}

// DeletePool deletes a Ceph pool
func (c *CephClient) DeletePool(poolName string) error {
    if c.Token == "" {
        if err := c.Authenticate(); err != nil {
            return err
        }
    }
    
    req, err := http.NewRequest("DELETE", c.Endpoint+"/api/pool/"+poolName, nil)
    if err != nil {
        return fmt.Errorf("failed to create delete pool request: %w", err)
    }
    
    req.Header.Set("Authorization", "Bearer "+c.Token)
    
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return fmt.Errorf("delete pool request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("failed to delete pool with status %d: %s", resp.StatusCode, string(bodyBytes))
    }
    
    return nil
}
