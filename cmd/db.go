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
    );
	
	CREATE TABLE IF NOT EXISTS History (
		puuid TEXT NOT NULL,
		data BLOB,
		PRIMARY KEY(puuid)
	);
	`

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

func getAllPlayerPuuIDs() ([]string, error) {
	db, err := openDB()
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT puuid FROM Players")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var puuIDs []string

	for rows.Next() {
		var puu string
		err = rows.Scan(&puu)
		if err != nil {
			return nil, err
		}
		puuIDs = append(puuIDs, puu)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return puuIDs, nil
}
