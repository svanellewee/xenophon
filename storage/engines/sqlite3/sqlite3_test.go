package sqlite3

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	storage "github.com/svanellewee/xenophon/storage"
)

func TestSqlite(t *testing.T) {
	sqliteDB := NewSqliteStorage(":memory:")

	defer sqliteDB.Close()

	mod := storage.NewStorageModule(sqliteDB)
	initialTestValues := []string{
		"cd /A",
		"cd /B",
		"cd /C",
	}

	for _, testCase := range initialTestValues {
		mod.Insert(testCase)
	}

	res := mod.LastEntries(100)
	assert.Equal(t, len(initialTestValues), len(res.Output()))

	for i, entry := range res.Output() {
		assert.Equal(t, initialTestValues[i], entry.Command)
	}

	timedTestCase := []string{
		"cd /1",
		"cd /2",
		"cd /3",
	}
	time.Sleep(time.Second)
	start := time.Now()
	for _, testCase := range timedTestCase {
		mod.Insert(testCase)
	}

	end := time.Now()
	time.Sleep(time.Second)

	testCommands := []string{
		"cd /4",
		"cd /5",
		"cd /6",
	}
	for _, testCommand := range testCommands {
		mod.Insert(testCommand)
	}

	r := mod.LastEntries(3)
	for i, entry := range r.Output() {
		fmt.Println("Entry,", entry)
		assert.Equal(t, testCommands[i], entry.Command)
	}

	r2 := mod.Period(start, end)
	for i, elem := range r2.Output() {
		assert.Equal(t, timedTestCase[i], elem.Command)
	}
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

func TestLocationFind(t *testing.T) {
	sqliteDB := NewSqliteStorage(":memory:")

	defer sqliteDB.Close()

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
		mod = storage.NewStorageModule(sqliteDB,
			storage.SetLocationGetter(&location{testCase.location, nil}),
			storage.SetEnvironmentGetter(&environment{[]string{}, nil}))

		mod.Insert(testCase.command)
	}

	slashLocationRes := mod.Location("/").Output()

	assert.Equal(t, 2, len(slashLocationRes))
	for _, e := range slashLocationRes {
		fmt.Println("Location found ->", e)
	}
}
