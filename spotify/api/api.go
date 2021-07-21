package api

import (
	"encoding/json"
	"fmt"
	"github.com/samuelstevens/spotifind/core"
	"io"
	"log"
	"net/http"
)

const ClientId = "8c5bfc7c82064304bf0dd0a902618144"
const ClientSecret = "e9ead86c1d7e4a3cba37a7b0cd18c19e"
const RedirectUri = "http://spotifind.com/auth"

type Authenticator interface {
	AccessToken() (string, error)
	Refresh() error
}

type SimpleSongProvider struct {
	Authenticator Authenticator
}

type Artist struct {
	Name string `json:"name"`
}

type Item struct {
	Track core.Song `json:"track"`
}

type GetTrackResult struct {
	Limit  int    `json:"limit"`
	Next   string `json:"next"`
	Offset int    `json:"offset"`
	Total  int    `json:"total"`
	Songs  []Item `json:"items"`
}

func (p *SimpleSongProvider) GetSongs() ([]core.Song, error) {
	if p.Authenticator == nil {
		return nil, fmt.Errorf("%v needs non-nil authenticator!", p)
	}
	client := &http.Client{}

	accessToken, err := p.Authenticator.AccessToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/tracks", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if resp.StatusCode == 401 {
      fmt.Printf("Unauthorized: %+v\ngoing to refresh token\n", resp)
			if err = p.Authenticator.Refresh(); err != nil {
				return nil, err
			}
			return p.GetSongs()
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}

	// var result GetTrackResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("Could not read json: %w", err)
	}

	log.Printf("%+v\n", result)

	return nil, fmt.Errorf("Get Songs not implemented yet")
}
