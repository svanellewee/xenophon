package storage

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	//	"github.com/svanellewee/xenophon/storage/engines/sqlite3"
)

// Testing Mock dependencies...

func newMemoryStore() *memoryStore {
	return &memoryStore{
		entries: make([]*Entry, 0, DefaultCapacity),
	}
}

type memoryStore struct {
	entries []*Entry
}

func (d *memoryStore) Output() []*Entry {
	return d.entries
}

// Filter implements ResultStreamer
func (d *memoryStore) Filter(flr FilterType) ResultStreamer {
	return filter(d.entries, flr)
}

// LastEntries implements ResultStreamer
func (d *memoryStore) LastEntries(n int) ResultStreamer {
	results := make([]*Entry, 0, DefaultCapacity)
	var index int
	if len(d.entries) > n {
		index = len(d.entries) - n
	}
	results = append(results, d.entries[index:]...)
	return &memoryStore{
		entries: results,
	}
}

// Location implements ResultStreamer
func (d *memoryStore) Location(location string) ResultStreamer {
	return filter(d.entries, func(index int, entry *Entry) bool {
		return string(entry.Location) == location
	})
}

func filter(entries []*Entry, fltr FilterType) *memoryStore {
	results := make([]*Entry, 0, DefaultCapacity)
	for i, entry := range entries {
		if fltr(i, entry) {
			results = append(results, entry)
		}
	}
	return &memoryStore{
		entries: results,
	}
}

// Period implements ResultStreamer
func (d *memoryStore) Period(start time.Time, end time.Time) ResultStreamer {
	return filter(d.entries, func(i int, entry *Entry) bool {
		return entry.Time.Before(end) && entry.Time.After(start)
	})
}

func (m *memoryStore) Add(e *Entry) (*Entry, error) {
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

// func (m *memoryStore) LastN(n int) ([]*Entry, error) {
// 	var startIndex int
// 	if len(m.entries) >= n {
// 		startIndex = len(m.entries) - n
// 	}
// 	return m.entries[startIndex:], nil
// }

// func (m *memoryStore) ForTime(start time.Time, end time.Time) ([]*Entry, error) {
// 	results := make([]*Entry, 0, DefaultCapacity)
// 	for _, entry := range m.entries {
// 		if entry.Time.After(start) && entry.Time.Before(end) {
// 			results = append(results, entry)
// 		}
// 	}

// 	return results, nil
// }

// func (m *memoryStore) ForLocation(location string) ([]*Entry, error) {
// 	results := make([]*Entry, 0, DefaultCapacity)
// 	for _, entry := range m.entries {
// 		if string(entry.Location) == location {
// 			results = append(results, entry)
// 		}
// 	}
// 	return results, nil
// }

type location struct {
	where string
	err   error
}

func newTestLocation() *location {
	return &location{}
}

func (l *location) Set(where string, err error) {
	l.where = where
	l.err = err
}

func (l *location) Get() (LocationPath, error) {
	if l.err != nil {
		return "", l.err
	}
	return LocationPath(l.where), nil
}

type environment struct {
	env []string
	err error
}

func (e *environment) Get() (Environment, error) {
	return e.env, nil
}

func (e *environment) Set(env []string, err error) {
	e.env = env
	e.err = err
}

func newTestEnv() *environment {
	return &environment{}
}

type testCase struct {
	TestName    string
	Location    *location
	Environment *environment
	Command     string
	//LocationError error
	//EnvError error
}

var testCases = []testCase{
	{
		TestName:    "Make directory and change to it",
		Location:    &location{"/home", nil},
		Environment: &environment{[]string{"PATH=/bin:/usr/local/bin", "PWD=/home"}, nil},
		Command:     "mkdir hello;cd /home/hello",
	},
	{
		TestName:    "Go to some system directory",
		Location:    &location{"/home/hello", nil},
		Environment: &environment{[]string{"PATH=/bin:/usr/local/bin", "PWD=/home/hello"}, nil},
		Command:     "cd /usr/local",
	},
	{
		TestName:    "Echo a friendly message",
		Location:    &location{"/", nil},
		Environment: &environment{[]string{"PATH=/bin:/usr/local/bin", "PWD=/usr/local"}, nil},
		Command:     "echo \"hello world\"",
	},
}

func GrepLocationFilter(matchString string) FilterType {
	return func(i int, e *Entry) bool {
		matched, err := regexp.Match(matchString, []byte(e.Location))
		if err != nil {
			return false
		}
		return matched
	}
}

func GrepCommandFilter(matchString string) FilterType {
	return func(i int, e *Entry) bool {
		matched, err := regexp.Match(matchString, []byte(e.Command))
		if err != nil {
			return false
		}
		return matched
	}
}

func TestGrepPipe(t *testing.T) {

	store := newMemoryStore()
	var mod *DatabaseModule
	testCases := []struct {
		command  string
		location string
	}{
		{"cd /", "/tmp"},
		{"echo $PATH", "/"},
		{"cd /tmp/hello", "/"},
		{"cd /tmp/bla", "/tmp/hello"},
		{"for i in {1..3}; do echo \"$i\"; done", "/tmp/bla"},
		{"vim", "/tmp/bla"},
	}
	for _, testCase := range testCases {
		mod = NewStorageModule(store,
			SetLocationGetter(&location{testCase.location, nil}),
			SetEnvironmentGetter(&environment{[]string{}, nil}))

		mod.Insert(testCase.command)
	}

	r := mod.LastEntries(100)
	entries := r.Filter(GrepCommandFilter("tmp")).Output()
	assert.Equal(t, 2, len(entries))

	for _, entry := range entries {
		fmt.Println("filter on command", entry.Command, entry.Id)
	}

	entries = r.Filter(GrepLocationFilter("tmp")).Output()
	assert.Equal(t, 4, len(entries))

	for _, entry := range entries {
		fmt.Println("filter on location", entry.Location, entry.Id)
	}

	tmpLocationAndCommand := r.Filter(GrepLocationFilter("tmp")).Filter(GrepCommandFilter("for i")).Output()
	assert.Equal(t, 1, len(tmpLocationAndCommand))
	for _, e := range tmpLocationAndCommand {
		matches, err := regexp.Match("^for", []byte(e.Command))
		assert.Nil(t, err)
		assert.True(t, matches)
	}

}

func TestLastEntries(t *testing.T) {
	store := newMemoryStore()
	for i, test := range testCases {
		fmt.Println(">>>", i)
		mod := NewStorageModule(store,
			SetLocationGetter(test.Location),
			SetEnvironmentGetter(test.Environment))

		t.Run(test.TestName, func(t *testing.T) {
			mod.Insert(test.Command)
			l := mod.LastEntries(1).Output()
			assert.Equal(t, test.Command, l[0].Command)

			// This should only return the max available entries
			ll := mod.LastEntries(10).Output()
			assert.Equal(t, i+1, len(ll))
		})
	}

	t.Run("No test entries", func(t *testing.T) {
		noMod := NewStorageModule(newMemoryStore())
		e := noMod.LastEntries(10).Output()
		fmt.Println(">>>>>>>>>", e)
		assert.Equal(t, 0, len(e))
	})
}

func TestSomething(t *testing.T) {
	t.Run("some test", func(t *testing.T) {
		mod := NewStorageModule(
			newMemoryStore(),
			SetLocationGetter(newTestLocation()),
			SetEnvironmentGetter(newTestEnv()),
		)

		mod.Insert("cd /")
		mod.Insert("mkdir hello")
		mod.Insert("touch hello/bla.txt")

		lastTwo := mod.LastEntries(2).Output()
		assert.Equal(t, 2, len(lastTwo))

		for _, s := range lastTwo {
			fmt.Println(s.Id, s.Time, s.Command)
		}

		lastThree := mod.LastEntries(3).Output()
		assert.Equal(t, 3, len(lastThree))

		start := time.Now()
		for _, command := range []string{
			"cd hello",
			"mkdir world",
			"cd ..",
			"touch hello/world/blahblah",
		} {
			_, err := mod.Insert(command)
			assert.Nil(t, err)
		}
		end := time.Now()

		rr := mod.Period(start, end)
		for _, ee := range rr.Output() {
			fmt.Println(ee.Command)
		}
	})
}
