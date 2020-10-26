package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/amsharinsky/SimpleDialer/app/internal/ami"
	"github.com/amsharinsky/SimpleDialer/app/internal/config"
	"github.com/amsharinsky/SimpleDialer/app/internal/db"
	"github.com/amsharinsky/SimpleDialer/app/internal/dialmodule"

	"net/http"
	"regexp"

	log "github.com/amsharinsky/SimpleDialer/app/internal/logger"
)

type DialerParams struct {
	Action        string
	ProjectId     string
	Ch            chan map[string]string
	Sql           db.DB
	Ami           ami.AMI
	Channel       chan map[string]map[string]string
	Type          string
	AgentsChannel chan map[string]map[string]string
}

func (a *DialerParams) middleware(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "GET" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			query := r.URL.Query()
			a.Action = query.Get("action")
			a.ProjectId = query.Get("projectid")
			if a.Action == "" || a.ProjectId == "" {
				err2 := errors.New("the name of one of the parameters is not correct")
				fmt.Fprintln(w, err2)
				log.MakeLog(1, err2)
				return
			}

			matchedAction, err := regexp.MatchString(`start|stop|pause`, a.Action)
			if err != nil {
				log.MakeLog(1, err)
				return
			}
			if !matchedAction {
				err2 := errors.New("not valid command")
				fmt.Fprintln(w, err2)
				log.MakeLog(1, err2)
				return
			}
			next(w, r)
		} else {
			w.WriteHeader(404)
			return
		}
	}

}
func (a *DialerParams) handler(w http.ResponseWriter, r *http.Request) {

	command := map[string]string{
		"action":    a.Action,
		"projectid": a.ProjectId,
	}
	switch a.Action {
	case "start":
		dialer := dialmodule.New()
		go dialer.Start(&a.Sql, &a.Ami, a.Ch, a.ProjectId, a.Channel, a.AgentsChannel)
	case "stop":
		a.Ch <- command
	case "pause":
		a.Ch <- command

	}

}

func main() {
	defer recoverAll()
	ws := &sync.WaitGroup{}
	var a DialerParams
	go func() {
		sigchan := make(chan os.Signal)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		fmt.Println("Program killed !")
		a.Ami.AMIClose()
		a.Sql.DBClose()
		close(a.AgentsChannel)
		close(a.Channel)
		close(a.Ch)
		// do last actions and wait for all write operations to end
		os.Exit(0)
	}()
	a.Sql.ConnectDB()
	a.Ami.AMIConnect()
	a.AgentsChannel = make(chan map[string]map[string]string)
	a.Ch = make(chan map[string]string)
	a.Channel = make(chan map[string]map[string]string)
	ws.Add(1)
	go a.Ami.GetCallEvent(a.Channel, a.AgentsChannel, ws)
	http.HandleFunc("/", a.middleware(a.handler))
	var cfg config.HttpConfig
	IP, PORT := cfg.GetHTTPConfig()
	log.MakeLog(1, http.ListenAndServe(""+IP+""+":"+""+PORT+"", nil))
	ws.Wait()
}
func recoverAll() {
	if err := recover(); err != nil {
		log.MakeLog(3, err)
	}
}
