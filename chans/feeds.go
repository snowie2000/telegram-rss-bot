package chans

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/0x111/telegram-rss-bot/feeds"
	"github.com/0x111/telegram-rss-bot/markdownhelper"
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
	var nDelayed int32 = 0
	var msgs map[int64]string = make(map[int64]string)

	delaySend := func(s string, chatid int64) {
		if atomic.CompareAndSwapInt32(&nDelayed, 0, 1) {
			msgs[chatid] = s
			go func() {
				defer atomic.StoreInt32(&nDelayed, 0)
				<-time.After(time.Second * 5)
				for id, msg := range msgs {
					if msg != "" {
						tgmsg := tgbotapi.NewMessage(id, msg)
						tgmsg.ParseMode = "markdown"
						tgmsg.DisableWebPagePreview = true
						Bot.Send(tgmsg)
					}
				}
				msgs = make(map[int64]string)
			}()
		} else {
			if _, ok := msgs[chatid]; ok {
				msgs[chatid] += "\n" + s
			} else {
				msgs[chatid] = s
			}
		}
	}
	for feedPost := range feedPosts {
		delaySend(fmt.Sprintf("[%s](%s)", markdownhelper.MDEscape(feedPost.Title), markdownhelper.MDEscape(feedPost.Link)), feedPost.ChatID)

		log.WithFields(log.Fields{"feedPost": feedPost, "chatID": feedPost.ChatID}).Debug("Posting feed update to the Telegram API")
		//if err == nil {
		//	log.WithFields(log.Fields{"feedPost": feedPost, "chatID": feedPost.ChatID}).Debug("Setting the Feed Data entry to published!")
		_, err := feeds.UpdateFeedDataPublished(&feedPost)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "feedPost": feedPost, "chatID": feedPost.ChatID}).Error("There was an error while updating the Feed Data entry!")
		}
		//} else {
		//	log.WithFields(log.Fields{"error": err, "feedPost": feedPost, "chatID": feedPost.ChatID}).Error("There was an error while posting the update to the feed!")
		//}
	}
}
