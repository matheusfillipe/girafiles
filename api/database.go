package api

import (
	"database/sql"
	"errors"
	"log"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

func openDatabase() *sql.DB {
	settings := GetSettings()

	db, err := sql.Open("sqlite3", filepath.Join(settings.StorePath, "files.db"))
	if err != nil {
		log.Fatal(err)
	}
	return db
}

type DBHelper struct {
	*sql.DB
}

var dbLock = &sync.Mutex{}
var dbInstance *sql.DB

func GetDB() *DBHelper {
	if dbInstance == nil {
		dbLock.Lock()
		defer dbLock.Unlock()
		if dbInstance == nil {
			dbInstance = openDatabase()
		}
	}
	return &DBHelper{dbInstance}
}

func (db *DBHelper) createTable() {
	_, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS files (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        filename TEXT NOT NULL UNIQUE,
        origin TEXT NOT NULL,
        timestamp INTEGER NOT NULL
    );
    CREATE INDEX IF NOT EXISTS files_origin ON files (origin);
    CREATE INDEX IF NOT EXISTS files_filename ON files (filename);
    CREATE INDEX IF NOT EXISTS files_timestamp ON files (timestamp);
  `)
	if err != nil {
		log.Fatal(err)
	}
}

type HitCounts struct {
	origin string
	minute int
	hour   int
	day    int
}

func (db *DBHelper) getHitCounts(origin string) HitCounts {
	var minute int
	var hour int
	var day int

	row := db.QueryRow(`
    SELECT
        COUNT(CASE WHEN timestamp >= strftime('%s', 'now', '-1 day') THEN 1 END) AS day,
        COUNT(CASE WHEN timestamp >= strftime('%s', 'now', '-1 hour') THEN 1 END) AS hour,
        COUNT(CASE WHEN timestamp >= strftime('%s', 'now', '-1 minute') THEN 1 END) AS minute
    FROM files
    WHERE origin = ?;
  `, origin)

	if err := row.Scan(&day, &hour, &minute); err != nil {
		log.Fatal(err)
	}
	return HitCounts{origin, minute, hour, day}
}

func (db *DBHelper) insertNode(node *Node) error {
	settings := GetSettings()
	hits := db.getHitCounts(node.ip)

	if hits.minute >= settings.IPMinRateLimit {
		return errors.New("Rate limit per minute exceeded")
	}
	if hits.hour >= settings.IPHourRateLimit {
		return errors.New("Rate limit per hour exceeded")
	}
	if hits.day >= settings.IPDayRateLimit {
		return errors.New("Rate limit per day exceeded")
	}

	_, err := db.Exec("INSERT INTO files (filename, origin, timestamp) VALUES (?, ?, ?)", node.name, node.ip, node.timestamp)
	return err
}

func (db *DBHelper) checkFileName(name string) (error, string) {
	// TODO we can introduce short filenames
	var path string
	row := db.QueryRow("SELECT filename FROM files WHERE filename = ?", name)
	if err := row.Scan(&path); err != nil {
		return err, ""
	}
	return nil, path
}
