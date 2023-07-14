package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

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

	if len(phrases) > 20 {
		return s.InteractionRespond(i.Interaction,
			BasicResponse("Há mais de 20 frases adicionadas ao servidor, use "+
				"'/custom list' e depois '/custom remove-id <id da frase>'"),
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

func handlePhraseRemoveId(
	pr *Repository[[]Phrase],
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	data := i.ApplicationCommandData()

	subCommand := GetSubCommand(data.Options)

	idOpt := GetOption(subCommand.Options, "id")

	if idOpt == nil {
		return fmt.Errorf("option 'amount' not provided, command '%s'", data.Name)
	}

	id := idOpt.StringValue()

	changed := false
	serverHasPhrases := pr.Transaction(i.GuildID, func(t []Phrase) []Phrase {
		for k, phrase := range t {
			if phrase.ID == id {
				changed = true
				return append(t[:k], t[k+1:]...)
			}
		}

		return t
	})

	if !serverHasPhrases {
		return s.InteractionRespond(i.Interaction, BasicEphemeralResponse(
			"O servidor não possui nenhuma frase no momento",
		))
	} else if !changed {
		return s.InteractionRespond(i.Interaction, BasicEphemeralResponse(
			"Essa frase já foi excluída por um outro usuário",
		))
	} else {
		return s.InteractionRespond(i.Interaction, BasicEphemeralResponse(
			"Frase exluída com sucesso",
		))
	}
}
