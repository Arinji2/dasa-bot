package main

import (
	"log"
	"time"

	"github.com/arinji2/dasa-bot/bot"
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

	discordBot, err := bot.NewBot(e.Bot)
	if err != nil {
		log.Panicf("Cannot create bot: %v", err)
	}
	discordBot.Run(pbAdmin)
}
