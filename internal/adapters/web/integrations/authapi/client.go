package authapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type UserInfo struct{ Username, FirstName, LastName, Email string }
type RegisterInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginResult struct {
	Token string `json:"token"`
}
type RegisterResult struct {
	Data struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"data"`
}
type Client interface {
	GetByEmail(context.Context, string, string) (*UserInfo, error)
	GetBatch(context.Context, string, []string) ([]UserInfo, error)
	GetTokenVersion(context.Context, string, string) (uint, error)
	GetMe(context.Context, string) (*UserInfo, error)
	Register(context.Context, RegisterInput) (RegisterResult, []byte, error)
	Login(context.Context, LoginInput) (LoginResult, error)
	AssignRoles(context.Context, string, string, []string) error
}

func (c *client) Register(ctx context.Context, input RegisterInput) (RegisterResult, []byte, error) {
	var parsed RegisterResult
	body, err := c.doJSONBytes(ctx, http.MethodPost, "/auth/register", "", input, &parsed)
	return parsed, body, err
}

func (c *client) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	var parsed LoginResult
	_, err := c.doJSONBytes(ctx, http.MethodPost, "/auth/login", "", input, &parsed)
	return parsed, err
}

func (c *client) AssignRoles(ctx context.Context, bearerToken, userID string, roles []string) error {
	_, err := c.doJSONBytes(ctx, http.MethodPut, "/user/"+url.PathEscape(userID)+"/roles", bearerToken, map[string]any{"roles": roles}, nil)
	return err
}

type client struct {
	baseURL string
	http    *http.Client
}
type UpstreamError struct {
	Status int
	Detail string
}

var ErrNotFound = errors.New("auth-api-be resource not found")

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("auth-api-be call failed (status %d): %s", e.Status, e.Detail)
}
func NewClient(baseURL string, timeout time.Duration) Client {
	return &client{baseURL: strings.TrimRight(baseURL, "/"), http: &http.Client{Timeout: timeout}}
}

func (c *client) GetByEmail(ctx context.Context, bearerToken, email string) (*UserInfo, error) {
	return c.getUser(ctx, bearerToken, "/user/find-by-email?email="+url.QueryEscape(email), false)
}
func (c *client) GetMe(ctx context.Context, bearerToken string) (*UserInfo, error) {
	return c.getUser(ctx, bearerToken, "/user/me", false)
}
func (c *client) GetBatch(ctx context.Context, bearerToken string, usernames []string) ([]UserInfo, error) {
	if len(usernames) == 0 {
		return nil, nil
	}
	var parsed struct {
		Data []struct {
			ID        string `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
		} `json:"data"`
	}
	err := c.doJSON(ctx, bearerToken, "/user/batch?ids="+url.QueryEscape(strings.Join(usernames, ",")), &parsed)
	if err != nil {
		return nil, err
	}
	output := make([]UserInfo, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		output = append(output, UserInfo{Username: item.ID, FirstName: item.FirstName, LastName: item.LastName, Email: item.Email})
	}
	return output, nil
}
func (c *client) GetTokenVersion(ctx context.Context, bearerToken, userID string) (uint, error) {
	var parsed struct {
		Data struct {
			TokenVersion uint `json:"token_version"`
		} `json:"data"`
	}
	err := c.doJSON(ctx, bearerToken, "/user/"+url.PathEscape(userID)+"/token-version", &parsed)
	return parsed.Data.TokenVersion, err
}

func (c *client) getUser(ctx context.Context, bearerToken, path string, notFoundError bool) (*UserInfo, error) {
	var parsed struct {
		Data struct {
			ID        string `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
		} `json:"data"`
	}
	err := c.doJSONWithNotFound(ctx, bearerToken, path, &parsed)
	if errors.Is(err, ErrNotFound) && !notFoundError {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &UserInfo{Username: parsed.Data.ID, FirstName: parsed.Data.FirstName, LastName: parsed.Data.LastName, Email: parsed.Data.Email}, nil
}
func (c *client) doJSON(ctx context.Context, bearerToken, path string, target any) error {
	return c.doJSONWithNotFound(ctx, bearerToken, path, target)
}
func (c *client) doJSONWithNotFound(ctx context.Context, bearerToken, path string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", bearerToken)
	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if response.StatusCode >= 300 {
		return &UpstreamError{Status: response.StatusCode, Detail: response.Status}
	}
	return json.NewDecoder(response.Body).Decode(target)
}

func (c *client) doJSONBytes(ctx context.Context, method, path, bearerToken string, input any, target any) ([]byte, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	if bearerToken != "" {
		request.Header.Set("Authorization", bearerToken)
	}
	response, err := c.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return responseBody, err
	}
	if response.StatusCode >= 300 {
		return responseBody, &UpstreamError{Status: response.StatusCode, Detail: response.Status}
	}
	if target != nil && len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, target); err != nil {
			return responseBody, err
		}
	}
	return responseBody, nil
}
