package feeder

import (
	"log"

	"github.com/canthefason/irc-k/client"
	"github.com/canthefason/irc-k/config"
)

func run() {
	conn := new(client.Connection)
	if err := conn.Connect(config.Conf.IRC.BotName); err != nil {
		log.Fatal("Error occurred in initialization: %s", err)
	}
}

// func main() {
// 	conn := new(client.Connection)
// 	// TODO config file
// 	if err := client.Connect("koding-bot"); err != nil {
// 		log.Fatal("Error occurred in initialization: %s", err)
// 	}

// 	http.HandleFunc("/join", joinChannel)
// }

// four different sets
// 1. user - subscribed channels
// 2. server - not connected channels
// 3. server - connected channels
// 4. server - channel messages
func initChannels() {
	for {

	}
}
