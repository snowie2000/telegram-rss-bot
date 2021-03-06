package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/snowie2000/telegram-rss-bot/chans"
	"github.com/snowie2000/telegram-rss-bot/commands"
	"github.com/snowie2000/telegram-rss-bot/conf"
	"github.com/snowie2000/telegram-rss-bot/db"
	"github.com/snowie2000/telegram-rss-bot/migrations"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func main() {
	conf.LoadConfig()
	dbc := db.ConnectDB()
	defer dbc.Close()

	var (
		err error
		Bot *tgbotapi.BotAPI
	)

	// Read config
	config := conf.GetConfig()

	for {
		Bot, err = tgbotapi.NewBotAPI(config.GetString("telegram_auth_key"))

		if err == nil {
			break
		} else {
			log.Println("Error:", err, "retrying in 10s")
			time.Sleep(10 * time.Second)
		}
	}

	Bot.Debug = config.GetBool("telegram_api_debug")

	log.Debug("Authorized on account ", Bot.Self.UserName)

	// create basic database structure
	migrations.Migrate()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// read rss data from channels
	go func() {
		chans.FeedUpdates()
	}()

	// post rss data to channels
	go func() {
		chans.FeedPosts(Bot)
	}()

	// read feed updates from the Telegram APIS
	updates, err := Bot.GetUpdatesChan(u)

	for update := range updates {

		// if the message is empty, we do not need to handle anything
		if update.Message == nil {
			continue
		}

		// allow only private conversations for the bot now
		if int64(update.Message.From.ID) != update.Message.Chat.ID {
			continue
		}

		// handle add command
		if update.Message.IsCommand() && update.Message.Command() == "add" {
			commands.AddCommand(Bot, &update)
		}

		// handle delete command
		if update.Message.IsCommand() && update.Message.Command() == "delete" {
			commands.DeleteCommand(Bot, &update)
		}

		// handle list command
		if update.Message.IsCommand() && update.Message.Command() == "list" {
			commands.ListCommand(Bot, &update)
		}

		// handle list command
		if update.Message.IsCommand() && update.Message.Command() == "help" {
			commands.HelpCommand(Bot, &update)
		}
	}
}
