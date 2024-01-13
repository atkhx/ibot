package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/robfig/cron"
)

var (
	configFileName = flag.String("config", "config.json", "config file name")
	parseImgRegexp *regexp.Regexp
)

type Config struct {
	ChatID       int64  `json:"chatID"`
	BotToken     string `json:"botToken"`
	Schedule     string `json:"schedule"`
	ParsePageURL string `json:"parsePageURL"`
	ParseImageRe string `json:"parseImageRe"`
	ImageBaseURL string `json:"imageBaseURL"`
}

func main() {
	flag.Parse()

	cfg, err := readConfig(*configFileName)
	if err != nil {
		log.Fatalln(err)
	}

	parseImgRegexp = regexp.MustCompile(cfg.ParseImageRe)

	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalln("init bot:", err)
	}

	c := cron.New()
	if err := c.AddFunc(cfg.Schedule, func() {
		if err := sendImageToChat(bot, cfg); err != nil {
			log.Println(sendImageToChat(bot, cfg))

			bot.Send(tgbotapi.NewMessage(
				cfg.ChatID,
				fmt.Sprintf("Не смог отправить картинку: %s", err.Error()),
			))
		}
	}); err != nil {
		log.Fatalln("schedule job:", err)
	}
	c.Run()
}

func readConfig(filename string) (*Config, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func sendImageToChat(bot *tgbotapi.BotAPI, cfg *Config) (err error) {
	imageURL, err := getImageURL(cfg.ParsePageURL, cfg.ImageBaseURL)
	if err != nil {
		return fmt.Errorf("getImageURL: %w", err)
	}

	imageBytes, err := getImageBytes(imageURL)
	if err != nil {
		return fmt.Errorf("getImageBytes: %w", err)
	}

	photo := tgbotapi.NewPhotoUpload(cfg.ChatID, tgbotapi.FileBytes{
		Name:  "stat.png",
		Bytes: imageBytes,
	})

	photo.Caption = fmt.Sprintf(time.Now().Format("2006-01-02 15:04:05Z07:00"))
	if _, err := bot.Send(photo); err != nil {
		return fmt.Errorf("sendImage: %w", err)
	}
	return nil
}

func getImageBytes(imageURL string) ([]byte, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("get image file: %w", err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read image body: %w", err)
	}
	defer resp.Body.Close()
	return b, nil
}

func getImageURL(parsePageURL, imageBaseURL string) (string, error) {
	b, err := getPageBody(parsePageURL)
	if err != nil {
		return "", err
	}

	matches := parseImgRegexp.FindAllStringSubmatch(string(b), -1)
	if len(matches) == 0 {
		return "", fmt.Errorf("no images on page")
	}
	return fmt.Sprintf("%s%s", imageBaseURL, matches[0][1]), nil
}

func getPageBody(parsePageURL string) ([]byte, error) {
	resp, err := http.Get(parsePageURL)
	if err != nil {
		return nil, fmt.Errorf("get page: %w", err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	defer resp.Body.Close()
	return b, nil
}
