package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/jackc/pgx/v4"
)

type x struct {
	InviteID   string
	InviteName string
}

func main() {
	ctx := context.Background()

	b, err := os.ReadFile("invites.json")
	if err != nil {
		log.Fatal("read file:", err)
	}

	var invites []x
	err = json.Unmarshal(b, &invites)
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

	for _, invite := range invites {
		_, err = tx.Exec(ctx, "insert into invites (code, name) values ($1, $2)", invite.InviteID, invite.InviteName)
		if err != nil {
			log.Printf("error inserting name for %v: %v", invite.InviteID, err)
			err = tx.Rollback(ctx)
			if err != nil {
				log.Fatal("rollback error:", err)
			}
			return
		}
		log.Printf("inserted name for %v", invite.InviteID)
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Fatal("commit error:", err)
	}

	log.Printf("success! imported %v invites", len(invites))

	err = conn.Close(ctx)
	if err != nil {
		log.Print("close db error:", err)
	}
}
