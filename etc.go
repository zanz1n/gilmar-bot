package main

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/zanz1n/gilmar-bot/logger"
)

type Phrase struct {
	ID       string  `bson:"i"`
	AuthorID *string `bson:"a"`
	Content  string  `bson:"c"`
}

func SliceIncludes[T comparable](s []T, item T) bool {
	for _, v := range s {
		if v == item {
			return true
		}
	}
	return false
}

func GetSubCommand(opts []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.ApplicationCommandInteractionDataOption {
	for _, opt := range opts {
		if opt.Type == discordgo.ApplicationCommandOptionSubCommand {
			return opt
		}
	}

	return nil
}

func HasPerm(memberPerm int64, target int64) bool {
	return memberPerm&target == target
}

// Can return nil pointer, so check:
//
//	if v := GetOption(opts, "name"); v == nil {
//	    // Handle possibility
//	}
func GetOption(opts []*discordgo.ApplicationCommandInteractionDataOption, name string) *discordgo.ApplicationCommandInteractionDataOption {
	for _, opt := range opts {
		if opt.Name == name {
			return opt
		}
	}

	return nil
}

func BasicResponse(format string, args ...any) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf(format, args...),
		},
	}
}

func BasicResponseEdit(format string, args ...any) *discordgo.WebhookEdit {
	fmt := fmt.Sprintf(format, args...)
	return &discordgo.WebhookEdit{
		Content: &fmt,
	}
}

func randr(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func randp(p uint8) bool {
	r := randr(0, 100)

	if uint8(r) > p {
		return false
	} else {
		return true
	}
}

func SetStatus(s *discordgo.Session, status StatusType) {
	str, name := "online", "/help"

	if status == StatusTypeIdle {
		str, name = "online", "/help"
	} else if status == StatusTypeStarting {
		str, name = "dnd", "Iniciando ..."
	} else if status == StatusTypeStopping {
		str, name = "dnd", "Desligando ..."
	} else {
		logger.Error("Failed to parse StatusType enumeration. Invalid")
	}

	s.UpdateStatusComplex(discordgo.UpdateStatusData{
		IdleSince: nil,
		Status:    str,
		AFK:       false,
		Activities: []*discordgo.Activity{
			{
				Name: name,
				Type: discordgo.ActivityTypeGame,
			},
		},
	})
}

func nanoid(l int) string {
	id, err := gonanoid.New(l)
	if err != nil {
		logger.Fatal(err)
	}
	return id
}
