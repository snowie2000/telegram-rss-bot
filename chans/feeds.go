package chans

import (
	"fmt"
	"html"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/snowie2000/telegram-rss-bot/feeds"
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

type feedcache map[int]string

// Post Feed data to the channel
func FeedPosts(Bot *tgbotapi.BotAPI) {
	feedPosts := feeds.PostFeedUpdatesChan()
	var nDelayed int32 = 0
	var msgs map[int64]feedcache = make(map[int64]feedcache)

	delaySend := func(s string, feedid int, chatid int64) {
		updateMsg := func() {
			if c, ok := msgs[chatid]; ok {
				if _, ok = c[feedid]; ok {
					c[feedid] += "\n" + s
				} else {
					c[feedid] = s
				}
			} else {
				msgs[chatid] = make(feedcache)
				msgs[chatid][feedid] = s
			}
		}

		if atomic.CompareAndSwapInt32(&nDelayed, 0, 1) {
			go func() {
				defer atomic.StoreInt32(&nDelayed, 0)
				<-time.After(time.Second * 5)
				for chatid, cache := range msgs {
					for feedid, msg := range cache {
						if msg != "" {
							msg = "<b>" + feeds.GetFeedName(feedid, chatid) + "</b>\n" + msg
							tgmsg := tgbotapi.NewMessage(chatid, msg)
							tgmsg.ParseMode = "HTML"
							tgmsg.DisableWebPagePreview = true
							Bot.Send(tgmsg)
						}
					}
				}
				msgs = make(map[int64]feedcache)
			}()
		}
		updateMsg()
	}
	for feedPost := range feedPosts {
		delaySend(fmt.Sprintf("<a href=\"%s\">%s</a>", html.EscapeString(feedPost.Link), html.EscapeString(feedPost.Title)), feedPost.FeedID, feedPost.ChatID)

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
