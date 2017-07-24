package yanktv

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const userAgent string = "yanktv/0.2"

type App struct {
	conf      Conf
	db        *database
	webClient *http.Client
}

func New(c Conf) (*App, error) {
	db, err := openDB(c.Database)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	app := &App{
		conf:      c,
		db:        db,
		webClient: client,
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
	var torrents []Torrent
	search := url.QueryEscape(strings.TrimSpace(show))

	for _, s := range AllSites() {
		url := s.Url(search)
		doc, err := app.openDoc(url)
		if err != nil {
			app.Log("Failed to open %s: %s", url, err)
			continue
		}
		new := s.ParseTorrents(show, doc)
		torrents = append(torrents, new...)
	}

	err := app.db.insertOrIgnoreTorrents(torrents)
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

func (app *App) openDoc(webURL string) (*goquery.Document, error) {
	u, err := url.Parse(webURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", userAgent)
	resp, err := app.webClient.Do(req)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
