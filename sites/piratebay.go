package sites

import (
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/lmas/yanktv"
)

func init() {
	s := &PirateBay{}
	yanktv.AddSite(s)
}

const piratebayUrlBase string = "https://thepiratebay.org/search/%SEARCH/0/99/0"

var reTvShow = regexp.MustCompile(`(.*).(S\d\dE\d\d).*(HDTV|WEBRIP).*X264`)

type PirateBay struct {
}

func (p *PirateBay) Url(search string) string {
	return strings.Replace(piratebayUrlBase, "%SEARCH", search, -1)
}

func (p *PirateBay) ParseTorrents(doc *goquery.Document) []yanktv.Torrent {
	now := time.Now()
	var torrents []yanktv.Torrent

	doc.Find("#searchResult tbody tr").Each(func(i int, s *goquery.Selection) {
		// Find and grab each torrents' title and make sure it looks like
		// a show episode.
		tmp := s.Find("td .detName a").Text()
		parts := reTvShow.FindStringSubmatch(strings.ToUpper(tmp))
		if len(parts) != 4 {
			// Doesn't look like an episode.
			return
		}

		title := strings.Title(strings.ToLower(strings.Replace(parts[1], ".", " ", -1)))
		episode := parts[2]

		// Finally grab the magnet link for it.
		var magnet string
		s.Find("td a").Each(func(i int, ss *goquery.Selection) {
			if ss.AttrOr("title", "") == "Download this torrent using magnet" {
				magnet = ss.AttrOr("href", "")
			}
		})

		t := yanktv.Torrent{
			Title:     title + " " + episode,
			MagnetUrl: magnet,
			Timestamp: now,
		}
		torrents = append(torrents, t)
	})
	return torrents
}
