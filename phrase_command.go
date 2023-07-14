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

func PhraseCommand(pr *Repository[[]Phrase]) Command {
	return Command{
		Accepts: CommandAccept{Slash: true, Button: false},
		Data:    phraseCommandData,
		Handler: handlePhrase(pr),
	}
}
