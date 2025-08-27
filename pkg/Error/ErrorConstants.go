package Error

const (
	LogFileDoesNotOpen         = "Log-file does not open"
	LogFileDoesNotWrite        = "Log-file does not write logs"
	SystemPromtFileDoesNotOpen = "System-promt-file does not open"

	SessionError = "Create session error"

	ApiKeyIsEmpty   = "Api-key value in evironment is empty"
	BotTokenIsEmpty = "Bot-token value in evironment is empty"

	RegisteringCommandsError = "Registering commands error"

	ChannelMessageError = "Create channel message error"

	AiMessageError = "Ai message error"

	SessionLimit = "Session limit"

	SqliteUniqueError = "UNIQUE constraint failed"

	DatabaseError      = "Database error"
	AIResponseError    = "Не хочу тебе отвечать, динаху"
	NonameErrorMessage = "Але, фигню мне тут не пиши. А то у ног моих лежать будешь"
	NonameError        = "Непредвиденная ошибка"
)

//TODO: тут много логов таких сделать надо
