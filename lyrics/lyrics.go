package lyrics

import (
	"github.com/samuelstevens/spotifind/core"
)

type SimpleLyricProvider struct {

}

func (l *SimpleLyricProvider) GetLyrics(core.Song) (core.SongWithLyrics, error) {
  panic("GetLyric() not implemented")
}
