package Constants

import "path/filepath"

var (
	PathToBotSystemtxt = filepath.Join("..", "..", "AI", "BotsystemPromt.txt")
	PathToDataBasetxt  = filepath.Join("..", "..", "databaseMethods", "database", "databaseAltuha.db")
)

const (
	TalksOnlyInServer = "Ни, я отвечаю только на сервере 'не придумал' и только для хороших мальчиков, "
	SessionSuccess    = "Create session success"

	MyServerId    = "537698381527777300"
	TextChannelID = "904769583540486164"
	CommandPing   = "ping"
	CommandTalk   = "talk"
	CommandYou    = "you"
	CommandTime   = "time"
	CommandAdd    = "add"
	CommandGet    = "get"
	CommandDelete = "delete"
)
