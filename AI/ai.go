package AI

import (
	"DiscordBotAltuha/cmd"
	"DiscordBotAltuha/pkg/Error"
	"DiscordBotAltuha/pkg/logger/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Promt - функция для общения с ии. Мы передаем промт и получаем ответ. Включает параметы:
// user - тот кто пишет
// promt - сообщение для ии
// sysPromt - системное сообщение для бота
// api - апи ключ для бота
func Promt(user, promt, sysPromt, api string, ratelimiter *cmd.SimpleRateLimiter, logs *logger.Log) (string, error) {
	_, ok := ratelimiter.CheckLimit()
	if !ok {
		logs.Warning(user+" спамит боту!", logger.GetPlace())
		return user + " не нужно так быстро писать! Я не скоростная.", nil
	}
	ratelimiter.Unlock(user)
	UserPromt := promt
	var response map[string]interface{}
	url := "https://api.intelligence.io.solutions/api/v1/chat/completions"

	// Создаем тело запроса (пример)
	payload := strings.NewReader(fmt.Sprintf(`{
		"model": "meta-llama/Llama-3.3-70B-Instruct",
		"messages": [
			{"role": "system", "content": "%s"},
			{"role": "user", "content": "%s"}
		],
		"temperature": 0.7,
		"max_tokens": 500
	}`, sysPromt, UserPromt))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		logs.Error("Error creating request: "+err.Error(), logger.GetPlace())
		return "error", err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+api)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.Error("Error creating request: "+err.Error(), logger.GetPlace())
		return "error", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logs.Error("Error reading response: "+err.Error(), logger.GetPlace())
		return "error", err
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		logs.Error("Ошибка чтения тела json ответа: "+err.Error(), logger.GetPlace())
		return "Не хочу тебе отвечать, динаху", err
	}
	if response["choices"] == nil {
		logs.Warning(user+" бота промтом ломает: ", logger.GetPlace())
		return user + ", le le le динаху, клоун", nil
	}
	content := response["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	logs.Info(user+" написал боту: "+promt, logger.GetPlace())
	return content, nil
}

func GetSystemPromt(path string, logs *logger.Log) (string, error) {
	file, errFile := os.ReadFile(path)
	if errFile != nil {
		logs.Error(Error.SystemPromtFileDoesNotOpen+"\n"+errFile.Error(), logger.GetPlace())
		return "", errFile
	}
	return string(file), nil
}
