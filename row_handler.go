package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func handleActionRow(
	fr *Repository[[]Phrase],
) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		if i.Member == nil {
			return nil
		}

		if !HasPerm(i.Member.Permissions, discordgo.PermissionAdministrator) {
			return s.InteractionRespond(i.Interaction, BasicEphemeralResponse(
				"Você não tem permissão para fazer isso, <@%s>",
				i.Member.User.ID,
			))
		}

		data := i.MessageComponentData()

		if !strings.HasPrefix(data.CustomID, "phrase-delete/") {
			return nil
		}

		id := data.CustomID[14:]

		if len(id) != 45 {
			return nil
		}

		userId := id[13:31]

		if userId != i.Member.User.ID {
			return s.InteractionRespond(i.Interaction, BasicEphemeralResponse(
				"Você não é o autor do comando, use '/custom remove'",
			))
		}

		phraseId := id[:12]

		wasExcluded := false
		wasFound := fr.Transaction(i.GuildID, func(t []Phrase) []Phrase {
			for k, v := range t {
				if v.ID == phraseId {
					wasExcluded = true
					return append(t[:k], t[k+1:]...)
				}
			}

			return t
		})

		if !wasFound {
			return s.InteractionRespond(i.Interaction, BasicEphemeralResponse(
				"O servidor não possui nenhuma frase no momento",
			))
		} else if !wasExcluded {
			return s.InteractionRespond(i.Interaction, BasicEphemeralResponse(
				"Essa frase já foi excluída por um outro usuário",
			))
		} else {
			return s.InteractionRespond(i.Interaction, BasicEphemeralResponse(
				"Frase exluída com sucesso",
			))
		}
	}
}
