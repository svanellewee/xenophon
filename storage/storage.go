package storage

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type EntryCount int
type LocationPath string
type Environment []string // encrypted bytestring ?

type Entry struct {
	Id       int64
	Time     *time.Time
	Location LocationPath
	Command  string
	Env      Environment
}

type QueryOpts func() bool

type StorageEngine interface {
	Add(*Entry) (*Entry, error)
	LastN(n int) ([]*Entry, error)
	ForTime(start time.Time, end time.Time) ([]*Entry, error)
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

func NewStandardModule(s StorageEngine) *DatabaseModule {
	return NewDatabaseModule(s, &DefaultLocation{}, &DefaultEnvironment{})
}

func NewDatabaseModule(s StorageEngine, l LocationGetter, e EnvironmentGetter) *DatabaseModule {
	return &DatabaseModule{
		Storage:     s,
		Location:    l,
		Environment: e,
	}
}

var ErrNotFound = errors.New("could not find entry")
var ErrBadDataInsert = errors.New("insert had an error")

type ResultSet struct {
	entries []*Entry
}

func (r *ResultSet) ForEach(doer func(int, *Entry)) {
	for i, e := range r.entries {
		doer(i, e)
	}
}

const defaultCapacity = 10

// Filter is equivalent to the bash |-pipe.
func (r *ResultSet) Filter(filter func(index int, entry *Entry) bool) *ResultSet {
	results := make([]*Entry, 0, defaultCapacity)
	for i, entry := range r.entries {
		if filter(i, entry) {
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

	results := make([]*Entry, 0, defaultCapacity)
	for _, entry := range entries {
		results = append(results, entry)
	}

	return &ResultSet{
		entries: results,
	}, nil
}
