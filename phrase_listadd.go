package main

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func handlePhraseList(
	pr *Repository[[]Phrase],
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	phrases, ok := pr.Get(i.GuildID)

	if !ok {
		return s.InteractionRespond(
			i.Interaction,
			BasicResponse("O servidor não possui nenhuma frase no momento"),
		)
	}

	text := ""

	fields := make([]*discordgo.MessageEmbedField, len(phrases))

	for i, phrase := range phrases {
		is := strconv.Itoa(i)

		fields[i] = &discordgo.MessageEmbedField{
			Name:   is + " - id: '" + phrase.ID + "'",
			Value:  phrase.Content,
			Inline: false,
		}
	}

	return s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Type:        discordgo.EmbedTypeArticle,
						Title:       "Frases",
						Fields:      fields,
						Description: text,
						Footer: &discordgo.MessageEmbedFooter{
							Text:    "Requisitado por " + i.Member.User.Username,
							IconURL: i.Member.AvatarURL("128"),
						},
					},
				},
			},
		},
	)
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
