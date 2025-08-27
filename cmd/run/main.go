package main

import (
	"DiscordBotAltuha/AI"
	"DiscordBotAltuha/SomeApies"
	"DiscordBotAltuha/cmd"
	"DiscordBotAltuha/databaseMethods"
	"DiscordBotAltuha/pkg/Constants"
	"DiscordBotAltuha/pkg/Error"
	"DiscordBotAltuha/pkg/logger/logger"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	//Логи
	logs := logger.NewLog()
	// Объявляем RateLimiter для общения с духом
	RateLimiter := cmd.NewSimpleRateLimiter("", time.Now()) // Для ИИ от спама, можно писать ИИ раз в 5 секунд

	// Достаем системный промт
	systemPromt, errSysPromt := AI.GetSystemPromt(Constants.PathToBotSystemtxt, logs)
	if errSysPromt != nil {
		logs.Error(Error.SystemPromtFileDoesNotOpen+": "+errSysPromt.Error(), logger.GetPlace())
		return
	}

	// Подключаемся к бд
	db, err := databaseMethods.OpenDatabase(Constants.PathToDataBasetxt, logs)
	if err != nil {
		logs.Error(Error.DatabaseError+": "+err.Error(), logger.GetPlace())
		return
	}
	// закрываем соединение с бд
	defer func() {
		sqldb, _ := db.DB()
		sqldb.Close()
	}()
	// Получаем DiscordID и SteamID
	Compares, errCmpr := databaseMethods.GetDiscordSteamID(db)
	if errCmpr != nil {
		logs.Error(Error.DatabaseError+": "+errCmpr.Error(), logger.GetPlace())
		return
	}

	// Настраиваем переменные среды
	AIApi := os.Getenv("AI_API_KEY_ALTUHA")
	if AIApi == "" {
		logs.Error(Error.ApiKeyIsEmpty, logger.GetPlace())
		panic("AI_API_KEY environment variable not set")
	}
	botToken := os.Getenv("DISCORD_BOT_TOKEN_ALTUHA")
	if botToken == "" {
		logs.Error(Error.BotTokenIsEmpty, logger.GetPlace())
		panic("DISCORD_BOT_TOKEN environment variable not set")
	}
	steamApi := os.Getenv("STEAM_API")
	if steamApi == "" {
		logs.Error(Error.BotTokenIsEmpty, logger.GetPlace())
		panic("STEAM_API environment variable not set")
	}

	// Создаем сессию Discord
	dg, errSession := cmd.StartBot(botToken, logs)
	defer dg.Close()
	if errSession != nil {
		// внутри функции логи уже сделаны
		return
	}

	// Обработчик события "готовности" бота
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		logs.Info(fmt.Sprintf("Бот запущен как %s#%s", r.User.Username, r.User.Discriminator), logger.GetPlace())
	})

	// Slash-команды
	commands := cmd.GetBotsCommands()

	// Сообщения из чата
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Игнорируем сообщения от ботов
		if m.Author.Bot {
			logs.Info("Альтухе написал бот: "+m.Author.Username, logger.GetPlace())
			return
		}
		//Если это личное сообщение
		if cmd.IsDirectMessage(s, m.ChannelID) {
			_, errChan := s.ChannelMessageSend(m.ChannelID, Constants.TalksOnlyInServer+m.Author.Username)
			if errChan != nil {
				logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
			}
			return
		}
		// Диалог с альтухой без слеш-команды
		if cmd.MessageForBot(m.Content) {
			message := "Тебе пишет " + m.Author.Username + ": " + m.Content
			AiMessage, _ := AI.Promt(m.Author.Username, message, systemPromt, AIApi, RateLimiter, logs)
			_, errSend := s.ChannelMessageSend(m.ChannelID, AiMessage)
			//DB
			databaseMethods.DBMutex.Lock()
			databaseMethods.DBNewAction(m.Author.Username, Constants.CommandTalk, db, logs) // заносим событие в базу данных
			databaseMethods.DBMutex.Unlock()

			if errSend != nil {
				logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
			}
		}
	})
	// Обработчик Slash-команд
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		cmds := i.ApplicationCommandData()
		//Если личное сообщение(нам такого не надо)
		if cmd.IsDirectMessage(s, i.ChannelID) {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: Constants.TalksOnlyInServer + i.User.Username,
				},
			})
			if err != nil {
				logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
			}
			return
		}
		// Идем дальше если это сообщения с сервера
		switch cmds.Name {
		case Constants.CommandPing:
			// Немедленный отложенный ответ (В дискорде появляется сообщение, что бот думает, он ждет ответа)
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			if err != nil {
				logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
				return
			}

			// Асинхронно обрабатываем запрос
			go func() {
				message := "Тебе пишет " + i.Member.User.GlobalName + ": /" + Constants.CommandPing
				aiResponse, errAi := AI.Promt(i.Member.User.GlobalName, message, systemPromt, AIApi, RateLimiter, logs)
				if errAi != nil {
					logs.Warning(Error.AiMessageError+": "+errAi.Error(), logger.GetPlace())
					aiResponse = Error.AIResponseError
				}

				// 3. Отправляем результат
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: aiResponse,
				})
				// заносим событие в базу данных
				databaseMethods.DBMutex.Lock()
				databaseMethods.DBNewAction(i.Member.User.Username, cmds.Name, db, logs) // заносим событие в базу данных
				databaseMethods.DBMutex.Unlock()
				if err != nil {
					logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
					return
				}
			}()
		case Constants.CommandYou:
			// Немедленный отложенный ответ (В дискорде появляется сообщение, что бот думает, он ждет ответа)
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			if err != nil {
				logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
				return
			}

			// Асинхронно обрабатываем запрос
			go func() {
				message := "Тебе пишет " + i.Member.User.GlobalName + ": /" + Constants.CommandYou
				aiResponse, errAi := AI.Promt(i.Member.User.GlobalName, message, systemPromt, AIApi, RateLimiter, logs)
				if errAi != nil {
					logs.Warning(Error.AiMessageError+": "+errAi.Error(), logger.GetPlace())
					aiResponse = Error.AIResponseError
				}

				// 3. Отправляем результат
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: aiResponse,
				})
				// заносим событие в базу данных
				databaseMethods.DBMutex.Lock()
				databaseMethods.DBNewAction(i.Member.User.Username, cmds.Name, db, logs) // заносим событие в базу данных
				databaseMethods.DBMutex.Unlock()
				if err != nil {
					logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
					return
				}
			}()

		case Constants.CommandTime:
			// Немедленный отложенный ответ (В дискорде появляется сообщение, что бот думает, он ждет ответа)
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			if err != nil {
				logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
				return
			}

			// Асинхронно обрабатываем запрос
			go func() {
				message := "Тебе пишет " + i.Member.User.GlobalName + ": /" + Constants.CommandTime
				aiResponse, errAi := AI.Promt(i.Member.User.GlobalName, message, systemPromt, AIApi, RateLimiter, logs)
				if errAi != nil {
					logs.Warning(Error.AiMessageError+": "+errAi.Error(), logger.GetPlace())
					aiResponse = Error.AIResponseError
				}

				// 3. Отправляем результат
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: aiResponse,
				})
				// заносим событие в базу данных
				databaseMethods.DBMutex.Lock()
				databaseMethods.DBNewAction(i.Member.User.Username, cmds.Name, db, logs) // заносим событие в базу данных
				databaseMethods.DBMutex.Unlock()
				if err != nil {
					logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
					return
				}
			}()

		case Constants.CommandTalk:

			// Немедленный отложенный ответ (В дискорде появляется сообщение, что бот думает, он ждет ответа)
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			if err != nil {
				logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
				return
			}

			// Асинхронно обрабатываем запрос
			go func() {
				message := cmds.Options[0].StringValue()
				message = "Тебе пишет " + i.Member.User.GlobalName + ": " + message
				aiResponse, errAi := AI.Promt(i.Member.User.GlobalName, message, systemPromt, AIApi, RateLimiter, logs)
				if errAi != nil {
					logs.Warning(Error.AiMessageError+": "+errAi.Error(), logger.GetPlace())
					aiResponse = Error.AIResponseError
				}

				// 3. Отправляем результат
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: aiResponse,
				})
				// заносим событие в базу данных
				databaseMethods.DBMutex.Lock()
				databaseMethods.DBNewAction(i.Member.User.Username, cmds.Name+" "+message, db, logs) // заносим событие в базу данных
				databaseMethods.DBMutex.Unlock()
				if err != nil {
					logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
					return
				}
			}()
		case Constants.CommandAdd:
			// Немедленный отложенный ответ (В дискорде появляется сообщение, что бот думает, он ждет ответа)
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			if err != nil {
				logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
				return
			}
			go func() {
				discordID := cmds.Options[0].UserValue(s).Username
				steamID := cmds.Options[1].StringValue()
				// Если ошибка, то такого айди нет, или если непредвиденная ошибка, то не добавляем пользователя
				if _, errSteam := SomeApies.GetSteamInfoBySteamID(steamApi, steamID, logs); errSteam != nil {
					_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: "Ошибка добавления сопоставления, возможно такого пользователя стим нет.",
					})

					if err != nil {
						logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
						return
					}
					return
					// Если верхних ошибок нет, то идем далее
				}
				// так как выше мы проверили, что такой steam профиль есть, то внутри нижней функции не будем проверять
				errorExec := SomeApies.CreateInfoByDiscordSteamID(i.Member.User.Username, discordID, steamID, db, logs)
				if errorExec != nil {
					logs.Warning(errorExec.Error(), logger.GetPlace())
					_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: "Ошибка добавления сопоставления, возможно такой discordID уже добавлен",
					})
					if err != nil {
						logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
						return
					}
					return
				}
				// Логи уже в функции выше есть

				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Сопоставление было успешно добавлено.",
				})
				SomeApies.ComparesMutex.Lock()
				Compares[discordID] = steamID
				SomeApies.ComparesMutex.Unlock()
				// заносим событие в базу данных
				databaseMethods.DBMutex.Lock()
				databaseMethods.DBNewAction(i.Member.User.Username, cmds.Name+" "+discordID+" <=> "+steamID,
					db, logs) // заносим событие в базу данных
				databaseMethods.DBMutex.Unlock()
				if err != nil {
					logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
					return
				}
			}()
		case Constants.CommandGet:
			// Немедленный отложенный ответ (В дискорде появляется сообщение, что бот думает, он ждет ответа)
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			if err != nil {
				logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
				return
			}

			// Асинхронно обрабатываем запрос
			go func() {
				discordID := cmds.Options[0].UserValue(s).Username
				result1, errInfo1 := SomeApies.GetSteamInfoByDiscordID(Compares, steamApi, discordID, logs)
				if errInfo1 != nil {
					logs.Warning("Ошибка get steam user: "+errInfo1.Error(), logger.GetPlace())
					_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: "Не удалось найти такого пользователя. Попробуйте его добавить через команду /add",
					})
					if err != nil {
						logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
						return
					}
					return
				}
				Minutes, errInfo2 := SomeApies.GetDota2HoursByDiscordID(Compares, steamApi, discordID, logs)
				if errInfo2 != nil {
					logs.Warning("Ошибка get steam user: "+errInfo1.Error(), logger.GetPlace())
					_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: "Не удалось найти такого пользователя. Попробуйте его добавить через команду /add",
					})
					if err != nil {
						logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
						return
					}
					return
				}
				MinutesRound := strconv.Itoa(int(Minutes))
				SteamName := result1["personaname"].(string)
				ProfileUrl := result1["profileurl"].(string)
				AvatarUrl := result1["avatarfull"].(string)
				AltuhaMessage := ""
				Promt := "У меня " + fmt.Sprintf("%f", MinutesRound) + " минут наигранно в дота2. " +
					"Переведи их в часы и только оценика, побольше, построже."
				Promt = "Тебе пишет " + discordID + ": " + Promt
				aiResponse, errAi := AI.Promt(i.Member.User.GlobalName, Promt, systemPromt, AIApi, RateLimiter, logs)
				if errAi != nil {
					AltuhaMessage = "Мне трудно дать оценку этому профилю. Мне лень."
				} else {
					AltuhaMessage = aiResponse
				}
				// Делаем изображение в сообщении
				embed := &discordgo.MessageEmbed{
					Title: "Профиль " + discordID,
					Image: &discordgo.MessageEmbedImage{
						URL: AvatarUrl,
					},
					Color: 0x0099ff, // Синий цвет
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Steam Nickname",
							Value: fmt.Sprintf("```%s```", SteamName),
						},
						{
							Name:  "Profile URL",
							Value: fmt.Sprintf("```%s```", ProfileUrl),
						},
						{
							Name:  "Минут наигранных в доте 2",
							Value: fmt.Sprintf("```%s```", MinutesRound),
						},
						{
							Name:  "Мнение Альтушки об этом пользователе",
							Value: fmt.Sprintf("```%s```", AltuhaMessage),
						},
						{
							Name: "Иконка профиля",
						},
					},
				}
				// 3. Отправляем результат
				// Отправляем финальное сообщение с embed
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Embeds: []*discordgo.MessageEmbed{embed},
				})
				if err != nil {
					logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
					return
				} else {
					databaseMethods.DBMutex.Lock()
					databaseMethods.DBNewAction(i.Member.User.Username, cmds.Name+" "+discordID, db, logs) // заносим событие в базу данных
					databaseMethods.DBMutex.Unlock()
					return
				}
				// заносим событие в базу данных
			}()
		case Constants.CommandDelete:
			discordID := cmds.Options[0].UserValue(s).Username
			v, finds := Compares[discordID]
			if !finds || v == "" {
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Такого пользователя и так не было. Удаление не требуется.",
					},
				})
				if err != nil {
					logs.Warning(Error.ChannelMessageError+": "+err.Error(), logger.GetPlace())
					return
				}
				return
			}
			databaseMethods.DBMutex.Lock()
			errDelete := databaseMethods.DeleteCompare(discordID, db)
			databaseMethods.DBMutex.Unlock()
			if errDelete != nil {
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Ошибка удаления сопоставления",
					},
				})
				if err != nil {
					logs.Warning(Error.ChannelMessageError+": "+err.Error(), logger.GetPlace())
					return
				}
			}
			SomeApies.ComparesMutex.Lock()
			Compares[discordID] = ""
			SomeApies.ComparesMutex.Unlock()

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Сопоставление было успешно удалено.",
				},
			})
			if err != nil {
				logs.Warning(Error.ChannelMessageError+": "+err.Error(), logger.GetPlace())
				return
			}
		default:
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: Error.NonameErrorMessage,
				},
			})
			logs.Warning(Error.NonameError, logger.GetPlace())
			if err != nil {
				logs.Warning(Error.ChannelMessageError+"\n"+err.Error(), logger.GetPlace())
			}
		}
	})

	// Открываем соединение
	err = dg.Open()
	if err != nil {
		logs.Error(Error.SessionError+": "+err.Error(), logger.GetPlace())
		return
	}
	defer dg.Close()

	// Регистрация команд
	registeredCommands, err := dg.ApplicationCommandBulkOverwrite(dg.State.User.ID, Constants.MyServerId, commands)
	if err != nil {
		logs.Warning(Error.RegisteringCommandsError+": "+err.Error(), logger.GetPlace())
		panic(Error.RegisteringCommandsError + ": " + err.Error())
	}
	if len(registeredCommands) != cmd.BotCommands {
		logs.Warning(Error.RegisteringCommandsError+": "+"не все команды заригестрированны", logger.GetPlace())
		return
	}
	log.Println("Зарегистрированные команды:", registeredCommands)
	// Ждем сигнала завершения (Ctrl+C)
	fmt.Println("Бот работает. Ctrl+C для выхода.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
