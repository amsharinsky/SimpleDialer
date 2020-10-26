package ami

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"

	"github.com/amsharinsky/SimpleDialer/app/internal/config"
	log "github.com/amsharinsky/SimpleDialer/app/internal/logger"
)

type AMI struct {
	AMIConn net.Conn
	V       int
}

func (a *AMI) AMIConnect() {

	var AsteriskConfig config.AsteriskConfig
	conf := AsteriskConfig.GetAsteriskConfig()
	conn, err := net.DialTimeout("tcp", conf.IP+":"+conf.Port, conf.Timeout*time.Second)
	if err != nil {
		log.MakeLog(1, err)
		return
	}
	log.MakeLog(2, "Connected to"+conf.IP+":"+conf.Port)
	fmt.Fprintf(conn, "Action: Login\r\n")
	fmt.Fprintf(conn, "UserName: "+conf.Username+"\r\n")
	fmt.Fprintf(conn, "Secret: "+conf.Password+"\r\n\r\n")
	//fmt.Fprintf(conn, "Action: Events\r\n")
	a.AMIConn = conn

}

func (a *AMI) AMIClose() {
	log.MakeLog(2, "Disconnected to"+a.AMIConn.RemoteAddr().String())
	a.AMIConn.Close()

}

func (a *AMI) MakeCall(number, timeout, exten, context string) {
	log.MakeLog(2, "Making a call to the number: "+number+" in context: "+context+",extension: "+exten)
	fmt.Fprintf(a.AMIConn, "Action: Originate\r\n")
	fmt.Fprintf(a.AMIConn, "Channel: SIP/"+number+"\r\n")
	fmt.Fprintf(a.AMIConn, "Timeout: "+timeout+"\r\n")
	fmt.Fprintf(a.AMIConn, "Context: "+context+"\r\n")
	fmt.Fprintf(a.AMIConn, "Exten: "+exten+"\r\n")
	fmt.Fprintf(a.AMIConn, "Async: yes\r\n\r\n")

}

func (a *AMI) GetCallEvent(channel, agentsChannel chan map[string]map[string]string, ws *sync.WaitGroup) {
	defer ws.Done()
	event := make(map[string][]string)
	data := make(map[string]map[string][]string)
	r := bufio.NewReader(a.AMIConn)
	ev := textproto.NewReader(r)
	for {
		event, _ = ev.ReadMIMEHeader()
		key := strings.Join(event["Event"], "")

		if key == "Hangup" || key == "OriginateResponse" || key == "QueueSummary" {
			data[key] = event

		}
		//Парсим события из АМИ,  отправляем номер абонента при финальном статусе звонка в канал
		if _, ok := data["Hangup"]; ok {
			finalResult := map[string]map[string]string{"event": {"Channel": data["Hangup"]["Channel"][0], "State": data["Hangup"]["Cause-Txt"][0]}}
			channel <- finalResult
			delete(data, "Hangup")

		} else if _, ok := data["OriginateResponse"]; ok {
			if data["OriginateResponse"]["Response"][0] == "Failure" {
				finalResult := map[string]map[string]string{"event": {"Channel": data["OriginateResponse"]["Channel"][0], "State": data["OriginateResponse"]["Response"][0]}}
				channel <- finalResult
				delete(data, "OriginateResponse")

			}
		}
		if _, ok := data["QueueSummary"]; ok {

			agentCount := map[string]map[string]string{"agentCount": {"projectid": data["QueueSummary"]["Queue"][0], "Available": data["QueueSummary"]["Available"][0]}}
			agentsChannel <- agentCount
			delete(data, "QueueSummary")

		}

	}
}

func (a *AMI) GetCountAgents(ctx context.Context, projectid string) {
	defer fmt.Println("GetCountAgent DONE", time.Now())

	for {
		select {
		case <-ctx.Done():
			return
		default:

			fmt.Fprintf(a.AMIConn, "Action: QueueSummary\r\n")
			fmt.Fprintf(a.AMIConn, "Queue: "+projectid+"\r\n\r\n")
			time.Sleep(1 * time.Second)
		}

	}

}
