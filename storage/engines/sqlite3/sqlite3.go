package sqlite3

import (
	"database/sql"
	"time"

	storage "github.com/svanellewee/xenophon/storage"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteStorage struct {
	db *sql.DB
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
	err = row.Scan()
	if err != nil {
		return nil, err
	}

	entryResult := &storage.Entry{}
	err = row.Scan(&entryResult.Id, &entryResult.Command, &entryResult.Location, &entryResult.Time)
	if err != nil {
		return nil, err
	}
	return entryResult, nil
}

func resultsFromRows(rows *sql.Rows) ([]*storage.Entry, error) {
	var err error
	results := make([]*storage.Entry, 0, storage.DefaultCapacity)
	for rows.Next() {
		e := &storage.Entry{}
		if err = rows.Scan(&e.Id, &e.Command, &e.Location, &e.Time); err != nil {
			return nil, err
		}
		results = append(results, e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// ForTime implements StorageEngine
func (s *sqliteStorage) ForTime(start time.Time, end time.Time) ([]*storage.Entry, error) {
	query := `
	SELECT entry_id, entry_command, entry_location, entry_time
	FROM entry
	WHERE entry_time >= ? AND entry_time <= ?
	ORDER BY entry_time ASC
	`
	rows, err := s.db.Query(query, start.UnixMilli()/1000, end.UnixMilli()/1000)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return resultsFromRows(rows)
}

// ForLocation finds all entries of the specified Location
func (s *sqliteStorage) ForLocation(location string) ([]*storage.Entry, error) {
	query := `
	SELECT entry_id, entry_command, entry_location, entry_time
	FROM entry
	WHERE entry_location = ?
	ORDER BY entry_time ASC
	`
	rows, err := s.db.Query(query, location)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return resultsFromRows(rows)
}

// LastN implements StorageEngine
func (s *sqliteStorage) LastN(n int) ([]*storage.Entry, error) {
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
		return nil, err
	}
	defer rows.Close()
	return resultsFromRows(rows)
}

func (s *sqliteStorage) Close() error {
	return s.db.Close()
}

func NewSqliteStorage(fileLocation string) (storage.StorageEngine, error) {
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
	if err != nil {
		return nil, err
	}

	return &sqliteStorage{
		db,
	}, nil
}
