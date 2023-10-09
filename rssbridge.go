package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	botToken := os.Getenv("TOKEN")     // токен бота
	channelID := os.Getenv("CHANELID") // идентификатор канала

	rssBridgeURL := os.Getenv("BRIDGEURL") // адрес бриджы

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	// включаем счетчик обосрышей
	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Создаем новый парсер для бриджи
	parser := gofeed.NewParser()

	// Загрузка списка уже отправленных постов из файла
	sentPosts, err := loadSentPosts()
	if err != nil {
		log.Println("Failed to load sent posts:", err)
	}

	// Непрерывный цикл для получения обновлений из RSS бриджи
	for {
		// Получаем данные из RSS Bridge
		feed, err := parser.ParseURL(rssBridgeURL)
		if err != nil {
			log.Println("Failed to parse RSS:", err)
			continue
		}

		// Перебираем элементы в RSS бридже
		for _, item := range feed.Items {
			// Проверяем, был ли этот пост уже отправлен
			if postSent(sentPosts, item.Title) {
				continue // Пропускаем уже отправленные посты
			}

			msg := formatMessage(item) // Форматируем текст сообщения

			// Создаем новое сообщение для канала Telegram
			message := tgbotapi.NewMessageToChannel(channelID, msg)

			// Устанавливаем режим разметки HTML для форматирования
			message.ParseMode = "HTML"

			// Отправляем сообщение в канал Telegram
			_, err = bot.Send(message)
			if err != nil {
				log.Println("Failed to send message to Telegram:", err)
				continue
			}

			log.Println("Message sent to Telegram:", msg)

			// Добавляем отправленный пост в список
			sentPosts = append(sentPosts, item.Title)
		}

		// Сохраняем список отправленных постов в файл
		err = saveSentPosts(sentPosts)
		if err != nil {
			log.Println("Failed to save sent posts:", err)
		}
	}
}

// Функция для форматирования текста сообщения из элемента RSS
func formatMessage(item *gofeed.Item) string {
	// Создаем переменную, которая будет хранить текст сообщения
	var formattedMsg string

	// Добавляем заголовок сообщения
	formattedMsg += fmt.Sprintf("<b>%s</b>\n", item.Title)

	// 	ДОбовляем описание из поста
	if item.Description != "" {
		formattedMsg += fmt.Sprintf("%s\n", item.Description)
	}

	// Вставляем ссылку на оригинальный пост
	formattedMsg += fmt.Sprintf("<a href=\"%s\">Read More</a>", item.Link)

	return formattedMsg
}

// Функция для проверки дублирования поста
func postSent(sentPosts []string, title string) bool {
	for _, sentPost := range sentPosts {
		if strings.TrimSpace(sentPost) == strings.TrimSpace(title) {
			return true
		}
	}
	return false
}

// Функция для загрузки списка отправленных постов из файла
func loadSentPosts() ([]string, error) {
	data, err := ioutil.ReadFile("sent_posts.go")
	if err != nil {
		return nil, err
	}

	sentPosts := strings.Split(string(data), "\n")

	// Удаляем насранные пустые строки
	for len(sentPosts) > 0 && sentPosts[len(sentPosts)-1] == "" {
		sentPosts = sentPosts[:len(sentPosts)-1]
	}

	return sentPosts, nil
}

// Функция для сохранения списка отправленных постов в файл
func saveSentPosts(sentPosts []string) error {
	data := strings.Join(sentPosts, "\n")

	err := ioutil.WriteFile("sent_posts.txt", []byte(data), 0644)
	if err != nil {
		return err
	}

	return nil
}
