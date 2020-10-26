package logger

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

var loglvl map[string]int

func MakeLog(logtype int, errlog interface{}) {

	if len(loglvl) == 0 {
		f, err := os.Open("./configs/dialer.json")
		json.NewDecoder(f).Decode(&loglvl)
		if err != nil {
			fmt.Println(err)
		}
		f.Close()
	}

	loglevelmap := map[int]log.Level{

		1: log.ErrorLevel,
		2: log.InfoLevel,
		3: log.DebugLevel,
	}

	logfile, err := os.OpenFile("./logs/dialer.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(loglevelmap[loglvl["loglevel"]])
	log.SetOutput(logfile)
	switch {
	case logtype == 1:
		log.Error(errlog)
	case logtype == 2:
		log.Info(errlog)
	case logtype == 3:
		log.Debug(errlog)
	}
	defer logfile.Close()

}
