package memory

import (
	"time"

	"github.com/svanellewee/xenophon/storage"
)

// Testing Mock dependencies...

func NewMemoryStore() storage.StorageStreamer {
	return &memoryStore{
		entries: make([]*storage.Entry, 0, storage.DefaultCapacity),
	}
}

type memoryStore struct {
	entries []*storage.Entry
}

func (d *memoryStore) Output() []*storage.Entry {
	return d.entries
}

// Filter implements storage.ResultStreamer
func (d *memoryStore) Filter(flr storage.FilterType) storage.ResultStreamer {
	return filter(d.entries, flr)
}

// LastEntries implements storage.ResultStreamer
func (d *memoryStore) LastEntries(n int) storage.ResultStreamer {
	results := make([]*storage.Entry, 0, storage.DefaultCapacity)
	var index int
	if len(d.entries) > n {
		index = len(d.entries) - n
	}
	results = append(results, d.entries[index:]...)
	return &memoryStore{
		entries: results,
	}
}

// Location implements storage.ResultStreamer
func (d *memoryStore) Location(location string) storage.ResultStreamer {
	return filter(d.entries, func(index int, entry *storage.Entry) bool {
		return string(entry.Location) == location
	})
}

func filter(entries []*storage.Entry, fltr storage.FilterType) *memoryStore {
	results := make([]*storage.Entry, 0, storage.DefaultCapacity)
	for i, entry := range entries {
		if fltr(i, entry) {
			results = append(results, entry)
		}
	}
	return &memoryStore{
		entries: results,
	}
}

// Period implements storage.ResultStreamer
func (d *memoryStore) Period(start time.Time, end time.Time) storage.ResultStreamer {
	return filter(d.entries, func(i int, entry *storage.Entry) bool {
		return entry.Time.Before(end) && entry.Time.After(start)
	})
}

func (m *memoryStore) Add(e *storage.Entry) (*storage.Entry, error) {
	index := len(m.entries) + 1
	e.Id = int64(index)
	t := time.Now()
	e.Time = &t
	m.entries = append(m.entries, e)
	return e, nil
}

func (m *memoryStore) Close() error {
	return nil
}
