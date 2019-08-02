package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/Sirupsen/logrus"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

//TelegramBot struct
type TelegramBot struct {
	Bot    *tgbotapi.BotAPI
	ChatID int64
}

var (
	telegrambot = TelegramBot{}
)

func main() {
	bottoken, devkey := GetTokenAndDevKey()
	telegrambot.botInit(bottoken)

	log.Printf("Authorized on account %s", telegrambot.Bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := telegrambot.Bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		user := update.Message.Chat.ID
		go func() {

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			text := update.Message.Text

			if text == "/start" || text == "/help" {
				msg := tgbotapi.NewMessage(user, "Введите название автора и название песни")
				telegrambot.Bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(user, "Поиск...")
				telegrambot.Bot.Send(msg)
				id, _ := SearchingVideo(text, devkey)

				var ok bool

				go func() {
					ticker := time.NewTicker(time.Second * 7)
					for range ticker.C {
						if ok == true {
							ticker.Stop()
						}
						msg := tgbotapi.NewMessage(user, "Конвертация...")
						telegrambot.Bot.Send(msg)
					}
				}()
				ok, _ = ConvertingVideo(id)

				filemp3 := tgbotapi.NewAudioUpload(user, id+".mp3")
				_, err := telegrambot.Bot.Send(filemp3)
				if err != nil {
					msg := tgbotapi.NewMessage(user, "Ошибка отправки файла")
					telegrambot.Bot.Send(msg)
				}

			}
		}()
	}
}

func youtubeClient(key string) (*youtube.Service, error) {
	client := &http.Client{
		Transport: &transport.APIKey{Key: key},
	}
	service, err := youtube.New(client)
	if err != nil {
		logrus.Info(err)
		return nil, err
	}
	return service, err
}

//SearchingVideo by keyword
func SearchingVideo(keyword, key string) (string, error) {
	service, _ := youtubeClient(key)
	var VideoID string

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").
		Q(keyword).
		MaxResults(1)

	response, err := call.Do()
	if err != nil {
		//SendMsg("Ошибка поиска...попробуйте еще раз")
		return "", err
	}

	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			VideoID = item.Id.VideoId
		}
	}
	return VideoID, err
}

//ConvertingVideo function
func ConvertingVideo(n string) (bool, error) {

	cmd := exec.Command("youtube-dl", "-x", "--audio-format", "mp3", "--audio-quality", "9", "-o", "%(id)s.%(ext)s", n)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		/* 	msg := tgbotapi.NewMessage(telegrambot.ChatID, "Ошибка конвертации")
		telegrambot.Bot.Send(msg) */
		return false, err
	}
	return true, err
}

//botInit function
func (b *TelegramBot) botInit(n string) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(n)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	b.Bot = bot
	return b.Bot
}

//GetTokenAndDevKey function
func GetTokenAndDevKey() (string, string) {
	bottoken := os.Getenv("BOTTOKEN")
	developerKey := os.Getenv("DEVELOPERKEY")

	if bottoken == "" || developerKey == "" {
		fmt.Print("BotToken or DeveloperKey is does not exist")
		return "", ""
		log.Fatal()
	}
	return bottoken, developerKey
}
