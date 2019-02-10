package chans

import (
	"fmt"
	"strings"

	"github.com/0x111/telegram-rss-bot/feeds"
	"github.com/0x111/telegram-rss-bot/replies"
	log "github.com/sirupsen/logrus"
	"gopkg.in/telegram-bot-api.v4"
)

// Get Feed Updates from the feeds
func FeedUpdates() {
	feedUpdates := feeds.GetFeedUpdatesChan()

	for feedUpdate := range feedUpdates {
		log.WithFields(log.Fields{"feedData": feedUpdate}).Debug("Requesting feed data")
		log.WithFields(log.Fields{"feedID": feedUpdate.ID, "feedUrl": feedUpdate.Url}).Info("Updating feeds")
		feeds.GetFeed(feedUpdate.Url, feedUpdate.ID)
	}
}

// Post Feed data to the channel
func FeedPosts(Bot *tgbotapi.BotAPI) {
	feedPosts := feeds.PostFeedUpdatesChan()

	msg := ""
	var chatid int64 = 0
	for feedPost := range feedPosts {
		title := strings.Replace(strings.Replace(feedPost.Title, "[", "\\[", -1), "]", "\\]", -1)
		msg += fmt.Sprintf(`
	[%s](%s)\r\n
	`, title, feedPost.Link)
		chatid = feedPost.ChatID
		log.WithFields(log.Fields{"feedPost": feedPost, "chatID": feedPost.ChatID}).Debug("Posting feed update to the Telegram API")
	}
	replies.SimpleMessage(Bot, chatid, 0, msg[:len(msg)-2])
	/*if err == nil {
		log.WithFields(log.Fields{"feedPost": feedPost, "chatID": chatid}).Debug("Setting the Feed Data entry to published!")
		_, err := feeds.UpdateFeedDataPublished(&feedPost)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "feedPost": feedPost, "chatID": chatid}).Error("There was an error while updating the Feed Data entry!")
		}
	} else {
		log.WithFields(log.Fields{"error": err, "feedPost": feedPost, "chatID": chatid}).Error("There was an error while posting the update to the feed!")
	}*/
}
