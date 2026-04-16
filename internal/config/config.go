package config

import (
	"log/slog"
	"os"
	"strings"
)

var address = ":8080"

func GetConfigValue(key string) string {
	switch strings.ToLower(key) {
	case "address":
		addr := getEnvVarValue("ADDRESS")
		if addr != "" {
			return addr
		}
		logger := slog.Default()
		logger.Debug("Environment variable ADDRESS not found, falling back to default: " + address)
		return address
	default:
		return getEnvVarValue(strings.ToUpper(key))
	}
}

func getEnvVarValue(key string) string {
	val, found := os.LookupEnv(key)
	if !found {
		logger := slog.Default()
		logger.Debug("Environment variable not found", "key", key)
	}
	return val
}
