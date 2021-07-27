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

	return nil, fmt.Errorf("Could not find a url")
}

func findDiv(node *html.Node) *html.Node {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.DataAtom == atom.Div && len(child.Attr) == 0 {
			return child
		} else {
			possibleDiv := findDiv(child)
			if possibleDiv != nil {
				return possibleDiv
			}
		}
	}
	return nil
}

func extractText(node *html.Node) []string {
	lyrics := []string{}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			lyrics = append(lyrics, child.Data)
		} else {
			for _, lyric := range extractText(child) {
				lyrics = append(lyrics, lyric)
			}
		}
	}
	return lyrics
}

func (l *AZLyricProvider) findLyrics(doc *html.Node) ([]string, error) {
	// Look for a div with no class or id.
	// Then get all the text from within that div and return it as a list of strings
	div := findDiv(doc)

	if div == nil {
		return nil, fmt.Errorf("There is no div without any attributes")
	}

	return extractText(div), nil
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

	// log.Printf("Searching for %s using %s\n", song.Formatted(), searchUrl.String())
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
		return nil, fmt.Errorf("Could not find song url on AZ lyrics: %s", searchUrl.String())
	}

	// With the song url, now scrape the lyrics from the actual page
	resp, err = http.Get(songUrl.String())
	if err != nil {
		return nil, fmt.Errorf("Could not complete request to AZ lyrics: %w", err)
	}

	body, err = html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Could not parse AZ song html: %w", err)
	}

	lyrics, err := l.findLyrics(body)
	if err != nil {
		return nil, fmt.Errorf("Could not find lyrics in AZ song html: %w", err)
	}

	return &core.SongWithLyrics{Song: *song, Lyrics: lyrics}, nil
}
