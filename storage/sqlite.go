package storage

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteStorage struct {
	db *sql.DB
	//	ctx context.Context
}

// Add implements StorageEngine
func (s *sqliteStorage) Add(e *Entry) (*Entry, error) {
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

	entryResult := &Entry{}
	err = row.Scan(&entryResult.Id, &entryResult.Command, &entryResult.Location, &entryResult.Time)
	if err != nil {
		return nil, err
	}
	return entryResult, nil
}

// ForTime implements StorageEngine
func (s *sqliteStorage) ForTime(start time.Time, end time.Time) ([]*Entry, error) {
	panic("unimplemented")
}

// LastN implements StorageEngine
func (s *sqliteStorage) LastN(n int) ([]*Entry, error) {
	query := `
	SELECT entry_id, entry_command, entry_location, entry_time
	FROM entry
	ORDER BY entry_time ASC
	LIMIT ?
	`
	rows, err := s.db.Query(query, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*Entry, 0, defaultCapacity)
	for rows.Next() {
		e := &Entry{}
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

func (s *sqliteStorage) Close() error {
	return s.db.Close()
}

func NewSqliteStorage(fileLocation string) (StorageEngine, error) {
	db, err := sql.Open("sqlite3", fileLocation)
	if err != nil {
		panic(err)
	}
	//defer db.Close()

	creationStatement := `
	CREATE TABLE IF NOT EXISTS entry (
		entry_id INTEGER PRIMARY KEY AUTOINCREMENT,
		entry_command VARCHAR,
		entry_location VARCHAR,
		entry_time TIMESTAMP DEFAULT (strftime('%s','now'))
	);
	`
	_, err = db.Exec(creationStatement)
	if err != nil {
		return nil, err
	}

	return &sqliteStorage{
		db,
	}, nil
}
