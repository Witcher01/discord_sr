package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/Witcher01/discord_sr/srcom"
	"github.com/bwmarrin/discordgo"

	"github.com/akamensky/argparse"
)

type commandArgs struct {
	game_id *string
	category *string
	amount *int
	platform *string
}

func main() {
	parser := argparse.NewParser("discord_sr", "A bot for discord that returns simple queries for speedrun.com")

	token := parser.String("t", "token", &argparse.Options{Required: true, Help: "Authentication token to use for the bot"})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Println(parser.Usage(err))
		return
	}

	discord, err := discordgo.New("Bot " + *token)

	// Cleanly close down the Discord session on program end
	defer discord.Close()

	if err != nil {
		fmt.Println("error creating bot,", err)
		return
	}

	discord.AddHandler(messageCreate);

	discord.Identify.Intents = discordgo.IntentsGuildMessages

	err = discord.Open()

	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "*ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
		return
	}

	if m.Content == "*pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
		return
	}

	if strings.HasPrefix(m.Content, "*records") {
		args, err := leaderboardArgsParser(m.Content, "records", "Return the name of the top runners for a specified game")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		top, err := srcom.GetTopRunnersByGame(*args.game_id, *args.category, *args.amount, *args.platform)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, top)
		return
	}

	if strings.HasPrefix(m.Content, "*top") {
		args, err := leaderboardArgsParser(m.Content, "top", "Return the top runs of a game")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		top, err := srcom.GetTopByGame(*args.game_id, *args.category, *args.amount, *args.platform)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, top)
		return
	}

	if strings.HasPrefix(m.Content, "*wr") {
		args, err := leaderboardArgsParser(m.Content, "wr", "Return the wr of the specified game")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		wr, err := srcom.GetWRByGame(*args.game_id, *args.category, *args.platform)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
		}

		s.ChannelMessageSend(m.ChannelID, wr)
		return
	}
}

func leaderboardArgsParser(input string, name string, desc string) (*commandArgs, error) {
	parser := argparse.NewParser(name, desc)
	gameid := parser.String("i", "game-id", &argparse.Options{Required: true, Help: "The game id found in the URL of the leaderboard"})
	amount := parser.Int("a", "amount", &argparse.Options{Required: false, Help: "The amount of records to be returned", Default: 3})
	platform := parser.String("p", "platform", &argparse.Options{Required: false, Help: "The platform for which the leaderboard should be returned", Default: *new(string)})
	category := parser.String("c", "category", &argparse.Options{Required: false, Help: "The category for which the leaderboard should be returned", Default: *new(string)})

	r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
	split := r.FindAllString(input, -1)
	err := parser.Parse(split)
	if err != nil {
		return nil, errors.New(parser.Usage(err))
	}

	// trim quotation marks from category to allow category names with spaces
	*category = strings.Trim(*category, "\"")

	return &commandArgs{gameid, category, amount, platform}, nil
}
