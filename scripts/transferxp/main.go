package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/jackc/pgx/v4"
)

type x struct {
	UserID uint64
	XP     int64
}

func main() {
	ctx := context.Background()

	gid, err := strconv.ParseUint(os.Getenv("GUILD"), 10, 64)
	if err != nil {
		log.Fatal("no or invalid guild ID provided ($GUILD environment variable)")
	}

	b, err := os.ReadFile("levels.json")
	if err != nil {
		log.Fatal("read file:", err)
	}

	var levels []x
	err = json.Unmarshal(b, &levels)
	if err != nil {
		log.Fatal("unmarshal:", err)
	}

	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE"))
	if err != nil {
		log.Fatal("connecting to db:", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatal("beginning transaction:", err)
	}

	// yeet all existing xp
	_, err = tx.Exec(ctx, "delete from levels where guild_id = $1", gid)
	if err != nil {
		log.Print("error deleting levels:", err)
		err = tx.Rollback(ctx)
		if err != nil {
			log.Fatal("rollback error:", err)
		}
	}

	for _, level := range levels {
		_, err = tx.Exec(ctx, "insert into levels (guild_id, user_id, xp) values ($1, $2, $3)", gid, level.UserID, level.XP)
		if err != nil {
			log.Printf("error inserting levels for %v: %v", level.UserID, err)
			err = tx.Rollback(ctx)
			if err != nil {
				log.Fatal("rollback error:", err)
			}
			return
		}
		log.Printf("inserted levels for %v (%v xp)", level.UserID, level.XP)
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Fatal("commit error:", err)
	}

	log.Printf("success! imported %v levels", len(levels))

	err = conn.Close(ctx)
	if err != nil {
		log.Print("close db error:", err)
	}
}
