package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	baseURL    string
	token      string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{},
	}
}

func (c *Client) SetAPIKey(key string) {
	c.apiKey = key
}

func (c *Client) request(method, endpoint string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal error: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) Authenticate(email, password string) (*AuthResponse, error) {
	body := &AuthRequest{
		Identity: email,
		Password: password,
	}
	data, err := c.request("POST", "/api/collections/users/auth-with-password", body)
	if err != nil {
		return nil, err
	}
	var resp AuthResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return &resp, nil
}

func (c *Client) ListTasks() ([]Task, error) {
	data, err := c.request("GET", "/api/tasks", nil)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return tasks, nil
}

func (c *Client) GetTask(id string) (*Task, error) {
	data, err := c.request("GET", "/api/tasks/"+id, nil)
	if err != nil {
		return nil, err
	}
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return &task, nil
}

func (c *Client) CreateTask(req *CreateTaskRequest) (*Task, error) {
	data, err := c.request("POST", "/api/tasks", req)
	if err != nil {
		return nil, err
	}
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return &task, nil
}

func (c *Client) UpdateTask(id string, req *UpdateTaskRequest) (*Task, error) {
	data, err := c.request("PATCH", "/api/tasks/"+id, req)
	if err != nil {
		return nil, err
	}
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return &task, nil
}

func (c *Client) SubmitTask(id string) (*Task, error) {
	data, err := c.request("POST", "/api/tasks/"+id+"/submit", nil)
	if err != nil {
		return nil, err
	}
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return &task, nil
}

func (c *Client) GradeTask(id string, req *GradeRequest) (*Task, error) {
	data, err := c.request("POST", "/api/tasks/"+id+"/grade", req)
	if err != nil {
		return nil, err
	}
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return &task, nil
}

func (c *Client) DeleteTask(id string) error {
	_, err := c.request("DELETE", "/api/tasks/"+id, nil)
	return err
}

func (c *Client) GetMe() (*User, error) {
	data, err := c.request("GET", "/api/me", nil)
	if err != nil {
		return nil, err
	}
	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return &user, nil
}