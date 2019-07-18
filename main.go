package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

type TelegramBot struct {
	Bot    *tgbotapi.BotAPI
	ChatID int64
}

var (
	telegrambot = TelegramBot{}
)

func main() {
	bottoken := os.Getenv("BOTTOKEN")
	developerKey := os.Getenv("DEVELOPERKEY")

	if bottoken == "" || developerKey == "" {
		fmt.Print("BotToken or DeveloperKey is does not exist")
		os.Exit(1)
	}

	telegrambot.botInit(bottoken)

	log.Printf("Authorized on account %s", telegrambot.Bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := telegrambot.Bot.GetUpdatesChan(u)

	for update := range updates {
		//
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		user := update.Message.Chat.ID
		go func() {

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			text := update.Message.Text

			if text == "/start" || text == "/help" {
				msg := tgbotapi.NewMessage(telegrambot.ChatID, "Введите название автора и название песни")
				_, err := telegrambot.Bot.Send(msg)
				if err != nil {
					log.Error(err)
				}
			} else {
				msg := tgbotapi.NewMessage(telegrambot.ChatID, "Поиск...")
				_, err := telegrambot.Bot.Send(msg)
				if err != nil {
					log.Error(err)
				}
				id, _, _ := SearchingVideo(text, developerKey)

				if id != "" {
					go func() {
						ticker := time.NewTicker(time.Second * 10)
						for range ticker.C {
							if song != "" {
								ticker.Stop()
							}
							msg := tgbotapi.NewMessage(telegrambot.ChatID, "Конвертация")
							_, err := telegrambot.Bot.Send(msg)
							if err != nil {
								//log.Error(err)
							}
						}
					}()
					ConvertingVideo(id)

					filemp3 := tgbotapi.NewAudioUpload(user, id+".mp3")
					_, err := telegrambot.Bot.Send(filemp3)
					if err != nil {
						msg := tgbotapi.NewMessage(telegrambot.ChatID, "Ошибка отправки файла")
						_, err := telegrambot.Bot.Send(msg)
						if err != nil {
							log.Error(err)
						}
					}
				}
			}
		}()
	}
}

func SearchingVideo(n, key string) (string, string, error) {
	var VideoID, VideoName string
	client := &http.Client{
		Transport: &transport.APIKey{Key: key},
	}

	service, err := youtube.New(client)
	if err != nil {
		msg := tgbotapi.NewMessage(telegrambot.ChatID, "Ошибка загрузки видео...")
		_, err := telegrambot.Bot.Send(msg)
		if err != nil {
			log.Error(err)
		}
		return "", "", err
	}

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").
		Q(n).
		MaxResults(1)

	response, err := call.Do()
	if err != nil {
		//SendMsg("Ошибка поиска...попробуйте еще раз")
		return "", "", err
	}

	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			VideoID = item.Id.VideoId
			VideoName = item.Snippet.Title
		}
	}
	return VideoID, VideoName, err
}

func ConvertingVideo(n string) error {

	cmd := exec.Command("youtube-dl", "-x", "--audio-format", "mp3", "--audio-quality", "9", "-o", "%(id)s.%(ext)s", n)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		msg := tgbotapi.NewMessage(telegrambot.ChatID, "Ошибка конвертации")
		_, err := telegrambot.Bot.Send(msg)
		if err != nil {
			log.Error(err)
		}
	}
	return err
}

func (b *TelegramBot) botInit(n string) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(n)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	b.Bot = bot
	return b.Bot
}
