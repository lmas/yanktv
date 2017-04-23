package yanktv

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
	sqlSchema string = `
CREATE TABLE IF NOT EXISTS torrents (
	id INTEGER PRIMARY KEY,
	title TEXT UNIQUE,
	magneturl TEXT,
	timestamp DATETIME
);

CREATE INDEX IF NOT EXISTS idx_shows ON torrents(timestamp);
`
	sqlInsertOrIgnoreTorrents string = `
INSERT OR IGNORE INTO torrents (title, magneturl, timestamp) VALUES (
	?, ?, ?
);
`
	sqlGetTorrentsBeforeTime string = `
SELECT title, magneturl, timestamp FROM torrents 
WHERE timestamp > ?
ORDER BY timestamp, title ASC;
`
)

type Torrent struct {
	Title     string
	MagnetUrl string
	Timestamp time.Time
}

type database struct {
	*sqlx.DB

	conf DatabaseConf
}

func openDB(c DatabaseConf) (*database, error) {
	db, err := sqlx.Connect("sqlite3", c.Path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(sqlSchema)
	if err != nil {
		return nil, err
	}

	return &database{db, c}, nil
}

func (db *database) getTorrentsFromLastMonth() ([]Torrent, error) {
	var torrents []Torrent
	t := time.Now().Add(time.Duration(-24*30) * time.Hour) // 30 days

	err := db.Select(&torrents, sqlGetTorrentsBeforeTime, t)
	if err != nil {
		return []Torrent{}, err
	}

	return torrents, nil
}

func (db *database) insertOrIgnoreTorrents(torrents []Torrent) error {
	now := time.Now()
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, t := range torrents {
		_, err := tx.Exec(sqlInsertOrIgnoreTorrents, t.Title, t.MagnetUrl, now)
		if err != nil {
			tx.Rollback() // TODO: handle ignored error here?
			// TODO: return a cleaner error
			return fmt.Errorf("error from db.insertOrIgnoreTorrents(%s): %s", t.Title, err)
		}
	}

	return tx.Commit()
}
