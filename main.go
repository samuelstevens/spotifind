package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/samuelstevens/spotifind/core"
	"github.com/samuelstevens/spotifind/lyrics"
	"github.com/samuelstevens/spotifind/spotify/api"
	"github.com/samuelstevens/spotifind/spotify/auth"
)

func cli() (string, error) {
	flag.Parse()
	if flag.NArg() == 0 {
		return "", fmt.Errorf("Must provide a lyric!\n")
	}

	lyric := flag.Arg(0)

	if lyric == "" {
		return "", fmt.Errorf("Lyric must be non-empty!\n")
	}

	return lyric, nil
}

func formatSongsWithLyrics(songs []core.SongWithLyrics) string {
	songStrings := []string{}
	for _, song := range songs {
		songStrings = append(songStrings, song.Formatted())
	}

	if len(songs) > 1 {
		return fmt.Sprintf("%s and %s", strings.Join(songStrings[0:len(songs)-1], ", "), songStrings[len(songs)-1])
	} else {
		return songStrings[0]
	}
}

func main() {
	lyric, err := cli()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	if lyric == "" {
		fmt.Printf("Need to provide a lyric to search for!\n")
		return
	}

	fmt.Printf("Searching for songs with lyric '%s'\n", lyric)

	config, err := core.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %s\n", err)
		return
	}

	songProvider := api.SimpleSongProvider{
		Authenticator: &auth.CachedAuthenticator{
			Authenticator: &auth.SimpleCliAuthenticator{
				ClientId:     config.Spotify.ClientId,
				ClientSecret: config.Spotify.ClientSecret,
				RedirectUri:  config.Spotify.RedirectUri,
			},
			CachePath: "/Users/samstevens/.spotifind.cache",
		},
	}
	lyricProvider := lyrics.AZLyricProvider{}

	songs, err := core.FindSongsWithLyric(lyric, &songProvider, &lyricProvider)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Printf("Lyric '%s' is in %s\n", lyric, formatSongsWithLyrics(songs))
}
