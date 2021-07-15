package auth

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/samuelstevens/spotifind/spotify/api"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

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
	resp, err := http.PostForm("https://accounts.spotify.com/api/token", postBody)
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
		return fmt.Errorf("Could not parse json: %w", err)
	}

	if result.AccessToken == "" {
		return fmt.Errorf("Could not get access_token: %s", string(respBody))
	}
	a.accessToken = result.AccessToken
	if a.accessToken == "" {
		fmt.Println("FUCK THE THIRD")
	}

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

func (a *SimpleCliAuthenticator) RefreshToken() (string, error) {
	if a.refreshToken == "" {
		err := a.authenticate()
		if err != nil {
			return "", err
		}
	}

	return a.refreshToken, nil
}

func (a *SimpleCliAuthenticator) Refresh() error {
	return fmt.Errorf("not implemented")
}

type CachedAuthenticator struct {
	Authenticator api.Authenticator
	CachePath     string
}

type credentials struct {
	AccessToken  string
	RefreshToken string
}

func (a *CachedAuthenticator) cacheCredentials(creds credentials) error {
	if creds.AccessToken == "" {
		return fmt.Errorf("AccessToken empty: %v\n", creds)
	}
	if creds.RefreshToken == "" {
		return fmt.Errorf("RefreshToken empty: %v\n", creds)
	}

	b, err := json.Marshal(creds)
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
	var creds credentials

	contents, err := os.ReadFile(a.CachePath)
	if err != nil {
		// ask underlying Authenticator for the access token
		accessToken, err := a.Authenticator.AccessToken()
		if err != nil {
			return "", fmt.Errorf("Could not get access token from root authenticator: %w", err)
		}
		creds.AccessToken = accessToken

		refreshToken, err := a.Authenticator.RefreshToken()
		if err != nil {
			return "", fmt.Errorf("Could not get refresh token from root authenticator: %w", err)
		}
		creds.RefreshToken = refreshToken

		a.cacheCredentials(creds)

		return accessToken, nil
	}

	err = json.Unmarshal(contents, &creds)
	if err != nil {
		// remove file since it is corrupted
		panic("handling corrupted files not implemented")
		// ask underlying Authenticator for the access token
		// accessToken, err := a.Authenticator.GetAccessToken()
		// if err != nil {
		// 	return "", fmt.Errorf("Could not get token from root authenticator: %w", err)
		// }
		// creds.AccessToken = accessToken
		// a.cacheCredentials(creds)
	}

	if creds.AccessToken == "" {
		// remove file since it is corrupted
		os.Remove(a.CachePath)
		// ask underlying Authenticator for the access token
		accessToken, err := a.Authenticator.AccessToken()
		if err != nil {
			return "", fmt.Errorf("Could not get token from root authenticator: %w", err)
		}
		creds.AccessToken = accessToken
		a.cacheCredentials(creds)
		return accessToken, nil
	}

	return creds.AccessToken, nil
}

func (a *CachedAuthenticator) Refresh() error {
	panic("not implemented")
}

func (a *CachedAuthenticator) RefreshToken() (string, error) {
	panic("not implemented")
}
