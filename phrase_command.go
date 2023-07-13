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
			Description: "Adicione uma frase customizada que só funciona no seu servidor",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "phrase",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "A frase que deseja adicionar",
					Required:    true,
				},
			},
		},
	},
}

func handlePhrase(pr *Repository[[]Phrase]) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		if i.Member == nil {
			return fmt.Errorf(
				"member is nil, command '%s', interaction id '%s'",
				i.ApplicationCommandData().Name,
				i.ID,
			)
		}

		if !HasPerm(i.Member.Permissions, discordgo.PermissionAdministrator) {
			s.InteractionRespond(i.Interaction,
				BasicResponse(
					"Você não tem permissão para usar esse comando, <@%s>",
					i.Member.User.ID,
				),
			)
		}

		data := i.ApplicationCommandData()

		subCommand := GetSubCommand(data.Options)

		if subCommand == nil {
			return fmt.Errorf("no subcommands provided, command '%s'", data.Name)
		}

		if subCommand.Name == "add" {
			phraseOpt := GetOption(subCommand.Options, "phrase")

			if phraseOpt == nil {
				return fmt.Errorf("option 'add' not provided, command '%s'", data.Name)
			}

			phrase := phraseOpt.StringValue()

			id := nanoid(12)

			userId := i.Member.User.ID

			pr.NotOverwriteSet(i.GuildID, []Phrase{})

			r := false
			var err error = nil

			pr.Transaction(i.GuildID, func(t []Phrase) []Phrase {
				for _, v := range t {
					if v.Content == phrase {
						r = true
						err = s.InteractionRespond(i.Interaction,
							BasicResponse("Essa frase já havia sido adicionada"),
						)
						return t
					}
				}
				return append(t, Phrase{
					ID:       id,
					AuthorID: &userId,
					Content:  phrase,
				})
			})

			if r {
				return err
			}

			return s.InteractionRespond(i.Interaction, BasicResponse("Frase adicionada!"))
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
