package core

import (
	"fmt"
	"log"
)

type Song struct {
	Title   string
	Uri     string
	Artists []string
}

type SongWithLyrics struct {
	Song
	Lyrics []string
}

func (s *SongWithLyrics) HasLyric(lyric Lyric) bool {
	panic("HasLyric() not implemented")
}

type Lyric string

func FindSongsWithLyric(lyric Lyric, songProvider SongProvider, lyricProvider LyricProvider) ([]SongWithLyrics, error) {
	songs, err := songProvider.GetSongs()
	if err != nil {
		return nil, fmt.Errorf("Could not get songs: %w", err)
	}

	result := []SongWithLyrics{}

	for _, song := range songs {
		songWithLyrics, err := lyricProvider.GetLyrics(song)
		if err != nil {
			log.Printf("Could not get lyrics for song %v: %v", song, err)
		}
		if songWithLyrics.HasLyric(lyric) {
			result = append(result, songWithLyrics)
		}
	}

	return result, nil
}

type SongProvider interface {
	GetSongs() ([]Song, error)
}

type LyricProvider interface {
	GetLyrics(song Song) (SongWithLyrics, error)
}
