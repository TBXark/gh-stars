package github

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

	"github.com/TBXark/gh-stars/internal/domain"
)

const (
	defaultBaseURL = "https://api.github.com"
	userAgent      = "gh-stars-gui"
)

type Client interface {
	ListStarred(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error)
	GetRepoDetails(ctx context.Context, fullName, token string) (domain.RepoDetails, error)
}

type HTTPClient struct {
	baseURL string
	http    *http.Client
}

func NewClient(httpClient *http.Client) *HTTPClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &HTTPClient{baseURL: defaultBaseURL, http: httpClient}
}

func (c *HTTPClient) ListStarred(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("username is required")
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 100
	}

	var all []domain.Repo
	for page := 1; ; page++ {
		endpoint := fmt.Sprintf("%s/users/%s/starred?per_page=%d&page=%d", c.baseURL, url.PathEscape(username), perPage, page)
		var resp []repoResponse
		_, err := c.getJSON(ctx, endpoint, token, &resp)
		if err != nil {
			return nil, err
		}
		if len(resp) == 0 {
			break
		}
		for _, r := range resp {
			all = append(all, r.toDomain())
		}
		if len(resp) < perPage {
			break
		}
	}
	return all, nil
}

func (c *HTTPClient) GetRepoDetails(ctx context.Context, fullName, token string) (domain.RepoDetails, error) {
	if strings.TrimSpace(fullName) == "" {
		return domain.RepoDetails{}, errors.New("repo full name is required")
	}
	owner, repo, err := splitFullName(fullName)
	if err != nil {
		return domain.RepoDetails{}, err
	}
	endpoint := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, url.PathEscape(owner), url.PathEscape(repo))
	var resp repoDetailsResponse
	_, err = c.getJSON(ctx, endpoint, token, &resp)
	if err != nil {
		return domain.RepoDetails{}, err
	}
	return resp.toDomain(), nil
}

func splitFullName(fullName string) (string, string, error) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.New("repo full name must be owner/name")
	}
	return parts[0], parts[1], nil
}

type repoResponse struct {
	FullName    string    `json:"full_name"`
	HTMLURL     string    `json:"html_url"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	Stars       int       `json:"stargazers_count"`
	Forks       int       `json:"forks_count"`
	UpdatedAt   time.Time `json:"updated_at"`
	Private     bool      `json:"private"`
}

func (r repoResponse) toDomain() domain.Repo {
	return domain.Repo{
		FullName:    r.FullName,
		HTMLURL:     r.HTMLURL,
		Description: r.Description,
		Language:    r.Language,
		Stars:       r.Stars,
		Forks:       r.Forks,
		UpdatedAt:   r.UpdatedAt,
		Private:     r.Private,
	}
}

type repoDetailsResponse struct {
	FullName      string    `json:"full_name"`
	HTMLURL       string    `json:"html_url"`
	Description   string    `json:"description"`
	Language      string    `json:"language"`
	Homepage      string    `json:"homepage"`
	DefaultBranch string    `json:"default_branch"`
	Topics        []string  `json:"topics"`
	Stars         int       `json:"stargazers_count"`
	Forks         int       `json:"forks_count"`
	Watchers      int       `json:"watchers_count"`
	OpenIssues    int       `json:"open_issues_count"`
	Size          int       `json:"size"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedAt     time.Time `json:"created_at"`
	PushedAt      time.Time `json:"pushed_at"`
	Private       bool      `json:"private"`
	License       *struct {
		Name string `json:"name"`
	} `json:"license"`
}

func (r repoDetailsResponse) toDomain() domain.RepoDetails {
	license := ""
	if r.License != nil {
		license = r.License.Name
	}
	return domain.RepoDetails{
		FullName:      r.FullName,
		HTMLURL:       r.HTMLURL,
		Description:   r.Description,
		Language:      r.Language,
		Homepage:      r.Homepage,
		DefaultBranch: r.DefaultBranch,
		License:       license,
		Topics:        r.Topics,
		Stars:         r.Stars,
		Forks:         r.Forks,
		Watchers:      r.Watchers,
		OpenIssues:    r.OpenIssues,
		Size:          r.Size,
		UpdatedAt:     r.UpdatedAt,
		CreatedAt:     r.CreatedAt,
		PushedAt:      r.PushedAt,
		Private:       r.Private,
	}
}

type apiError struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
}

func (c *HTTPClient) getJSON(ctx context.Context, endpoint, token string, target any) (http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		msg := strings.TrimSpace(string(body))
		var apiErr apiError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			msg = apiErr.Message
			if apiErr.DocumentationURL != "" {
				msg = msg + " (" + apiErr.DocumentationURL + ")"
			}
		}
		if msg == "" {
			msg = "unknown error"
		}
		return nil, fmt.Errorf("github api error: %s: %s", resp.Status, msg)
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(target); err != nil {
		return nil, err
	}
	return resp.Header, nil
}
