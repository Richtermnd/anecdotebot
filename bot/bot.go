package bot

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Richtermnd/anecdotebot/client"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	helpText = `/help - список команд
	/list - список категорий
	/delay <задержка> - задержка между анекдотами
	/category <номер категории> - выбор категории
	/anecdote - случайный анекдот`

	listText = `1 - Анекдот;
	2 - Рассказы;
	3 - Стишки;
	4 - Афоризмы;
	5 - Цитаты;
	6 - Тосты;
	8 - Статусы;
	11 - Анекдот (+18);
	12 - Рассказы (+18);
	13 - Стишки (+18);
	14 - Афоризмы (+18);
	15 - Цитаты (+18);
	16 - Тосты (+18);
	18 - Статусы (+18);`

	sessionFile = "storage/sessions.json"
)

var (
	bot      *tgbotapi.BotAPI
	sessions = make(map[int64]*session)
	log      = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
)

type session struct {
	ID       int64
	Category int           `json:"category"`
	Delay    time.Duration `json:"delay"`
	cancel   func()        `json:"-"`
}

func Session(id int64) *session {
	s, ok := sessions[id]
	if ok {
		return s
	}
	s = &session{
		ID:       id,
		Category: 1,
		Delay:    time.Hour,
	}
	sessions[id] = s
	return s
}

func (s *session) SendAnecdote() {
	log := log.With("id", s.ID)
	log.Debug("send anecdote")
	SendMessage(s.ID, client.GetAnecdote(s.Category))
	log.Debug("anecdote sent")
}

func (s *session) Notify() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go func(ctx context.Context) {
		for {
			log := log.With("id", s.ID)
			select {
			case <-ctx.Done():
				return
			default:
				log.Debug("send anecdote")
				s.SendAnecdote()
			}
			time.Sleep(s.Delay)
		}
	}(ctx)
}

func (s *session) StopNotify() {
	if s.cancel == nil {
		return
	}
	s.cancel()
}

func InitBot() {
	uploadSessions()
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		panic("empty token")
	}
	var err error
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}
}

func Listen() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)

	log.Info("Bot started")

	for update := range updates {
		// If not message - skip
		if update.Message == nil {
			continue
		}
		// Handle command
		if update.Message.IsCommand() {
			log.Info("Receive command", "from", update.Message.From.ID, "command", update.Message.Command())
			handleCommand(update.Message)
			continue
		}
		SendMessage(update.Message.Chat.ID, "/help чтобы узнать список команд.")
	}
}

func SendMessage(id int64, text string) {
	reply := tgbotapi.NewMessage(id, text)
	bot.Send(reply)
}

func SaveSessions() {
	log.Info("saving sessions")
	f, err := os.Create(sessionFile)
	if err != nil {
		panic(err)
	}
	json.NewEncoder(f).Encode(sessions)
}

func handleCommand(msg *tgbotapi.Message) {
	// msg.Text
	switch msg.Command() {
	case "category":
		// set category
		setCategory(msg)
	case "delay":
		// set delay
		setDelay(msg)
	case "stop":
		s := Session(msg.Chat.ID)
		s.StopNotify()
	case "anecdote":
		// send anecdote
		s := Session(msg.Chat.ID)
		s.SendAnecdote()
	case "list":
		// send list of categories
		SendMessage(msg.Chat.ID, listText)
	case "help":
		SendMessage(msg.Chat.ID, helpText)
	default:
		SendMessage(msg.Chat.ID, "/help чтобы узнать список команд.")
	}
}

func setCategory(msg *tgbotapi.Message) {
	log := log.With("id", msg.Chat.ID)
	log.Debug("set category for")
	args := getArgs(msg)
	if len(args) != 1 {
		log.Debug("invalid args")
		SendMessage(msg.Chat.ID, "Нужно указать категорию.")
		return
	}

	category, err := strconv.Atoi(args[0])
	if err != nil {
		log.Debug("invalid category")
		SendMessage(msg.Chat.ID, "Категория должна быть числом от 1 до 18")
		return
	}

	if category < 1 || category > 18 {
		log.Debug("invalid category")
		SendMessage(msg.Chat.ID, "Категория должна быть числом от 1 до 18")
		return
	}
	s := Session(msg.Chat.ID)
	s.Category = category
	SendMessage(msg.Chat.ID, "Категория изменена.")
	log.Debug("category setted", "category", category)
}

func setDelay(msg *tgbotapi.Message) {
	log := log.With("id", msg.Chat.ID)
	log.Debug("set delay for")
	args := getArgs(msg)
	if len(args) != 1 {
		log.Debug("invalid args")
		SendMessage(msg.Chat.ID, "Нужно указать время.\nПримеры:\n1d - 1 день\n2h10m - 2 часа 10 минут \n30m - 30 мин\n15s - 15 сек")
		return
	}
	dur, err := time.ParseDuration(args[0])
	if err != nil {
		log.Debug("invalid time")
		SendMessage(msg.Chat.ID, "Нужно указать время.\nПримеры:\n1d - 1 день\n2h10m - 2 часа 10 минут \n30m - 30 мин\n15s - 15 сек")
		return
	}
	s := Session(msg.Chat.ID)
	s.Delay = dur
	s.StopNotify()
	s.Notify()
	log.Debug("delay setted", "delay", dur)
}

func getArgs(msg *tgbotapi.Message) []string {
	splitted := strings.Split(msg.Text, " ")
	return splitted[1:]
}

func uploadSessions() {
	log.Info("loading sessions")
	f, err := os.OpenFile(sessionFile, os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	json.NewDecoder(f).Decode(&sessions)
}
