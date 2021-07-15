package main

import (
	"flag"
	"fmt"
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

func main() {
	lyric, err := cli()
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	fmt.Printf("Searching for songs with lyric '%s'\n", lyric)

	songProvider := api.SimpleSongProvider{
		Authenticator: &auth.CachedAuthenticator{
			Authenticator: &auth.SimpleCliAuthenticator{},
			CachePath:     "/Users/samstevens/.spotifind.cache",
		},
	}
	lyricProvider := lyrics.SimpleLyricProvider{}

	songs, err := core.FindSongsWithLyric(core.Lyric(lyric), &songProvider, &lyricProvider)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Printf("Lyric '%s' is in these songs: %v\n", lyric, songs)
}
