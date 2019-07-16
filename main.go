package main

import (
	"log"
	"os"
	"fmt"
	"net/http"
	"os/exec"
	"time"
	"path/filepath"
	"regexp"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)
type Video struct {
	VideoID           string
	VideoName         string
}

type TelegramBot struct {
	Bot *tgbotapi.BotAPI
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
	var (
		video = Video{}
	)
	telegrambot.botInit(bottoken)

	log.Printf("Authorized on account %s", telegrambot.Bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	
	updates, _ := telegrambot.Bot.GetUpdatesChan(u)
	
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
			
		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			
		text := update.Message.Text
		telegrambot.ChatID = update.Message.Chat.ID
       
		if text == "/start" || text == "/help" {
			SendMsg("Введите автора и название песни")
		} else {
			SendMsg("Поиск...")
			video.SearchingVideo(text,developerKey)
	        song := ""
			if video.VideoID != ""{
			 	go func() {
					ticker := time.NewTicker(time.Second * 5)
					for  range ticker.C {
						if song != "" {
						ticker.Stop()
						}
						SendMsg("Конвертация...")
					}
				}()
				video.ConvertingVideo(video.VideoID)
				song = checkExt(".mp3")
				
				
				filemp3 := tgbotapi.NewAudioUpload(update.Message.Chat.ID, song)
				_, err := telegrambot.Bot.Send(filemp3)
				if err != nil {
				SendMsg("Ошибка отправки файла...")
				}
				os.Remove(song)
			}
		}
	}
}


func (v *Video) SearchingVideo(n,key string) (string, error) {

	client := &http.Client{
		Transport: &transport.APIKey{Key: key},
	}

	service, err := youtube.New(client)
	if err != nil {
		SendMsg("Ошибка поиска...попробуйте еще раз")
		return "", err
	}

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").
		Q(n).
		MaxResults(1)

	response, err := call.Do()
	if err != nil {
		SendMsg("Ошибка поиска...попробуйте еще раз")
		return "", err
	}

	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
		v.VideoID   = item.Id.VideoId
		}
	}
	return v.VideoID, err
}

func (v *Video) ConvertingVideo(n string) (error) {

	cmd := exec.Command("youtube-dl","-x", "--audio-format", "mp3", "--audio-quality", "9", n)
 	cmd.Stdout = os.Stdout
	cmd.Stdin  = os.Stdin
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()

	if err != nil {
		SendMsg("Ошибка конвертации...")
		return err
	}
	return err
}

func (b *TelegramBot) botInit(n string) *tgbotapi.BotAPI{
	bot, err := tgbotapi.NewBotAPI(n)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	b.Bot = bot
	return b.Bot
}

func SendMsg(textMsg string) error {
  	msg := tgbotapi.NewMessage(telegrambot.ChatID, textMsg)
	_, err := telegrambot.Bot.Send(msg)
	if err != nil {
		log.Fatal()
		return err
	}
	return nil
}

func checkExt(ext string) string {
	pathS, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var files string
	filepath.Walk(pathS, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			r, err := regexp.MatchString(ext, f.Name())
			if err == nil && r {
				files = f.Name()
			}
		}
		return nil
	})
	return files
}