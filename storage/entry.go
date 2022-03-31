package storage

import "time"

type Entry struct {
	Id       int64
	Time     *time.Time
	Location LocationPath
	Command  string
	Env      Environment
}

func (source *Entry) Copy(dest *Entry) {
	dest.Location = source.Location
	dest.Command = source.Command
	dest.Env = source.Env
}
