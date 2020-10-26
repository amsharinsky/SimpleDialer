package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/amsharinsky/SimpleDialer/app/internal/logger"
)

type HttpConfig struct {
	IP   string `json:"ip"`
	PORT string `json:"port"`
}

func (a *HttpConfig) GetHTTPConfig() (IP, PORT string) {

	f, err := os.Open("./configs/http.json")
	if err != nil {
		logger.MakeLog(1, err)
	}
	json.NewDecoder(f).Decode(&a)
	return a.IP, a.PORT
}

type DBConfig struct {
	Driver          string        `json:"driver"`
	DBServer        string        `json:"db_server"`
	Port            int           `json:"port"`
	DBName          string        `json:"db_name"`
	Scheme          string        `json:"scheme"`
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
}

func (b *DBConfig) GetDBConfig() *DBConfig {

	f, err := os.Open("./configs/dbconfig.json")
	if err != nil {
		logger.MakeLog(1, err)
	}
	json.NewDecoder(f).Decode(&b)
	return b
}

type AsteriskConfig struct {
	IP       string        `json:"ip"`
	Port     string        `json:"port"`
	Timeout  time.Duration `json:"timeout"`
	Username string        `json:"username"`
	Password string        `json:"password"`
}

func (c *AsteriskConfig) GetAsteriskConfig() *AsteriskConfig {

	f, err := os.Open("./configs/asterisk.json")
	if err != nil {
		logger.MakeLog(1, err)
	}
	json.NewDecoder(f).Decode(&c)
	return c

}
