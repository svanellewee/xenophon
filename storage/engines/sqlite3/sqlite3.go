package sqlite3

import (
	"database/sql"
	"time"

	storage "github.com/svanellewee/xenophon/storage"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteStorage struct {
	db      *sql.DB
	entries []*storage.Entry
	err     error
}

// Filter implements storage.StorageStreamer
func (sq *sqliteStorage) Filter(filter storage.FilterType) storage.ResultStreamer {
	var err error
	results := make([]*storage.Entry, 0, storage.DefaultCapacity)
	for i, entry := range sq.entries {
		if filter(i, entry) {
			results = append(results, entry)
		}
	}
	return &sqliteStorage{
		db:      sq.db,
		entries: results,
		err:     err,
	}
}

// Output implements storage.StorageStreamer
func (db *sqliteStorage) Output() []*storage.Entry {
	return db.entries
}

// Add implements StorageEngine
func (s *sqliteStorage) Add(e *storage.Entry) (*storage.Entry, error) {
	insertQuery := `
	INSERT INTO entry(entry_command, entry_location)
	VALUES (?, ?)
	`
	r, err := s.db.Exec(insertQuery, e.Command, e.Location)
	if err != nil {
		return nil, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return nil, err
	}

	query := `
	SELECT entry_id, entry_command, entry_location, entry_time
	FROM entry WHERE entry_id = ?
	`
	row := s.db.QueryRow(query, id)

	entryResult := &storage.Entry{}
	err = row.Scan(&entryResult.Id, &entryResult.Command, &entryResult.Location, &entryResult.Time)
	if err != nil {
		return nil, err
	}
	return entryResult, nil
}

func resultsFromRows(db *sql.DB, rows *sql.Rows, err error) storage.ResultStreamer {
	results := make([]*storage.Entry, 0, storage.DefaultCapacity)
	for rows.Next() {
		e := &storage.Entry{}
		if err = rows.Scan(&e.Id, &e.Command, &e.Location, &e.Time); err != nil {
			return nil
		}
		results = append(results, e)
	}
	if err = rows.Err(); err != nil {
		return nil
	}
	return &sqliteStorage{
		db:      db,
		entries: results,
		err:     err,
	}
}

// ForTime implements StorageEngine
func (s *sqliteStorage) Period(start time.Time, end time.Time) storage.ResultStreamer {
	query := `
	SELECT entry_id, entry_command, entry_location, entry_time
	FROM entry
	WHERE entry_time >= ? AND entry_time <= ?
	ORDER BY entry_time ASC
	`
	rows, err := s.db.Query(query, start.UnixMilli()/1000, end.UnixMilli()/1000)
	if err != nil {
		return nil
	}
	defer rows.Close()
	return resultsFromRows(s.db, rows, err)
}

// ForLocation finds all entries of the specified Location
func (s *sqliteStorage) Location(location string) storage.ResultStreamer {
	query := `
	SELECT entry_id, entry_command, entry_location, entry_time
	FROM entry
	WHERE entry_location = ?
	ORDER BY entry_time ASC
	`
	rows, err := s.db.Query(query, location)
	if err != nil {
		return nil
	}
	defer rows.Close()
	return resultsFromRows(s.db, rows, err)
}

// LastN implements StorageEngine
func (s *sqliteStorage) LastEntries(n int) storage.ResultStreamer {
	query := `
	WITH bw_results AS (
		SELECT * 
		FROM entry
		ORDER BY entry_time DESC
		LIMIT ?
	) 
	SELECT entry_id, entry_command, entry_location, entry_time
	FROM bw_results ORDER BY bw_results.entry_id ASC;
	`
	rows, err := s.db.Query(query, n)
	if err != nil {
		return nil
	}
	defer rows.Close()
	return resultsFromRows(s.db, rows, err)
}

func (s *sqliteStorage) Close() error {
	return s.db.Close()
}

func NewSqliteStorage(fileLocation string) storage.StorageStreamer {
	db, err := sql.Open("sqlite3", fileLocation)
	if err != nil {
		panic(err)
	}

	creationStatement := `
	CREATE TABLE IF NOT EXISTS entry (
		entry_id INTEGER PRIMARY KEY AUTOINCREMENT,
		entry_command VARCHAR,
		entry_location VARCHAR,
		entry_time TIMESTAMP DEFAULT (strftime('%s','now'))
	);
	CREATE INDEX IF NOT EXISTS entry_location_index ON entry (entry_location);
	`
	_, err = db.Exec(creationStatement)
	return &sqliteStorage{
		db:      db,
		entries: make([]*storage.Entry, 0, storage.DefaultCapacity),
		err:     err,
	}
}
