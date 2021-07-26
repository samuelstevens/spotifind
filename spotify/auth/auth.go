package auth

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/samuelstevens/spotifind/spotify/api"
)

func postFormWithTimeout(url string, data url.Values, timeout time.Duration) (*http.Response, error) {
	client := http.Client{Timeout: timeout}

	return client.PostForm(url, data)
}

type SimpleCliAuthenticator struct {
	accessToken  string
	refreshToken string
}

func (a *SimpleCliAuthenticator) authenticate() error {
	u := &url.URL{
		Scheme: "https",
		Host:   "accounts.spotify.com",
		Path:   "authorize",
	}

	state := "123" // should be random value

	query := url.Values{}
	query.Add("client_id", api.ClientId)
	query.Add("redirect_uri", api.RedirectUri)
	query.Add("response_type", "code")
	query.Add("state", state)
	query.Add("scope", "user-library-read")

	u.RawQuery = query.Encode()

	fmt.Printf("Click on this link:\n\n%s\n\nThen paste the URL back here: ", u.String())

	reader := bufio.NewReader(os.Stdin)
	redirect, err := reader.ReadString('\n')

	if err != nil {
		return fmt.Errorf("Could not read redirect URL from stdin: %w", err)
	}
	redirect = strings.TrimSuffix(redirect, "\n")

	u, err = url.Parse(redirect)
	if err != nil {
		return fmt.Errorf("Pasted URL was not a valid URL: %w", err)
	}

	authCode := u.Query().Get("code")

	postBody := url.Values{
		"grant_type":    {"authorization_code"},
		"response_type": {"code"},
		"code":          {authCode},
		"redirect_uri":  {api.RedirectUri},
		"client_id":     {api.ClientId},
		"client_secret": {api.ClientSecret},
	}

	// Request refresh and access tokens
	resp, err := postFormWithTimeout("https://accounts.spotify.com/api/token", postBody, time.Second*5)
	if err != nil {
		return fmt.Errorf("Could not exchange access code for tokens: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Could not read body: %w", err)
	}

	type Resp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}

	var result Resp

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return fmt.Errorf("Could not parse json (%s): %w", string(respBody), err)
	}

	if result.AccessToken == "" {
		return fmt.Errorf("Could not get access_token: %s", string(respBody))
	}
	a.accessToken = result.AccessToken

	if result.RefreshToken == "" {
		return fmt.Errorf("Could not get refresh_token")
	}
	a.refreshToken = result.RefreshToken

	return nil
}

func (a *SimpleCliAuthenticator) AccessToken() (string, error) {
	if a.accessToken == "" {
		err := a.authenticate()
		if err != nil {
			return "", err
		}
	}

	return a.accessToken, nil
}

func (a *SimpleCliAuthenticator) Refresh() error {
	if a.accessToken == "" || a.refreshToken == "" {
		err := a.authenticate()
		if err != nil {
			return err
		}
	}

	// request a new accessToken
	postBody := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {a.refreshToken},
		"client_id":     {api.ClientId},
		"client_secret": {api.ClientSecret},
	}

	// Request refresh and access tokens
	resp, err := postFormWithTimeout("https://accounts.spotify.com/api/token", postBody, time.Second*5)
	if err != nil {
		return fmt.Errorf("Could not exchange refresh token for access token: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Could not read body: %w", err)
	}

	type Resp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		ExpiresIn   int    `json:"expires_in"`
	}

	var result Resp

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return fmt.Errorf("Could not parse json: %w", err)
	}

	if result.AccessToken == "" {
		return fmt.Errorf("Could not get access_token: %s", string(respBody))
	}

	if a.accessToken == result.AccessToken {
		return fmt.Errorf("New access token is the same: %s", a.accessToken)
	}

	a.accessToken = result.AccessToken

	return nil
}

type CachedAuthenticator struct {
	Authenticator api.Authenticator
	CachePath     string
}

func (a *CachedAuthenticator) saveAccessToken(accessToken string) error {
	if accessToken == "" {
		return fmt.Errorf("Will not save empty access token")
	}

	b, err := json.Marshal(accessToken)
	if err != nil {
		return fmt.Errorf("Could not convert credentials to json: %w", err)
	}
	err = os.WriteFile(a.CachePath, b, 0644)
	if err != nil {
		return fmt.Errorf("Could not write json to file: %w", err)
	}

	return nil
}

func (a *CachedAuthenticator) AccessToken() (string, error) {
	var accessToken string
	contents, err := os.ReadFile(a.CachePath)
	if err != nil {
		// ask underlying Authenticator for the access token
		accessToken, err := a.Authenticator.AccessToken()
		if err != nil {
			return "", fmt.Errorf("Could not get access token from root authenticator: %w", err)
		}
		a.saveAccessToken(accessToken)

		return accessToken, nil
	}
	err = json.Unmarshal(contents, &accessToken)
	if err != nil {
		// remove file since it is corrupted
		os.Remove(a.CachePath)
		// ask underlying Authenticator for the access token
		accessToken, err := a.Authenticator.AccessToken()
		if err != nil {
			return "", fmt.Errorf("Could not get access token from root authenticator: %w", err)
		}
		a.saveAccessToken(accessToken)

		return accessToken, nil
	}

	if accessToken == "" {
		// remove file since it is corrupted
		os.Remove(a.CachePath)
		// ask underlying Authenticator for the access token
		accessToken, err := a.Authenticator.AccessToken()
		if err != nil {
			return "", fmt.Errorf("Could not get token from root authenticator: %w", err)
		}
		a.saveAccessToken(accessToken)
		return accessToken, nil
	}

	return accessToken, nil
}

func (a *CachedAuthenticator) Refresh() error {
	err := a.Authenticator.Refresh()
	if err != nil {
		return fmt.Errorf("Root authenticator failed to refresh: %w", err)
	}

	accessToken, err := a.Authenticator.AccessToken()
	if err != nil {
		return fmt.Errorf("Root authenticator failed to provide an access token: %w", err)
	}

	err = a.saveAccessToken(accessToken)
	if err != nil {
		fmt.Printf("Could not cache access token: %s", err.Error())
	}

	return nil
}
