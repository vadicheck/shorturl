package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags_DefaultValues(t *testing.T) {
	os.Clearenv()

	os.Args = []string{"cmd"}

	ParseFlags()

	assert.Equal(t, "prod", Config.AppEnv)
	assert.Equal(t, "localhost:8080", Config.ServerAddress)
	assert.Equal(t, "http://localhost:8080", Config.BaseURL)
	assert.Equal(t, "", Config.DatabaseDsn)
	assert.Equal(t, "./storage/filestorage.txt", Config.FileStoragePath)
	assert.Equal(t, "secretkey", Config.JwtSecret)
	assert.Equal(t, time.Hour*24, Config.JwtTokenExpire)
	assert.Equal(t, "very-secret", Config.SecureCookieHashKey)
	assert.Equal(t, "alotsecretalotsecretalotsecretgr", Config.SecureCookieBlockKey)
	assert.Equal(t, time.Hour*24, Config.SecureCookieExpire)
}
