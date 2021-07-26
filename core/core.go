package core

import (
	"log"
	"strings"
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
