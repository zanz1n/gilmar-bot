package main

import "github.com/bwmarrin/discordgo"

var helpCommandData = &discordgo.ApplicationCommand{
	Name:        "help",
	Description: "Mostra os comandos do bot",
}

const helpEmbedDescription = "Isso aí machão!"

func cmdsIntoFields(cmds []*discordgo.ApplicationCommand) []*discordgo.MessageEmbedField {
	fields := []*discordgo.MessageEmbedField{}

	for _, cmd := range cmds {
		isRoot := true

		for _, opt := range cmd.Options {
			if opt.Type == discordgo.ApplicationCommandOptionSubCommand {
				isRoot = false

				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   cmd.Name + " " + opt.Name,
					Value:  opt.Description,
					Inline: true,
				})
			} else if opt.Type == discordgo.ApplicationCommandOptionSubCommandGroup {
				isRoot = false

				for _, gOpt := range opt.Options {
					fields = append(fields, &discordgo.MessageEmbedField{
						Name:   cmd.Name + " " + opt.Name + " " + gOpt.Name,
						Value:  gOpt.Description,
						Inline: true,
					})
				}
			}
		}

		if isRoot {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   cmd.Name,
				Value:  cmd.Description,
				Inline: true,
			})
		}
	}

	return fields
}

func handleHelp(cm *CommandHandler) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		cmds := cm.GetData(CommandAccept{Slash: true, Button: false})

		fields := cmdsIntoFields(cmds)

		embed := discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeArticle,
			Title:       "Comandos",
			Description: helpEmbedDescription,
			Fields:      fields,
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "Requisitado por " + i.Member.User.Username,
				IconURL: i.Member.AvatarURL("128"),
			},
		}

		row := discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					Label: "Github",
					Style: discordgo.LinkButton,
					Emoji: discordgo.ComponentEmoji{Name: "🔗"},
					URL:   "https://github.com/zanz1n/gilmar-bot",
				},
			},
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{&embed},
				Components: []discordgo.MessageComponent{&row},
			},
		})
	}
}

func HelpCommand(cm *CommandHandler) Command {
	return Command{
		Accepts: CommandAccept{Slash: true, Button: false},
		Data:    helpCommandData,
		Handler: handleHelp(cm),
	}
}
