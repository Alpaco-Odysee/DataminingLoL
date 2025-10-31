package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func openDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "db")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func createDB() {
	db, err := openDB()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer db.Close()

	schema := `
	CREATE TABLE IF NOT EXISTS Players (
        league_id TEXT NOT NULL,
        puuid TEXT NOT NULL
    );`

	db.Exec(schema)
}

func addPlayer(l LeagueEntry) error {
	db, err := openDB()
	if err != nil {
		fmt.Println("error:", err)
		return err
	}
	defer db.Close()

	sql := "INSERT INTO Players(league_id, puuid) VALUES(?, ?)"

	_, err = db.Exec(sql, l.LeagueID, l.PuuID)
	if err != nil {
		return err
	}

	return nil
}
