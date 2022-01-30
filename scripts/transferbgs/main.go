package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v4"
)

type x struct {
	UserID     uint64
	Background string
}

func main() {
	ctx := context.Background()

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

	for _, level := range levels {
		if strings.EqualFold(level.Background, "random") {
			continue
		}

		_, err = tx.Exec(ctx, "update levels set background = (select id from backgrounds where name ilike $1) where user_id = $2", level.Background, level.UserID)
		if err != nil {
			log.Printf("error updating background for %v: %v", level.UserID, err)
			err = tx.Rollback(ctx)
			if err != nil {
				log.Fatal("rollback error:", err)
			}
			return
		}
		log.Printf("updated background for %v (%v)", level.UserID, level.Background)
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
