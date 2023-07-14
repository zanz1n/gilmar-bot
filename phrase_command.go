package main

import (
	"fmt"
	"strconv"
	"time"

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
			return handlePhraseAdd(pr, s, i)
		} else if subCommand.Name == "remove" {
			return handlePhraseRemove(pr, s, i)
		}

		return nil
	}
}

func handlePhraseAdd(
	pr *Repository[[]Phrase],
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	data := i.ApplicationCommandData()

	subCommand := GetSubCommand(data.Options)

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

func handlePhraseRemove(
	pr *Repository[[]Phrase],
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	phrases, ok := pr.Get(i.GuildID)

	if !ok || len(phrases) == 0 {
		return s.InteractionRespond(i.Interaction,
			BasicResponse(
				"Não há nenhuma frase registrada nesse servidor, <@%s>",
				i.Member.User.ID,
			),
		)
	}

	text := ""

	rows := []discordgo.ActionsRow{}

	ri, ai := 0, 0

	for k, v := range phrases {
		if ri == 5 {
			ri = 0
			ai++
		}
		if ri == 0 {
			rows = append(rows, discordgo.ActionsRow{})
		}

		ks := strconv.Itoa(k)
		text += "**" + ks + "** - " + v.Content + "\n"

		timeStamp := time.Now().UnixMilli()

		rows[ai].Components = append(rows[ai].Components, &discordgo.Button{
			Style: discordgo.DangerButton,
			Label: ks,
			Emoji: discordgo.ComponentEmoji{Name: "✖️"},
			CustomID: "phrase-delete/" +
				v.ID + "/" +
				i.Member.User.ID + "/" +
				strconv.FormatInt(timeStamp, 10),
		})

		ri++
	}

	components := make([]discordgo.MessageComponent, len(rows))

	for i, v := range rows {
		components[i] = v
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{{
				Title:       "Frases",
				Description: text,
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "Requisitado por " + i.Member.User.Username,
					IconURL: i.Member.AvatarURL("128"),
				},
			}},
			Components: components,
		},
	})
}

func PhraseCommand(pr *Repository[[]Phrase]) Command {
	return Command{
		Accepts: CommandAccept{Slash: true, Button: false},
		Data:    phraseCommandData,
		Handler: handlePhrase(pr),
	}
}
