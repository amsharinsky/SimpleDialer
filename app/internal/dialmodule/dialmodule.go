package dialmodule

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/amsharinsky/SimpleDialer/app/internal/logger"

	"github.com/amsharinsky/SimpleDialer/app/internal/ami"
	"github.com/amsharinsky/SimpleDialer/app/internal/db"
)

type DialModule struct {
	dialerParams  *db.DialerParams
	countLine     int
	channel       chan map[string]map[string]string
	callbackCases map[int]map[string]interface{}
	callResult    map[string]map[string]string
	Ami           *ami.AMI
	countAgents   map[string]map[string]string
	agentsChannel chan map[string]map[string]string
	Check         bool
	projectid     string
	sql           *db.DB
	result        string
	ch            chan map[string]string
	countCase     int
	ticker        *time.Ticker
	defferedCases map[int]map[string]interface{}
}

func New() *DialModule {

	return &DialModule{}
}

func (a *DialModule) Start(sql *db.DB, ami *ami.AMI, ch chan map[string]string, projectid string, channel, agentsChannel chan map[string]map[string]string) {
	a.sql = sql
	a.Ami = ami
	a.ch = ch
	ctx, cancel := context.WithCancel(context.Background())
	dialerParams := a.sql.GetProjectParams(projectid)
	a.channel = channel
	a.agentsChannel = agentsChannel
	a.callbackCases = make(map[int]map[string]interface{})
	a.dialerParams = &dialerParams
	a.projectid = projectid
	a.countCase = sql.GetCountCases(projectid)
	if a.countCase <= 0 {
		loginfo := "No cases for call"
		log.MakeLog(2, loginfo)
		return
	}
	//a.defferedCases = a.sql.GetDifferedCases(*a.dialerParams)
	cases := sql.GetCases(dialerParams)
	fmt.Println(cases)
	if dialerParams.Type == "progressive" {
		go ami.GetCountAgents(ctx, projectid)
	}
	//Если есть кейсы для обзвона
	go a.dial(cases)
	a.manager()
	cancel()
}

func (a *DialModule) dial(cases map[int]map[string]interface{}) {
	defer recoverAll()
	var res bool
	var caseTime string
	for {
		select {
		case a.callResult = <-a.channel:

		case a.countAgents = <-a.agentsChannel:
		default:
		}
		if len(cases) > 0 {
			for key, value := range cases {
				if len(a.callResult) > 0 {
					resev := a.callResult["event"]["Channel"]
					res, _ = regexp.Match(value["phone_number"].(string), []byte(resev))
					if res {
						//Если была у кейса установлен статус перезвон и была одна попытка дозвона и не было ответа абонента, то перекидываем в мапу перезвона
						if cases[key]["call_back"].(bool) && cases[key]["dial_count"].(int) == 1 && a.callResult["event"]["State"] != "Normal Clearing" {
							a.callbackCases[key] = value
						} else {
							a.sql.SetStatusCall(key, a.callResult["event"]["State"])
						}
						delete(a.callResult, "event")
						delete(cases, key)
						//Освобождаем линию
						a.countLine--

					}
				}
				// Проверяем существует ли дааный кейс,так может быть удален и код упадет в ошибку
				if _, ok := cases[key]; ok {
					timeNow := getTime("", 0, false)
					//Проверяем ,если utc указан то получаем время по utc иначе берем локальное время с сервера
					if cases[key]["utc"].(sql.NullString).Valid {
						caseTime = getTime(cases[key]["utc"].(sql.NullString).String, 0, false)
					} else {
						caseTime = timeNow
					}
					allowStart := cases[key]["allowed_start"].(string)
					allowStop := cases[key]["allowed_stop"].(string)
					if caseTime <= timeNow {
						//Проверяем попадает ли кейс в разрешенное время обзвона
						if allowStart <= timeNow && timeNow <= allowStop || allowStart == "00:00:00" && allowStop == "00:00:00" {
							//Если попало то смотрим режим и делаем проверки согласно режиму
							if a.dialerParams.Type != "autoinformator" {
								av := a.countAgents["agentCount"]["Available"]
								avail, _ := strconv.Atoi(av)
								a.Check = a.countLine < a.dialerParams.Lines && !cases[key]["used"].(bool) && avail != 0 && a.countLine <= avail
							}
							if a.dialerParams.Type == "autoinformator" {
								a.Check = a.countLine < a.dialerParams.Lines && !cases[key]["used"].(bool)
							}
							//Если проверка параметров прошла то звоним
							if a.Check {
								fmt.Println("dial", value["phone_number"].(string))

								//Занимаем линию
								a.countLine++
								a.sql.InsertDialCase(cases[key])
								cases[key]["dial_count"] = cases[key]["dial_count"].(int) + 1
								a.Ami.MakeCall(value["phone_number"].(string), a.dialerParams.CallTime, a.dialerParams.Exten, a.dialerParams.Context)
								cases[key]["used"] = true

							}

						}
					}
				}
			}

		}
		if len(cases) <= 0 && a.countCase > 0 {
			cases = a.sql.GetCases(*a.dialerParams)
			a.countCase = a.countCase - 1
		}
		if len(a.callbackCases) > 0 {
			a.callback()
		}
		if len(cases) <= 0 && a.countCase <= 0 && len(a.callbackCases) <= 0 {
			log.MakeLog(2, "Cases are over.The command to stop calling has been sent")
			a.ch <- map[string]string{
				"action":    "stop",
				"projectid": a.projectid,
			}
			return
		}
		//Вызов функции перезвона кейсов
		time.Sleep(1 * time.Millisecond)
	}

}

func (a *DialModule) callback() {
	var res bool
	var caseTime string
	var callbackMinute string
	var callbackPeriod []string

	for key, value := range a.callbackCases {

		if _, ok := a.callbackCases[key]; ok {
			if a.callbackCases[key]["callbackTime"] == nil {
				// Если период перезвона больше чем один раз
				if len(a.callbackCases[key]["callback_period"].(string)) > 1 {
					callbackPeriod := strings.Split(a.callbackCases[key]["call_back_period"].(string), ",")
					callbackMinute = callbackPeriod[0]
				} else {
					callbackMinute = a.callbackCases[key]["callback_period"].(string)
				}
				rMinute, _ := strconv.Atoi(callbackMinute)
				callbackTime := getTime("+03.00h", time.Duration(rMinute), false)
				a.callbackCases[key]["callbackTime"] = callbackTime
				fmt.Println("Set First call back", a.callbackCases)
			}
			timeNow := getTime("", 0, false)
			//Смотрим utc,если не указано в кейсе то берем текущее время сервера
			if a.callbackCases[key]["utc"].(sql.NullString).Valid {
				caseTime = getTime(a.callbackCases[key]["utc"].(sql.NullString).String, 0, false)
			} else {
				caseTime = timeNow
			}
			//Начинаем звонить
			if a.callbackCases[key]["callbackTime"] != nil {
				if caseTime <= timeNow {
					allowStart := a.callbackCases[key]["allowed_start"].(string)
					allowStop := a.callbackCases[key]["allowed_stop"].(string)
					if allowStart <= timeNow && timeNow <= allowStop && a.callbackCases[key]["callbackTime"].(string) <= timeNow || allowStart == "00:00:00" && allowStop == "00:00:00" {
						//Звоним если есть свободное кол-во линий
						if a.dialerParams.Type != "autoinformator" {
							avail, _ := strconv.Atoi(a.countAgents["agentCount"]["Available"])
							a.Check = a.countLine < a.dialerParams.Lines && avail != 0 && a.countLine <= avail && a.callbackCases[key]["callback_count"].(int) < 1
						} else {
							a.Check = a.countLine < a.dialerParams.Lines && a.callbackCases[key]["callback_count"].(int) < 1
						}
						if a.Check {
							//Занимаем линию
							a.callbackCases[key]["dial_count"] = a.callbackCases[key]["dial_count"].(int) + 1
							a.callbackCases[key]["callback_count"] = a.callbackCases[key]["callback_count"].(int) + 1
							a.countLine++
							fmt.Println("call back", value["phone_number"].(string))
							a.sql.SetDialCount(key)
							a.Ami.MakeCall(value["phone_number"].(string), a.dialerParams.CallTime, a.dialerParams.Exten, a.dialerParams.Context)

						}

					}
				}
			}
		}
		if len(a.callResult) > 0 {
			res, _ = regexp.Match(value["phone_number"].(string), []byte(a.callResult["event"]["Channel"]))
			if res {
				//Если ответ корректный то удаляем кейс из пула перезвонов
				if a.callResult["event"]["State"] == "Normal Clearing" {
					fmt.Println("Answer")
					a.sql.SetStatusCall(key, a.callResult["event"]["State"])
					delete(a.callbackCases, key)
				}
				//Если не корректный ответ то устанавливаем время перезвона
				if a.callResult["event"]["State"] != "Normal Clearing" {

					callbackPeriod = strings.Split(a.callbackCases[key]["callback_period"].(string), ",")
					if int64(len(callbackPeriod)) > 1 && len(callbackPeriod) >= a.callbackCases[key]["dial_count"].(int) {
						i := a.callbackCases[key]["dial_count"].(int) - 1
						callbackMinute = callbackPeriod[i]
						rMinute, _ := strconv.Atoi(callbackMinute)
						recallTime := getTime("+03.00h", time.Duration(rMinute), false)
						a.callbackCases[key]["callbackTime"] = recallTime

					}
					if _, ok := a.callbackCases[key]["callback_count"]; ok {
						if a.callbackCases[key]["callback_count"].(int) > 0 {
							a.callbackCases[key]["callback_count"] = a.callbackCases[key]["callback_count"].(int) - 1
						}
					}
				}
				//Проверяем,если перезвон уже раз был то уменьшаем на единицу.Чтобы не было нескольких перезвонов по одному номеру сразу
				a.countLine--
				//Если выполнено перезвонов ,равное числу перезвонов в кейсе то удаляем кейс из пула перезвонов
				if _, ok := a.callbackCases[key]; ok {
					if a.callbackCases[key]["dial_count"].(int)-1 >= len(callbackPeriod) {
						fmt.Println("delete ", key)
						a.sql.SetStatusCall(key, a.callResult["event"]["State"])
						delete(a.callbackCases, key)
					}
				}
				delete(a.callResult, "event")
			}
		}
	}
	if len(a.callbackCases) <= 0 {
		return
	}
}

func getTime(utc string, recallMinute time.Duration, defferedCall bool) string {

	var retTime string
	var format string
	if defferedCall {
		format = "2006.01.02 15:04:05"
	} else {
		format = "15:04:05"
	}
	if utc != "" {
		h, _ := time.ParseDuration(utc)
		retTime = time.Now().UTC().Add(h).Add(recallMinute * time.Minute).Format(format)
	} else {
		retTime = time.Now().Format(format)

	}

	return retTime

}

func (a *DialModule) tickers() map[int]map[string]interface{} {

	a.ticker = time.NewTicker(time.Second * 30)
	select {
	case <-a.ticker.C:
		a.defferedCases = a.sql.GetDifferedCases(*a.dialerParams)
		if len(a.defferedCases) > 0 {
			return a.defferedCases
		}
	default:
	}
	return nil
}

func (a *DialModule) manager() {

	for {
		select {
		case command := <-a.ch:
			fmt.Println(command)
			if command["action"] == "stop" && command["projectid"] == a.projectid {
				a.result = "stop"
				return
			}
		default:
		}
	}

}
func recoverAll() {
	if err := recover(); err != nil {
		log.MakeLog(1, err)

	}
}
