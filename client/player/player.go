package player

import (
	"discord-sound/player/requests"
	"discord-sound/player/server"
	"discord-sound/utils/discord"
	"discord-sound/utils/kafka"
	"discord-sound/utils/redis"
	"discord-sound/utils/uuid"
	"fmt"
	"os"
)

// Start player
func Start() {
	kafkaURL := os.Getenv("KAFKA_URL")
	err := kafka.Init(kafkaURL, "player-"+uuid.Gen())

	if err != nil {
		panic(err)
	}

	redisURL := os.Getenv("REDIS_URL")
	err = redis.Init(redisURL)

	if err != nil {
		panic(err)
	}

	defer kafka.Close()

	discordToken := os.Getenv("DISCORD_TOKEN")
	err = discord.Init(discordToken)

	if err != nil {
		panic(err)
	}

	requests.StartDone()

	discord.Client.Client.AddHandler(server.HandleMessage)
	err = discord.Open()

	if err != nil {
		panic(err)
	}

	defer discord.Close()

	fmt.Println("Player Started")

	select {}

}
