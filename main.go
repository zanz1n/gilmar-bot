package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/gilmar-bot/logger"
)

type StatusType uint8

const (
	StatusTypeStarting StatusType = 96
	StatusTypeStopping StatusType = 73
	StatusTypeIdle     StatusType = 24
)

var (
	token   = flag.String("token", "", "The discord bot token")
	dataDir = flag.String("data-dir", "./data", "The application data directory")
)

func init() {
	if *token == "" {
		*token = os.Getenv("DISCORD_TOKEN")
	}
	flag.Parse()
}

func onReady(manager *CommandHandler) func(s *discordgo.Session, r *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		logger.Info(
			"Logged in as %s#%s",
			s.State.User.Username,
			s.State.User.Discriminator,
		)

		manager.PostCommands(s)

		SetStatus(s, StatusTypeIdle)
	}
}

func main() {
	s, err := discordgo.New("Bot " + *token)

	if err != nil {
		logger.Fatal(err)
	}

	manager := NewCommandHandler()

	phrasesRepo := NewRepository[[]Phrase](*dataDir + "/phrases.obj")
	go phrasesRepo.BackgroundSave()

	percentRepo := NewRepository[uint8](*dataDir + "/percentage.obj")
	go percentRepo.BackgroundSave()

	defaultPhraseRepo := NewRepository[Phrase](*dataDir + "/default-phrases.obj")
	go defaultPhraseRepo.BackgroundSave()

	manager.Add(PingCommand())
	manager.Add(PhraseCommand(phrasesRepo))

	manager.AutoHandle(s)
	s.AddHandler(onReady(manager))
	s.AddHandler(onMessage(phrasesRepo, percentRepo, defaultPhraseRepo))

	endCh := make(chan os.Signal, 1)

	signal.Notify(endCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	if err = s.Open(); err != nil {
		logger.Fatal(err)
	}

	SetStatus(s, StatusTypeStarting)

	<-endCh
	SetStatus(s, StatusTypeStopping)
	time.Sleep(200 * time.Millisecond)
	logger.Info("Stopping...")

	s.Close()

	phrasesRepo.Save()
	percentRepo.Save()
	defaultPhraseRepo.Save()
}
