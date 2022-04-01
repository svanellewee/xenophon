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

type StorageStreamer interface {
	ResultStreamer
	StorageEngine
}

type StorageEngine interface {
	Add(*Entry) (*Entry, error)
	Close() error
}

type LocationGetter interface {
	Get() (LocationPath, error)
}

type EnvironmentGetter interface {
	Get() (Environment, error)
}

type DatabaseModule struct {
	Storage     StorageStreamer
	Locator     LocationGetter
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

const DefaultCapacity = 10

// Insert inserts a command, env data into the datastore and ensures timestamp,id is returned.
func (d *DatabaseModule) Insert(command string) (*Entry, error) {

	location, err := d.Locator.Get()
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

// LastEntries provides the last N entries
func (d *DatabaseModule) LastEntries(n int) ResultStreamer {
	return d.Storage.LastEntries(n)
}

func (d *DatabaseModule) Period(start time.Time, end time.Time) ResultStreamer {
	return d.Storage.Period(start, end)
}

func (d *DatabaseModule) Location(location string) ResultStreamer {
	return d.Storage.Location(location)
}
