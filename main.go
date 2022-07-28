package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

type User struct {
	userId   int64
	name     string
	position string
	phone    string
	state    string
	fl       string
}

const ADMIN = 398076071

func main() {
	bot, err := tgbotapi.NewBotAPI("<bot_token>")
	if err != nil {
		log.Panic(err)
	}

	users := make([]User, 0)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		fmt.Println("Users: ", users)

		userId := update.Message.Chat.ID
		user := findOrCreateUser(&users, update.Message.Chat.ID)

		fmt.Println("Users: ", users)
		fmt.Println("User: ", user)
		if update.Message.Command() == "start" {
			sendTextMessage(bot, userId, "Welcome!")
			user.state = "initial"
		}
		switch user.state {
		case "initial":
			sendTextMessage(bot, userId, "Hello! Send your name")
			user.state = "receive_name"
		case "receive_name":
			if update.Message.Text == "" {
				sendTextMessage(bot, userId, "You should send your name as text")
				break
			}

			sendTextMessage(bot, userId, "Great! Now send your applying position")
			user.name = update.Message.Text
			user.state = "receive_position"
		case "receive_position":
			if update.Message.Text == "" {
				sendTextMessage(bot, userId, "You should send your position as text")
				break
			}

			sendTextMessage(bot, userId, "Great! Now send your phone number")
			user.position = update.Message.Text
			user.state = "receive_phone"
		case "receive_phone":
			phone_number := ""
			if update.Message.Text != "" {
				for _, entity := range update.Message.Entities {
					if entity.Type == "phone_number" {
						phone_number = strings.Join(
							strings.Split(
								update.Message.Text, "",
							)[entity.Offset:entity.Offset+entity.Length],
							"",
						)
					}
				}
			} else if update.Message.Contact != nil {
				phone_number = update.Message.Contact.PhoneNumber
			}

			if phone_number == "" {
				sendTextMessage(bot, userId, "You should send phone number as +998901234567")
				break
			}

			sendTextMessage(bot, userId, "Great! Now send your file")
			user.phone = phone_number
			user.state = "receive_file"
		case "receive_file":
			if update.Message.Document == nil {
				sendTextMessage(bot, userId, "You should a file")
				break
			}

			sendTextMessage(bot, userId, "Thank you! Your documents were sent to admin")
			user.fl = update.Message.Document.FileID
			forwardToAdmin(bot, *user)
			removeUser(&users, user.userId)
		}
	}
}

func findOrCreateUser(users *[]User, userId int64) *User {
	for i := 0; i < len(*users); i++ {
		user := &(*users)[i]
		if user.userId == userId {
			fmt.Println("User found!")
			return &(*users)[i]
		}
	}

	fmt.Println("User not found!")

	user := User{
		userId:   userId,
		name:     "",
		position: "",
		phone:    "",
		state:    "initial",
		fl:       "",
	}

	*users = append(*users, user)

	return findOrCreateUser(users, userId)
}

func removeUser(users *[]User, userId int64) {
	for i := 0; i < len(*users); i++ {
		user := &(*users)[i]
		if user.userId == userId {
			(*users)[i] = (*users)[len(*users)-1]
			*users = (*users)[:len(*users)-1]
			return
		}
	}
}

func forwardToAdmin(bot *tgbotapi.BotAPI, user User) {
	file := tgbotapi.FileID(user.fl)

	doc := tgbotapi.NewDocument(ADMIN, file)
	doc.Caption = fmt.Sprintf(
		"[%s](tg://user?id=%d)'s file\nPosition: %s\nPhone: %s",
		user.name,
		user.userId,
		user.position,
		user.phone,
	)
	doc.ParseMode = "markdown"

	if _, err := bot.Send(doc); err != nil {
		log.Panic(err)
	}
}

func sendTextMessage(bot *tgbotapi.BotAPI, userId int64, message string) {
	msg := tgbotapi.NewMessage(userId, message)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}
