package cmd

import (
	"DiscordBotAltuha/pkg/Constants"
	"DiscordBotAltuha/pkg/Error"
	"DiscordBotAltuha/pkg/logger/logger"
	"errors"
	"github.com/bwmarrin/discordgo"
	"strings"
	"sync"
)

const BotCommands = 7

// GetBotsCommands - функция, возвращающая команды используемые ботом
func GetBotsCommands() []*discordgo.ApplicationCommand {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Проверка работоспособности бота",
		},
		{
			Name:        "you",
			Description: "Рассказывает о себе",
		},
		{
			Name:        "time",
			Description: "Пишет текущее время. А хули, может кому-то надо",
		},
		{
			Name:        "talk",
			Description: "Отправить сообщение",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "Текст сообщения",
					Required:    true,
				},
			},
		},
		{
			Name:        "get",
			Description: "Получить steam информацию о участнике сервера",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "discordid",
					Description: "Discord ID",
					Required:    true,
				},
			},
		},
		{
			Name:        "add",
			Description: "добавить сопоставление discordID <=> steamID",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "discordid",
					Description: "Discord ID",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "steamid",
					Description: "Steam ID",
					Required:    true,
				},
			},
		},
		{
			Name:        "delete",
			Description: "удалить сопоставление discordID <=> steamID",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "discordid",
					Description: "Discord ID",
					Required:    true,
				},
			},
		},
	}
	return commands
}

// MessageForBot определяет, адресовано ли сообщение боту
func MessageForBot(message string) bool {
	ToLowerMessage := strings.ToLower(strings.TrimSpace(message))
	forBot := []string{"альтуха", "альтуху", "альтухе", "альтушка", "альтушку", "альтушке"}
	for _, v := range forBot {
		if strings.Contains(ToLowerMessage, v) {
			return true
		}
	}
	return false
}

// StartBot - функция для создания сессии бота, чтобы несколько раз не создавалась
func StartBot(botToken string, logs *logger.Log) (*discordgo.Session, error) {
	bot, errSession := discordgo.New("Bot " + botToken)
	if errSession != nil {
		logs.Error(Error.SessionError+"\n"+errSession.Error(), logger.GetPlace())
		return nil, errSession
	}
	info, err := bot.GatewayBot()
	if err != nil {
		logs.Error(Error.SessionLimit+"\n"+err.Error(), logger.GetPlace())
		return nil, errors.New(Error.SessionLimit + ": " + err.Error())
	}
	_ = info
	//Недоделанная часть, но ее смысла особо делать нет наверное
	//fmt.Println(info.SessionStartLimit)
	//if info.SessionStartLimit.Total < 990 {
	//	logs.Error(Error.SessionLimit, logger.GetPlace())
	//	return nil, errors.New(Error.SessionLimit)
	//}
	logs.Info(Constants.SessionSuccess, logger.GetPlace())
	return bot, nil
}

// Код для определения, является ли сообщение сообщением из канала или это личное сообщение для бота
// Этот бот работает только на сервере "не придумал"
var (
	dmCache    = make(map[string]bool) // Кеширование, чтобы быстрее потом работало
	cacheMutex sync.Mutex              //mutex для мапы
)

func IsDirectMessage(s *discordgo.Session, channelID string) bool {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if isDM, exists := dmCache[channelID]; exists {
		return isDM
	}

	channel, err := s.Channel(channelID)
	if err != nil {
		return false
	}

	isDM := channel.Type == discordgo.ChannelTypeDM
	dmCache[channelID] = isDM
	return isDM
}
