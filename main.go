package main

import (
	"flag"
	"fmt"
	"github.com/samuelstevens/spotifind/core"
	"github.com/samuelstevens/spotifind/lyrics"
	"github.com/samuelstevens/spotifind/spotify/api"
	"github.com/samuelstevens/spotifind/spotify/auth"
	"log"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatalf("Must provide a lyric!\n")
	}

	lyric := flag.Arg(0)

	if lyric == "" {
		log.Fatalf("Lyric must be non-empty!\n")
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
		log.Fatalf("Error finding songs: %v\n", err)
	}

	fmt.Printf("Lyric '%s' is in these songs: %v\n", lyric, songs)
}
