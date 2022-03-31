package storage

import "time"

type inMemoryStreamer struct {
	enties []*Entry
}

// Filter implements ResultStreamer
func (d *inMemoryStreamer) Filter(flr filterType) ResultStreamer {
	return filter(d.enties, flr)
}

// LastEntries implements ResultStreamer
func (d *inMemoryStreamer) LastEntries(n int) ResultStreamer {
	results := make([]*Entry, 0, DefaultCapacity)
	var index int
	if len(d.enties) > n {
		index = len(d.enties) - n
	}
	for _, e := range d.enties[index:] {
		results = append(results, e)
	}
	return &inMemoryStreamer{
		enties: results,
	}
}

// Location implements ResultStreamer
func (d *inMemoryStreamer) Location(location string) ResultStreamer {
	return filter(d.enties, func(index int, entry *Entry) bool {
		return string(entry.Location) == location
	})
}

func filter(entries []*Entry, fltr filterType) *inMemoryStreamer {
	results := make([]*Entry, 0, DefaultCapacity)
	for i, entry := range entries {
		if fltr(i, entry) {
			results = append(results, entry)
		}
	}
	return &inMemoryStreamer{
		enties: results,
	}

}

// Period implements ResultStreamer
func (d *inMemoryStreamer) Period(start time.Time, end time.Time) ResultStreamer {

	return filter(d.enties, func(i int, entry *Entry) bool {
		return entry.Time.Before(end) && entry.Time.After(start)
	})
}

func NewStreamer() ResultStreamer {
	return &inMemoryStreamer{
		enties: make([]*Entry, 0, DefaultCapacity),
	}
}
