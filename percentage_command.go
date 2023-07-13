package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var percentageCommandData = &discordgo.ApplicationCommand{
	Name:        "percentage",
	Description: "Ajusta a porcentagem de chance de uma mensagem ser mandada",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "amount",
			Type:        discordgo.ApplicationCommandOptionInteger,
			Description: "Um número de 0 a 100",
			Required:    true,
		},
	},
}

func handlePercentage(r *Repository[uint8]) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		if i.Member == nil {
			return fmt.Errorf(
				"member is nil, command '%s', interaction id '%s'",
				i.ApplicationCommandData().Name,
				i.ID,
			)
		}

		if !HasPerm(i.Member.Permissions, discordgo.PermissionAdministrator) {
			return s.InteractionRespond(i.Interaction,
				BasicResponse(
					"Você não tem permissão para usar esse comando, <@%s>",
					i.Member.User.ID,
				),
			)
		}

		data := i.ApplicationCommandData()

		option := GetOption(data.Options, "amount")

		if option == nil {
			return fmt.Errorf("option 'amount' not provided, command '%s'", data.Name)
		}

		n := option.IntValue()

		if n > 100 || n < 0 {
			return s.InteractionRespond(i.Interaction,
				BasicResponse(
					"A porcentagem deve ser um número de 0 a 100, <@%s>",
					i.Member.User.ID,
				),
			)
		}

		nu8 := uint8(n)

		r.Set(i.GuildID, nu8)

		return s.InteractionRespond(i.Interaction,
			BasicResponse(
				"Chance alterada para %v por <@%s>",
				nu8,
				i.Member.User.ID,
			),
		)
	}
}

func PercentageCommand(r *Repository[uint8]) Command {
	return Command{
		Accepts: CommandAccept{Slash: true, Button: false},
		Data:    percentageCommandData,
		Handler: handlePercentage(r),
	}
}
