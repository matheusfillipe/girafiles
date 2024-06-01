package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

const DUP_ENTRY_ERROR = "UNIQUE constraint failed: files.filename"

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
        bucket TEXT,
        alias TEXT,
        origin TEXT NOT NULL,
        timestamp INTEGER NOT NULL,
        UNIQUE (filename, bucket, alias) ON CONFLICT REPLACE
    );
    CREATE INDEX IF NOT EXISTS files_origin ON files (origin);
    CREATE INDEX IF NOT EXISTS files_filename ON files (filename);
    CREATE INDEX IF NOT EXISTS files_timestamp ON files (timestamp);
    CREATE INDEX IF NOT EXISTS files_bucket ON files (bucket);
    CREATE INDEX IF NOT EXISTS files_alias ON files (alias);
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
        COUNT(CASE WHEN timestamp >= strftime('%s', DATETIME(), '-1 day') THEN 1 END) AS day,
        COUNT(CASE WHEN timestamp >= strftime('%s', DATETIME(), '-1 hour') THEN 1 END) AS hour,
        COUNT(CASE WHEN timestamp >= strftime('%s', DATETIME(), '-1 minute') THEN 1 END) AS minute
    FROM files
    WHERE origin = ?;
  `, origin)

	if err := row.Scan(&day, &hour, &minute); err != nil {
		log.Fatal(err)
	}
	return HitCounts{origin, minute, hour, day}
}

func (db *DBHelper) CheckRateLimit(ip string) error {
	settings := GetSettings()
	hits := db.getHitCounts(ip)

	if settings.IPMinRateLimit > 0 && hits.minute >= settings.IPMinRateLimit {
		return errors.New("Rate limit per minute exceeded")
	}
	if settings.IPHourRateLimit > 0 && hits.hour >= settings.IPHourRateLimit {
		return errors.New("Rate limit per hour exceeded")
	}
	if settings.IPDayRateLimit > 0 && hits.day >= settings.IPDayRateLimit {
		return errors.New("Rate limit per day exceeded")
	}
	return nil
}

// Modifies node adding shortname to it
func (db *DBHelper) insertNode(node *Node) error {
	result, err := db.Exec("INSERT INTO files (filename, origin, timestamp) VALUES (?, ?, ?)", node.name, node.ip, node.timestamp)
	if err != nil {
		return err
	}

	idx, err := result.LastInsertId()
	if err != nil {
		return err
	}
	node.shortname = IdxToString(idx) + node.extension
	return nil
}

func (db *DBHelper) insertAlias(bucket string, alias string, node *Node) error {
	result, err := db.Exec("INSERT INTO files (filename, origin, timestamp, bucket, alias) VALUES (?, ?, ?, ?, ?)", node.name, node.ip, node.timestamp, bucket, alias)
	if err != nil {
		return err
	}
	idx, err := result.LastInsertId()
	if err != nil {
		return err
	}
	node.shortname = IdxToString(idx) + node.extension
	return nil
}

func (db *DBHelper) checkShortName(name string) (string, error) {
	// remove the extension from the filename
	name = strings.TrimSuffix(name, filepath.Ext(name))

	var path string
	index, err := StringToIdx(name)
	if err != nil {
		return "", fmt.Errorf("Failed to short filename!")
	}
	row := db.QueryRow("SELECT filename FROM files WHERE id = ?", index)
	if err := row.Scan(&path); err != nil {
		return "", err
	}
	return path, nil
}

func (db *DBHelper) checkFilename(name string) (string, error) {
	var path string
	row := db.QueryRow("SELECT filename FROM files WHERE filename = ?", name)
	if err := row.Scan(&path); err != nil {
		return "", err
	}
	return path, nil
}

func (db *DBHelper) checkAlias(bucket string, alias string) (string, error) {
	var path string
	row := db.QueryRow("SELECT filename FROM files WHERE bucket = ? AND alias = ?", bucket, alias)
	if err := row.Scan(&path); err != nil {
		return "", err
	}
	return path, nil
}

func (db *DBHelper) deleteExpiredFiles() ([]string, error) {
	settings := GetSettings()
	if settings.FilePersistanceTime == 0 {
		slog.Debug("File persistance time is unlimited")
		return []string{}, nil
	}

	// Just debugging
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		query := fmt.Sprintf(`
      SELECT filename,
        timestamp <= strftime('%%s', DATETIME(), '-%d hour') AS expired,
        timestamp || '(db)' || ' <= ' || strftime('%%s', DATETIME(), '-%d hour')
      FROM files
    `,
			settings.FilePersistanceTime,
			settings.FilePersistanceTime,
		)
		rows, err := db.Query(query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var results []map[string]string
		for rows.Next() {
			var filename string
			var expired bool
			var reason string
			if err := rows.Scan(&filename, &expired, &reason); err != nil {
				return nil, err
			}
			results = append(results, map[string]string{
				"filename": filename,
				"expired":  fmt.Sprintf("%t", expired),
				"reason":   reason,
			})
		}
		slog.Debug(fmt.Sprintf("Files and their expiration status: %v", results))
	}

	query := fmt.Sprintf("SELECT filename FROM files WHERE timestamp <= strftime('%%s', DATETIME(), '-%d hour')", settings.FilePersistanceTime)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		files = append(files, filename)
	}

	query = fmt.Sprintf("DELETE FROM files WHERE timestamp <= strftime('%%s', DATETIME(), '-%d hour')", settings.FilePersistanceTime)
	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (db *DBHelper) deleteOldestFiles(n int) ([]string, error) {
	settings := GetSettings()
	if settings.StorePathSizeLimit == 0 {
		return []string{}, nil
	}

	rows, err := db.Query("SELECT filename FROM files ORDER BY timestamp ASC LIMIT ?", n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	_, err = db.Exec("DELETE FROM files WHERE id IN (SELECT id FROM files ORDER BY timestamp ASC LIMIT ?)", n)
	if err != nil {
		return nil, err
	}
	var files []string
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		files = append(files, filename)
	}

	return files, nil
}

// / Update the timestamp of a file setting it to the current time
//func (db *DBHelper) updateTimestamp(filename string, timestamp int64) error {
//	_, err := db.Exec("UPDATE files SET timestamp = ? WHERE filename = ?", timestamp, filename)
//	return err
//}
