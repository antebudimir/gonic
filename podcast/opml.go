package podcast

import (
	"encoding/xml"
	"fmt"
	"io"
)

// ParseOPML reads podcast feed URLs from an OPML document. it returns the
// xmlUrl of every outline that has one, in document order, including outlines
// nested in grouping elements.
func ParseOPML(r io.Reader) ([]string, error) {
	var doc opml
	if err := xml.NewDecoder(r).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode opml: %w", err)
	}
	return doc.Body.Outline.feedURLs(), nil
}

type opml struct {
	XMLName xml.Name `xml:"opml"`
	Body    struct {
		Outline opmlOutlines `xml:"outline"`
	} `xml:"body"`
}

type opmlOutlines []opmlOutline

type opmlOutline struct {
	Text     string       `xml:"text,attr"`
	XMLURL   string       `xml:"xmlUrl,attr"`
	Children opmlOutlines `xml:"outline"`
}

func (o opmlOutlines) feedURLs() []string {
	var urls []string
	for _, outline := range o {
		if outline.XMLURL != "" {
			urls = append(urls, outline.XMLURL)
		}
		urls = append(urls, outline.Children.feedURLs()...)
	}
	return urls
}
