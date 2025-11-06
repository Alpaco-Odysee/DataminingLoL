package main

import (
	"database/sql"
	"encoding/json"
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

	CREATE TABLE IF NOT EXISTS FirstBloods (
    	champion TEXT PRIMARY KEY,
    	win INTEGER NOT NULL,
    	maxGames INTEGER NOT NULL
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

func addPlayerHistory(history GameHistory) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	data, err := json.Marshal(history.MatchhistoryIDs)
	if err != nil {
		return err
	}

	sql := "INSERT INTO History(puuid, data) VALUES(?, ?) ON CONFLICT (puuid) DO UPDATE SET data=excluded.data"

	_, err = db.Exec(sql, history.PuuID, data)
	if err != nil {
		return err
	}

	return nil
}

func getMatchHistoryFromPlayer() ([]GameHistory, error) {
	db, err := openDB()
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT puuid, data FROM History")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []GameHistory

	for rows.Next() {
		var puuid string
		var data []byte

		err = rows.Scan(&puuid, &data)
		if err != nil {
			return nil, err
		}

		var matchIDs []string
		err = json.Unmarshal(data, &matchIDs)
		if err != nil {
			return nil, err
		}

		histories = append(histories, GameHistory{
			PuuID:           puuid,
			MatchhistoryIDs: matchIDs,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return histories, nil
}

func saveFirstBloodsToDB(stats map[string]*FirstBlood) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	sql := `
	INSERT INTO FirstBloods (champion, win, maxGames)
	VALUES (?, ?, ?)
	ON CONFLICT(champion) DO UPDATE SET
		win = excluded.win,
		maxGames = excluded.maxGames;
	`

	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, fb := range stats {
		_, err := stmt.Exec(fb.Champion, fb.Win, fb.MaxGames)
		if err != nil {
			fmt.Printf("DB insert error for %s: %v\n", fb.Champion, err)
		}
	}

	return nil
}
