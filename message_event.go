package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/gilmar-bot/logger"
)

func onMessage(
	fsr *Repository[[]Phrase],
	prr *Repository[uint8],
	defaultPhraseRepo *Repository[Phrase],
) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		gfss := defaultPhraseRepo.GetValues()
		if len(gfss) == 0 || m.Author.Bot {
			return
		}

		prr.NotOverwriteSet(m.GuildID, 30)

		percentage, ok := prr.Get(m.GuildID)

		_ = percentage

		if !ok {
			logger.Error(
				"Inconsistent write result caught in Repository[uint8] instance, guild '%s'",
				m.GuildID,
			)
			percentage = 30
		}

		if !randp(percentage) {
			return
		}

		var phrase Phrase

		phrases, ok := fsr.Get(m.GuildID)

		if !ok || len(phrases) == 0 {
			ridx := randr(1, len(gfss)) - 1

			phrase = gfss[ridx]
		} else {
			custom := randp(40)

			if custom {
				ridx := randr(1, len(phrases)) - 1

				phrase = phrases[ridx]
			} else {
				ridx := randr(1, len(gfss)) - 1

				phrase = gfss[ridx]
			}
		}

		content := phrase.Content

		if phrase.AuthorID != nil {
			content += "\n- <@" + *phrase.AuthorID + ">"
		}

		content = strings.ReplaceAll(content, "{USER}", "<@"+*phrase.AuthorID+">")

		_, err := s.ChannelMessageSendReply(m.ChannelID, content, m.Reference())

		if err != nil {
			logger.Error("Error caught while replying message, '%s'", err.Error())
		}
	}
}
