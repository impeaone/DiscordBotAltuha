package SomeApies

import (
	"DiscordBotAltuha/databaseMethods"
	"DiscordBotAltuha/pkg/logger/logger"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"io"
	"net/http"
	"sync"
)

const Dota2AppID = 570

// GetSteamInfoBySteamID - функция взовращающая некоторую информацию о пользователе
// возвращается map[string]interface{} с параметрами
// steamid - тут все понятно что за поле
// communityvisibilitystate - параметр 1, 2 или 3. 1 - приватный профиль, 2 - только для друзей, 3 - публичный
// profilestate - параметр 0 или 1, 0 - профиль не настроен, 1 - профиль настроен
// personaname - ник пользователя
// commentpermission - 0, 1, 2. 0 - никто не может оставлять коментарии в профиль, 1 - только друзья, 2 - все
// profileurl - ссылка на профиль юзера
// поля avatar, avatarmedium, avatarfull - иконка в jpg формате
// avatarhash - какой-то хеш, не разбирался
// lastlogoff - время последнего выхода из сети. Считается от 1970года, там большое число, карочы хз
// personastate - местонохождение аккаунта, тоесть страна, регион и т.д. параметры какие-то, у меня 0, тоесть нельзя
// там шото еще есть, но мне лень писать, основное вот

func GetSteamInfoBySteamID(steamAPI, steamID string, logs *logger.Log) (map[string]interface{}, error) {
	var response = make(map[string]interface{})
	url := fmt.Sprintf(`http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?key=%s&steamids=%s`,
		steamAPI, steamID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, errBody := io.ReadAll(resp.Body)
	if errBody != nil {
		return nil, errBody
	}
	errJson := json.Unmarshal(body, &response)
	if errJson != nil {
		return nil, errJson
	}
	User := response["response"].(map[string]interface{})["players"].([]interface{})
	if len(User) == 0 {
		return nil, errors.New("no players")
	}
	UserParams := User[0].(map[string]interface{})
	return UserParams, nil
}

func GetSteamInfoByDiscordID(DiscordSteam map[string]string, steamAPI, DiscordID string, logs *logger.Log) (
	map[string]interface{}, error) {
	steamID, finds := DiscordSteam[DiscordID]
	if !finds {
		return nil, errors.New("no player data")
	}
	return GetSteamInfoBySteamID(steamAPI, steamID, logs)
}

func CreateInfoByDiscordSteamID(UserName, discordID, steamID string, db *gorm.DB, logs *logger.Log) error {
	databaseMethods.DBMutex.Lock()
	err := databaseMethods.NewUser(UserName, discordID, steamID, db, logs)
	databaseMethods.DBMutex.Unlock()
	fmt.Println(err)
	return err
	// Логи внутри функции выше уже есть
}

func GetDota2HoursBySteamID(steamAPI, steamID string) (float64, error) {
	var response = make(map[string]interface{})
	url := fmt.Sprintf(`https://api.steampowered.com/IPlayerService/GetOwnedGames/v1/?key=%s&steamid=%s&include_appinfo=true&include_played_free_games=true`,
		steamAPI, steamID)
	resp, err := http.Get(url)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	body, errBody := io.ReadAll(resp.Body)
	if errBody != nil {
		return -1, errBody
	}
	errJson := json.Unmarshal(body, &response)
	if errJson != nil {
		return -1, errJson
	}

	User := response["response"].(map[string]interface{})["games"].([]interface{})
	for _, game := range User {
		dota := game.(map[string]interface{})["appid"].(float64)
		if dota == Dota2AppID {
			hours := game.(map[string]interface{})["playtime_forever"].(float64)
			return hours, nil
		}
	}
	return 0, nil
}

func GetDota2HoursByDiscordID(DiscordSteam map[string]string, steamAPI, DiscordID string, logs *logger.Log) (
	float64, error) {
	steamID, finds := DiscordSteam[DiscordID]
	if !finds {
		return 0, errors.New("no player data")
	}
	return GetDota2HoursBySteamID(steamAPI, steamID)
}

// ComparesMutex : Все функции выше идут для одного embed сообщения от бота, там в коде нужен mutex, потому что мапа
var (
	ComparesMutex sync.Mutex //mutex для мапы Compares
)
