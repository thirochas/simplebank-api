package fake

import "github.com/thirochas/simplebank-golang-api/internal/util"

func Config() util.Config {
	return util.Config{
		DBDriver:      "DB_DRIVER",
		DBSource:      "DB_SOURCE",
		ServerAddress: "SERVER_ADDRESS",
		SecretKey:     "FAKE_SECRET_KEY_WITH_32_CHARS_11",
		TokenType:     "paseto",
	}
}
