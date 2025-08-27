package databaseMethods

import (
	"DiscordBotAltuha/pkg/Error"
	"DiscordBotAltuha/pkg/logger/logger"
	"errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"
	"sync"
	"time"
)

var (
	DBMutex sync.Mutex
)

// BotUsage - таблица использования бота
type BotUsage struct {
	gorm.Model
	UserGlobalName string    `gorm:"type:varchar(255);not null"`
	Command        string    `gorm:"type:varchar(255);not null"`
	DateUsage      time.Time `gorm:"type:datetime;not null"`
}

// DiscordSteamID - таблица в которой будет сопоставление discordID и steamID
type DiscordSteamID struct {
	gorm.Model
	DiscordID string `gorm:"type:varchar(255);not null;unique"`
	SteamID   string `gorm:"type:varchar(255);not null"`
}

// OpenDatabase - подключение к базе данных
func OpenDatabase(dbPath string, log *logger.Log) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Error("ошибка базы данных", logger.GetPlace())
		return nil, err
	}
	if !db.Migrator().HasTable(&BotUsage{}) {
		err = db.AutoMigrate(&BotUsage{})
		if err != nil {
			log.Error("Ошибка создания таблицы BotUsage", logger.GetPlace())
			return nil, err
		}
	}
	if !db.Migrator().HasTable(&DiscordSteamID{}) {
		err = db.AutoMigrate(&DiscordSteamID{})
		if err != nil {
			log.Error("Ошибка создания таблицы DiscordSteamID", logger.GetPlace())
			return nil, err
		}
	}
	return db, nil
}

func DBNewAction(User, Message string, db *gorm.DB, logs *logger.Log) {
	db.Create(&BotUsage{
		UserGlobalName: User,
		Command:        Message,
		DateUsage:      time.Now(),
	})
	logs.Info(User+" использовал команду </"+Message+">", logger.GetPlace())
}

func NewUser(User, discordID, steamID string, db *gorm.DB, logs *logger.Log) error {
	execute := db.Create(&DiscordSteamID{
		DiscordID: discordID,
		SteamID:   steamID,
	})
	if execute.Error != nil {
		if strings.Contains(execute.Error.Error(), Error.SqliteUniqueError) {
			return errors.New(Error.SqliteUniqueError)
		}
		return execute.Error
	}
	logs.Info(User+" добавил сопоставление: "+discordID+" <=> "+steamID, logger.GetPlace())
	return nil
}

func GetDiscordSteamID(db *gorm.DB) (map[string]string, error) {
	var id []DiscordSteamID
	execute := db.Find(&id)
	if execute.Error != nil {
		return nil, execute.Error
	}
	ids := map[string]string{}
	for _, v := range id {
		ids[v.DiscordID] = v.SteamID
	}
	return ids, nil
}

func DeleteCompare(discordId string, db *gorm.DB) error {
	execute := db.Where("discord_id = ?", discordId).Delete(&DiscordSteamID{})
	return execute.Error
}
