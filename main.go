package main

import (
	"fmt"
	"log"

	storage "github.com/svanellewee/xenophon/storage"
)

func main() {
	fmt.Println("Hiya")
	sqliteDB, err := storage.NewSqliteStorage("/tmp/hello.db")
	if err != nil {
		log.Fatal(err)
	}
	defer sqliteDB.Close()

	mod := storage.NewStandardModule(sqliteDB)
	mod.Insert("cd /")
	r, err := mod.LastN(10)
	if err != nil {
		log.Fatal(err)
	}
	r.ForEach(func(i int, elem *storage.Entry) {
		fmt.Println("...", i, elem)
	})
}