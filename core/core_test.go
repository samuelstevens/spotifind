package core

import (
	"testing"
)

type DummySongProvider struct{}

func (s *DummySongProvider) GetSongs() ([]Song, error) {
	return []Song{
		{Title: "Lil Bit", Artists: []string{"Nelly", "Florida Georgia Line"}, Uri: ""},
		{Title: "How Long", Artists: []string{"Charlie Puth"}, Uri: ""},
	}, nil
}

type DummyLyricProvider struct{}

func (l *DummyLyricProvider) GetLyrics(song Song) (SongWithLyrics, error) {
	return SongWithLyrics{
		Song: song,
		Lyrics: []string{
			"",
			"",
		},
	}, nil
}

func TestFindSongsWithLyric(t *testing.T) {
  panic("Not implemented") 
}
