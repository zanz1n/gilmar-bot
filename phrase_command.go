package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var phraseCommandData = &discordgo.ApplicationCommand{
	Name:        "custom",
	Description: "Frases customizadas do bot",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "add",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "Adicione uma frase customizada que só funciona nesse servidor",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "phrase",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "A frase que deseja adicionar",
					Required:    true,
				},
			},
		},
		{
			Name:        "remove",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "Remova uma frase customizada",
			Options:     []*discordgo.ApplicationCommandOption{},
		},
		{
			Name:        "remove-id",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "Remova uma frase customizada por seu id",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "id",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "O id da frase customizada",
					Required:    true,
				},
			},
		},
		{
			Name:        "list",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "Liste as frases customizadas do servidor",
			Options:     []*discordgo.ApplicationCommandOption{},
		},
	},
}

func handlePhrase(
	pr *Repository[[]Phrase],
) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		if i.Member == nil {
			return fmt.Errorf(
				"member is nil, command '%s', interaction id '%s'",
				i.ApplicationCommandData().Name,
				i.ID,
			)
		}

		data := i.ApplicationCommandData()

		subCommand := GetSubCommand(data.Options)

		if subCommand == nil {
			return fmt.Errorf("no subcommands provided, command '%s'", data.Name)
		}

		if subCommand.Name == "list" {
			return handlePhraseList(pr, s, i)
		}

		if !HasPerm(i.Member.Permissions, discordgo.PermissionAdministrator) {
			s.InteractionRespond(i.Interaction,
				BasicResponse(
					"Você não tem permissão para usar esse comando, <@%s>",
					i.Member.User.ID,
				),
			)
		}

		if subCommand.Name == "add" {
			return handlePhraseAdd(pr, s, i)
		} else if subCommand.Name == "remove" {
			return handlePhraseRemove(pr, s, i)
		} else if subCommand.Name == "remove-id" {
			return handlePhraseRemoveId(pr, s, i)
		}

		return nil
	}
}

func PhraseCommand(pr *Repository[[]Phrase]) Command {
	return Command{
		Accepts: CommandAccept{Slash: true, Button: false},
		Data:    phraseCommandData,
		Handler: handlePhrase(pr),
	}
}
