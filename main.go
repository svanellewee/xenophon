package main

import (
	"fmt"
	"log"
	"time"

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
	r, err := mod.LastN(2)
	if err != nil {
		log.Fatal(err)
	}
	r.ForEach(func(i int, elem *storage.Entry) {
		fmt.Println("...", i, elem)
	})

	r2, err := mod.ForTime(time.Unix(1648674065, 0), time.Unix(1648674678, 0))
	r2.ForEach(func(i int, elem *storage.Entry) {
		fmt.Println("...", i, elem)
	})
}
