package main

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/gilmar-bot/logger"
)

type CommandAccept struct {
	Slash  bool
	Button bool
}

type Command struct {
	Accepts CommandAccept
	Data    *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate) error
}

type CommandHandler struct {
	cmds   map[string]*Command
	cmdsMu sync.RWMutex
}

func NewCommandHandler() *CommandHandler {
	return &CommandHandler{
		cmds:   make(map[string]*Command),
		cmdsMu: sync.RWMutex{},
	}
}

func (ch *CommandHandler) Add(cmd Command) {
	ch.cmdsMu.Lock()
	defer ch.cmdsMu.Unlock()

	ch.cmds[cmd.Data.Name] = &cmd
}

func (ch *CommandHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	startTime := time.Now()
	ch.cmdsMu.RLock()

	var (
		cmd *Command
		ok  bool
	)
	if i.Type == discordgo.InteractionApplicationCommand ||
		i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		if cmd, ok = ch.cmds[i.ApplicationCommandData().Name]; ok {
			if !cmd.Accepts.Slash {
				return
			}
		} else {
			return
		}
	} else if i.Type == discordgo.InteractionMessageComponent {
		if cmd, ok = ch.cmds[i.MessageComponentData().CustomID]; ok {
			if !cmd.Accepts.Button {
				return
			}
		} else {
			return
		}
	}

	ch.cmdsMu.RUnlock()

	if err := cmd.Handler(s, i); err != nil {
		logger.Error(
			"Exception caught when executing a command %s, took %v - %s",
			cmd.Data.Name,
			time.Since(startTime),
			err.Error(),
		)
	} else {
		logger.Info(
			"Command %s executed, took %v",
			cmd.Data.Name,
			time.Since(startTime),
		)
	}
}

func (ch *CommandHandler) AutoHandle(s *discordgo.Session) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		go ch.Handle(s, i)
	})
}

func (ch *CommandHandler) PostCommands(s *discordgo.Session) {
	arr := []*discordgo.ApplicationCommand{}

	for _, cmd := range ch.cmds {
		if cmd.Accepts.Slash {
			d := *cmd.Data
			arr = append(arr, &d)
		}
	}

	created, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", arr)

	if err != nil {
		logger.Error("Something went wrong while posting commands, '%s'", err.Error())

		return
	}

	logger.Info("%v commands posted, %v failed", len(arr), len(arr)-len(created))
}
