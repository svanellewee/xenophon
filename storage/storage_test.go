package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Testing Mock dependencies...

func newMemoryStore() *memoryStore {
	return &memoryStore{
		entries: make([]*Entry, 0, defaultCapacity),
	}
}

type memoryStore struct {
	entries []*Entry
}

func (m *memoryStore) Add(e *Entry) (*Entry, error) {
	index := len(m.entries) + 1
	e.Id = int64(index)
	t := time.Now()
	e.Time = &t
	m.entries = append(m.entries, e)
	return e, nil
}

func (m *memoryStore) LastN(n int) ([]*Entry, error) {
	var startIndex int
	if len(m.entries) >= n {
		startIndex = len(m.entries) - n
	}
	return m.entries[startIndex:], nil
}

func (m *memoryStore) ForTime(start time.Time, end time.Time) ([]*Entry, error) {
	results := make([]*Entry, 0, defaultCapacity)
	for _, entry := range m.entries {
		if entry.Time.After(start) && entry.Time.Before(end) {
			results = append(results, entry)
		}
	}

	return results, nil
}

func (m *memoryStore) Close() error {
	return nil
}

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
	return e.env, nil //make([]string, 0, defaultCapacity), nil
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

func TestGrepPipe(t *testing.T) {

}

func TestSomethingElse(t *testing.T) {
	store := newMemoryStore()
	for i, test := range testCases {
		mod := NewDatabaseModule(store, test.Location, test.Environment)
		t.Run(test.TestName, func(t *testing.T) {
			//location.Set(test.Location, test.LocationError)
			mod.Insert(test.Command)
			lastEntry, err := mod.LastN(1)
			assert.Nil(t, err)
			assert.Equal(t, test.Command, lastEntry.entries[0].Command)
			assert.Equal(t, len(store.entries), i+1)

			// This should only return the max
			rr, err := mod.LastN(10)
			assert.Nil(t, err)
			assert.Equal(t, i+1, len(rr.entries))
		})
	}
}

func TestSomething(t *testing.T) {
	t.Run("some test", func(t *testing.T) {
		mod := NewDatabaseModule(newMemoryStore(), newTestLocation(), newTestEnv())

		mod.Insert("cd /")
		mod.Insert("mkdir hello")
		mod.Insert("touch hello/bla.txt")

		lastTwo, err := mod.LastN(2)
		assert.Nil(t, err)

		assert.Equal(t, 2, len(lastTwo.entries))
		for _, s := range lastTwo.entries {
			fmt.Println(s.Id, s.Time, s.Command)
		}

		lastThree, err := mod.LastN(3)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(lastThree.entries))

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

		rr, err := mod.ForTime(start, end)
		assert.Nil(t, err)
		for _, ee := range rr.entries {
			fmt.Println(ee.Command)
		}
	})
}
