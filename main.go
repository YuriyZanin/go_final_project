package main

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	dbFile := getDbFile()
	install := !isExistDb(dbFile)

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	if install {
		installDb(db)
	}

	store := NewStore(db)
	service := NewService(store)
	handler := NewHandler(service)

	server := Server{}
	server.Run(getServerPort(), handler.InitRouter())
}

func isExistDb(dbPath string) bool {
	_, err := os.Stat(dbPath)
	return err == nil
}

func getDbFile() string {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "./scheduler.db"
	}
	return dbFile
}

func getServerPort() string {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}
	return port
}

func installDb(db *sql.DB) {
	log.Println("Создание таблицы")
	_, err := db.Exec("CREATE TABLE scheduler ( id integer PRIMARY KEY, date varchar(8), title varchar(64), comment varchar(128), repeat varchar(128))")
	if err != nil {
		panic(err)
	}

	log.Println("Создание индекса")
	_, err = db.Exec("CREATE INDEX idx_scheduler_date ON scheduler (date)")
	if err != nil {
		panic(err)
	}
}