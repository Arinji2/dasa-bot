package main

import (
	"time"

	"github.com/arinji2/dasa-bot/env"
	"github.com/arinji2/dasa-bot/pb"
)

func main() {
	e := env.SetupEnv()

	pbAdmin := pb.SetupPocketbase(e.PB)

	ticker := time.NewTicker(5 * time.Hour)
	defer ticker.Stop()

	// Refreshing Admin Token
	go func() {
		for range ticker.C {
			pbAdmin = pb.SetupPocketbase(e.PB)
		}
	}()
}
