package yanktv

import (
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Site interface {
	Url(string) string
	ParseTorrents(string, *goquery.Document) []Torrent
}

var (
	sites     []Site
	sitesLock sync.Mutex
)

func AddSite(s Site) {
	sitesLock.Lock()
	sites = append(sites, s)
	sitesLock.Unlock()
}

func AllSites() []Site {
	sitesLock.Lock()
	s := sites
	sitesLock.Unlock()
	return s
}
