package core

import (
	"fmt"
	"log"
	"strings"
)

type Song struct {
	Title   string
	Uri     string
	Artists []string
}

func (s *Song) Formatted() string {
	var artistStr string

	if len(s.Artists) > 1 {
		mostArtists := strings.Join(s.Artists[0:len(s.Artists)-1], ", ")
		lastArtist := s.Artists[len(s.Artists)-1]
		artistStr = fmt.Sprintf("%s and %s", mostArtists, lastArtist)
	} else {
    artistStr = s.Artists[0]
  }

	return fmt.Sprintf("'%s' by %s", s.Title, artistStr)
}

type SongWithLyrics struct {
	Song
	Lyrics []string
}

func (s *SongWithLyrics) HasLyric(lyric string) bool {
	if s == nil {
		return false
	}
	for _, line := range s.Lyrics {
		if strings.Contains(line, lyric) {
			return true
		}
	}

	totalLyrics := strings.Join(s.Lyrics, " ")

	if strings.Contains(totalLyrics, lyric) {
		return true
	}

	// TODO: needs additional checks
	return false
}

func FindSongsWithLyric(lyric string, songProvider SongProvider, lyricProvider LyricProvider) ([]SongWithLyrics, error) {
	songs := make(chan Song)
	go songProvider.GetSongs(songs)

	result := []SongWithLyrics{}

	for song := range songs {
		songWithLyrics, err := lyricProvider.GetLyrics(&song)
		if err != nil {
			log.Printf("Could not get lyrics for song %v: %v", song, err)
			break
		}
		if songWithLyrics.HasLyric(lyric) {
			result = append(result, *songWithLyrics)
		}
	}

	return result, nil
}

type SongProvider interface {
	GetSongs(out chan Song)
}

type LyricProvider interface {
	GetLyrics(song *Song) (*SongWithLyrics, error)
}
