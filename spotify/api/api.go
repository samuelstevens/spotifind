package api

import (
	"encoding/json"
	"fmt"
	"github.com/samuelstevens/spotifind/core"
	"io"
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

type Track struct {
	Name    string   `json:"name"`
	Artists []Artist `json:"artists"`
	Uri     string   `json:"uri"`
}

type Item struct {
	Track Track `json:"track"`
}

type GetTrackResult struct {
	Limit  int    `json:"limit"`
	Next   string `json:"next"`
	Offset int    `json:"offset"`
	Total  int    `json:"total"`
	Items  []Item `json:"items"`
}

func (p *SimpleSongProvider) requestSongs(url string) ([]core.Song, string, error) {
	if p.Authenticator == nil {
		return nil, "", fmt.Errorf("%v needs non-nil authenticator!", p)
	}
	client := &http.Client{}

	accessToken, err := p.Authenticator.AccessToken()
	if err != nil {
		return nil, "", err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if resp.StatusCode == 401 {
			fmt.Printf("Unauthorized: %+v\ngoing to refresh token\n", resp.Header)
			if err = p.Authenticator.Refresh(); err != nil {
				return nil, "", err
			}
			return p.requestSongs(url)
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var result GetTrackResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, "", fmt.Errorf("Could not read json: %w", err)
	}

	songs := []core.Song{}

	for _, item := range result.Items {
		artists := []string{}
		for _, artist := range item.Track.Artists {
			artists = append(artists, artist.Name)
		}

		song := core.Song{
			Title:   item.Track.Name,
			Uri:     item.Track.Uri,
			Artists: artists,
		}

		songs = append(songs, song)
	}

	return songs, result.Next, nil
}

func (p *SimpleSongProvider) GetSongs(out chan core.Song) {
	nextUrl := "https://api.spotify.com/v1/me/tracks"

	for nextUrl != "" {
		var err error
		var songs []core.Song

		songs, nextUrl, err = p.requestSongs(nextUrl)
		if err != nil {
			fmt.Printf("Error in GetSongs: %s", err.Error())
			close(out)
		}

		for _, song := range songs {
			out <- song
		}
	}

	close(out)
}
