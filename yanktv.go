package yanktv

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var reTvShow = regexp.MustCompile(`(.*).(S\d\dE\d\d).*(HDTV|WEBRIP).*X264`)

type App struct {
	conf Conf
	db   *database
}

func New(c Conf) (*App, error) {
	db, err := openDB(c.Database)
	if err != nil {
		return nil, err
	}

	app := &App{
		conf: c,
		db:   db,
	}
	return app, nil
}

func (app *App) Log(msg string, args ...interface{}) {
	if app.conf.Verbose {
		log.Printf(msg+"\n", args...)
	}
}

func (app *App) UpdateShows() error {
	for _, s := range app.conf.Shows {
		app.Log("Updating %s", s)

		err := app.updateShow(s)
		if err != nil {
			app.Log("Failed to update %s: %s", s, err)
		}

		time.Sleep(time.Duration(app.conf.DownloadTimeout) * time.Second)
	}
	return nil
}

func (app *App) updateShow(show string) error {
	url := app.conf.searchUrl(show)
	doc, err := app.openDoc(url)
	if err != nil {
		return err
	}

	torrents := app.parseTorrents(doc)
	err = app.db.insertOrIgnoreTorrents(torrents)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) GetTorrentsFromLastMonth() ([]Torrent, error) {
	torrents, err := app.db.getTorrentsFromLastMonth()
	if err != nil {
		return []Torrent{}, err
	}
	return torrents, nil
}

func (app *App) openDoc(url string) (*goquery.Document, error) {
	//if *fDebug {
	//f, err := os.Open("tmp/dumps/" + url)
	//if err != nil {
	//return nil, err
	//}
	//return goquery.NewDocumentFromReader(f)
	//}
	return goquery.NewDocument(url)
}

func (app *App) parseTorrents(doc *goquery.Document) []Torrent {
	now := time.Now()
	var torrents []Torrent

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
		t := Torrent{
			Title:     title + " " + episode,
			MagnetUrl: magnet,
			Timestamp: now,
		}
		torrents = append(torrents, t)
	})
	return torrents
}
