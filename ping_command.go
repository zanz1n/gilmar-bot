package main

import "github.com/bwmarrin/discordgo"

var pingCommandData = &discordgo.ApplicationCommand{
	Name:        "ping",
	Description: "Responde com pong e mostr o ping do bot",
	// Description: "Replies with pong and shows the current bot ping",
}

func PingCommand() Command {
	return Command{
		Accepts: CommandAccept{Slash: true, Button: false},
		Data:    pingCommandData,
		Handler: handlePing(),
	}
}

func handlePing() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		return s.InteractionRespond(
			i.Interaction,
			BasicResponse(
				"🏓 **Pong!**\nPing do bot: %vms",
				s.HeartbeatLatency().Milliseconds(),
			),
		)
	}
}
