package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"
)

type EntryCount int
type LocationPath string
type Environment []string // encrypted bytestring ?
type filterType func(i int, entry *Entry) bool

type ResultStreamer interface {
	LastEntries(n int) ResultStreamer
	Period(start time.Time, end time.Time) ResultStreamer
	Location(location string) ResultStreamer
	Filter(filter filterType) ResultStreamer
}

type StorageEngine interface {
	Add(*Entry) (*Entry, error)
	LastN(n int) ([]*Entry, error)
	ForTime(start time.Time, end time.Time) ([]*Entry, error)
	ForLocation(location string) ([]*Entry, error)
	Close() error
}

type LocationGetter interface {
	Get() (LocationPath, error)
}

type EnvironmentGetter interface {
	Get() (Environment, error)
}

type DatabaseModule struct {
	Storage     StorageEngine
	Location    LocationGetter
	Environment EnvironmentGetter
	// TimeGetter?
}

type DefaultEnvironment struct{}

func (*DefaultEnvironment) Get() (Environment, error) {
	return os.Environ(), nil
}

type DefaultLocation struct{}

func (*DefaultLocation) Get() (LocationPath, error) {
	location, err := os.Getwd()
	if err != nil {
		return "", nil
	}
	return LocationPath(location), nil
}

var ErrNotFound = errors.New("could not find entry")
var ErrBadDataInsert = errors.New("insert had an error")

type ResultSet struct {
	entries []*Entry
	db      *sql.DB
}

func (r *ResultSet) ForEach(doer func(int, *Entry)) {
	for i, e := range r.entries {
		doer(i, e)
	}
}

func (r *ResultSet) Count() int {
	return len(r.entries)
}

func (r *ResultSet) Entries() ([]Entry, error) {
	returnEntries := make([]Entry, 0, DefaultCapacity)
	for _, r := range r.entries {
		e := Entry{}
		r.Copy(&e)
		returnEntries = append(returnEntries, e)
	}
	return returnEntries, nil
}

const DefaultCapacity = 10

type FilterFunc func(index int, entry *Entry) error
type ErrMatchNotFound struct {
	searchString string
	err          error
}

func (e *ErrMatchNotFound) Error() string {
	return fmt.Sprintf("could not find %s", e.searchString)
}

func GrepCommandFilter(matchString string) FilterFunc {
	return func(index int, entry *Entry) error {
		matched, err := regexp.Match(matchString, []byte(entry.Command))
		if err != nil {
			return fmt.Errorf("could not find due to error %w", err)
		}

		if !matched {
			return &ErrMatchNotFound{}
		}
		return nil
	}
}

func GrepLocationFilter(matchString string) FilterFunc {
	return func(index int, entry *Entry) error {
		matched, err := regexp.Match(matchString, []byte(entry.Location))
		if err != nil {
			return fmt.Errorf("could not find due to error %w", err)
		}

		if !matched {
			return &ErrMatchNotFound{}
		}
		return nil
	}
}

// Filter is equivalent to the bash |-pipe.
func (r *ResultSet) Filter(filter FilterFunc) *ResultSet {
	results := make([]*Entry, 0, DefaultCapacity)
	for i, entry := range r.entries {
		if err := filter(i, entry); err == nil {
			results = append(results, entry)
		}
	}
	if len(results) == 0 {
		return nil
	}
	resultSet := &ResultSet{
		entries: results,
	}
	return resultSet
}

// Insert inserts a command, env data into the datastore and ensures timestamp,id is returned.
func (d *DatabaseModule) Insert(command string) (*Entry, error) {

	location, err := d.Location.Get()
	if err != nil {
		return nil, fmt.Errorf("location could not be determined: %w", err)
	}

	environment, err := d.Environment.Get()
	if err != nil {
		return nil, fmt.Errorf("environment could not be determined: %w", err)
	}

	e, err := d.Storage.Add(&Entry{
		Location: location,
		Command:  command,
		Env:      environment,
	})

	if err != nil {
		return nil, err
	}

	if e == nil {
		return nil, ErrNotFound
	}

	if e.Time == nil {
		return nil, ErrBadDataInsert
	}

	if e.Id <= 0 {
		return nil, ErrBadDataInsert
	}
	return e, nil
}

// LastN provides the last N entries
func (d *DatabaseModule) LastN(n int) (*ResultSet, error) {
	entries, err := d.Storage.LastN(n)
	if err != nil {
		return nil, err
	}

	results := make([]*Entry, 0, n)
	for _, entry := range entries {
		results = append(results, entry)
	}

	return &ResultSet{
		entries: results,
	}, nil
}

// ForTime provides the entries over a certain time range
func (d *DatabaseModule) ForTime(start time.Time, end time.Time) (*ResultSet, error) {
	entries, err := d.Storage.ForTime(start, end)
	if err != nil {
		return nil, err
	}

	results := make([]*Entry, 0, DefaultCapacity)
	for _, entry := range entries {
		results = append(results, entry)
	}

	return &ResultSet{
		entries: results,
	}, nil
}

func (d *DatabaseModule) ForLocation(locationRegex string) (*ResultSet, error) {
	entries, err := d.Storage.ForLocation(locationRegex)
	if err != nil {
		return nil, err
	}

	results := make([]*Entry, 0, DefaultCapacity)
	for _, entry := range entries {
		results = append(results, entry)
	}

	return &ResultSet{
		entries: results,
	}, nil
}
