package lyrics

import (
	"fmt"

	"github.com/samuelstevens/spotifind/core"
)

type SimpleLyricProvider struct {

}

func (l *SimpleLyricProvider) GetLyrics(song *core.Song) (*core.SongWithLyrics, error) {
  fmt.Printf("Need to find lyrics for '%s'\n", song.Title)
  return nil, fmt.Errorf("GetLyrics not implemented")
}
