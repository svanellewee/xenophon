package storage

import (
	"time"
)

type FilterType func(i int, entry *Entry) bool

type ResultStreamer interface {
	LastEntries(n int) ResultStreamer
	Period(start time.Time, end time.Time) ResultStreamer
	Location(location string) ResultStreamer
	Filter(filter FilterType) ResultStreamer
	Output() []*Entry
}

// type inMemoryStreamer struct {
// 	entries []*Entry
// }

// func (d *inMemoryStreamer) Output() []*Entry {
// 	return d.entries
// }

// // Filter implements ResultStreamer
// func (d *inMemoryStreamer) Filter(flr filterType) ResultStreamer {
// 	return filter(d.entries, flr)
// }

// // LastEntries implements ResultStreamer
// func (d *inMemoryStreamer) LastEntries(n int) ResultStreamer {
// 	results := make([]*Entry, 0, DefaultCapacity)
// 	var index int
// 	if len(d.entries) > n {
// 		index = len(d.entries) - n
// 	}
// 	results = append(results, d.entries[index:]...)
// 	return &inMemoryStreamer{
// 		entries: results,
// 	}
// }

// // Location implements ResultStreamer
// func (d *inMemoryStreamer) Location(location string) ResultStreamer {
// 	return filter(d.entries, func(index int, entry *Entry) bool {
// 		return string(entry.Location) == location
// 	})
// }

// // func filter(entries []*Entry, fltr filterType) *inMemoryStreamer {
// // 	results := make([]*Entry, 0, DefaultCapacity)
// // 	for i, entry := range entries {
// // 		if fltr(i, entry) {
// // 			results = append(results, entry)
// // 		}
// // 	}
// // 	return &inMemoryStreamer{
// // 		entries: results,
// // 	}
// // }

// // Period implements ResultStreamer
// func (d *inMemoryStreamer) Period(start time.Time, end time.Time) ResultStreamer {
// 	return filter(d.entries, func(i int, entry *Entry) bool {
// 		return entry.Time.Before(end) && entry.Time.After(start)
// 	})
// }

// func NewInMemoryStreamer() ResultStreamer {
// 	return &inMemoryStreamer{
// 		entries: make([]*Entry, 0, DefaultCapacity),
// 	}
// }

// func TestStreamIdea(t *testing.T) {
// 	s := NewInMemoryStreamer()

// 	start := time.Now()
// 	end := time.Now().Add(time.Minute)
// 	s.Location("/").
// 		LastEntries(10).
// 		Period(start, end).
// 		Filter(func(i int, e *Entry) bool { return true }).
// 		Filter(func(i int, e *Entry) bool { return false })
// }
