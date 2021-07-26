package lyrics

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/samuelstevens/spotifind/core"
)

type AZLyricProvider struct {
}

func (l *AZLyricProvider) findSongUrl(doc *html.Node) (*url.URL, error) {
	// Needs to find all <tr> elements with .visitedlyr class
	// Recursively look for nodes with ElementNode NodeType with Attribute class and .visitedlyr
	for child := doc.FirstChild; child != nil; child = child.NextSibling {
		url, _ := l.findSongUrl(child)
		if url != nil {
			return url, nil
		}
	}

	// check this element itself for a url
	if doc.Type == html.ElementNode && doc.DataAtom == atom.A {
		hasClassAttr := false
		for _, attr := range doc.Parent.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, "visitedlyr") {
				hasClassAttr = true
			}
		}

		if hasClassAttr {
			// extract the url
			for _, attr := range doc.Attr {
				if attr.Key == "href" {
          songUrl, err := url.Parse(attr.Val)
          return songUrl, err
				}
			}
		}
	}

	return nil, fmt.Errorf("Could not find a url in %v", doc)
}

func (l *AZLyricProvider) GetLyrics(song *core.Song) (*core.SongWithLyrics, error) {
	searchUrl := &url.URL{
		Scheme: "https",
		Host:   "search.azlyrics.com",
		Path:   "search.php",
	}

	query := url.Values{}
	query.Add("q", fmt.Sprintf("%s %s", song.Title, strings.Join(song.Artists, " ")))

	searchUrl.RawQuery = query.Encode()

	resp, err := http.Get(searchUrl.String())
	if err != nil {
		return nil, fmt.Errorf("Could not complete request to AZ lyrics: %w", err)
	}

	body, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Could not parse AZ search html: %w", err)
	}

	songUrl, err := l.findSongUrl(body)
	if err != nil {
		return nil, fmt.Errorf("Could not find song url on AZ lyrics: %w", err)
	}

	fmt.Printf("song url: %s\n", songUrl)

	return nil, fmt.Errorf("GetLyrics not implemented")
}
