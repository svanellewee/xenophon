package memory

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/svanellewee/xenophon/storage"
)

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

func (l *location) Get() (storage.LocationPath, error) {
	if l.err != nil {
		return "", l.err
	}
	return storage.LocationPath(l.where), nil
}

type environment struct {
	env []string
	err error
}

func (e *environment) Get() (storage.Environment, error) {
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

func GrepLocationFilter(matchString string) storage.FilterType {
	return func(i int, e *storage.Entry) bool {
		matched, err := regexp.Match(matchString, []byte(e.Location))
		if err != nil {
			return false
		}
		return matched
	}
}

func GrepCommandFilter(matchString string) storage.FilterType {
	return func(i int, e *storage.Entry) bool {
		matched, err := regexp.Match(matchString, []byte(e.Command))
		if err != nil {
			return false
		}
		return matched
	}
}

func TestGrepPipe(t *testing.T) {

	store := NewMemoryStore()
	var mod *storage.DatabaseModule
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
		mod = storage.NewStorageModule(store,
			storage.SetLocationGetter(&location{testCase.location, nil}),
			storage.SetEnvironmentGetter(&environment{[]string{}, nil}))

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
	store := NewMemoryStore()
	for i, test := range testCases {
		fmt.Println(">>>", i)
		mod := storage.NewStorageModule(store,
			storage.SetLocationGetter(test.Location),
			storage.SetEnvironmentGetter(test.Environment))

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
		noMod := storage.NewStorageModule(NewMemoryStore())
		e := noMod.LastEntries(10).Output()
		fmt.Println(">>>>>>>>>", e)
		assert.Equal(t, 0, len(e))
	})
}

func TestSomething(t *testing.T) {
	t.Run("some test", func(t *testing.T) {
		mod := storage.NewStorageModule(
			NewMemoryStore(),
			storage.SetLocationGetter(newTestLocation()),
			storage.SetEnvironmentGetter(newTestEnv()),
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
