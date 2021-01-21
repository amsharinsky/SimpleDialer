package db

import (
	"database/sql"
	"fmt"
	"github.com/amsharinsky/SimpleDialer/app/internal/config"
	log "github.com/amsharinsky/SimpleDialer/app/internal/logger"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
	"time"
)

type DB struct {
	Config *config.DBConfig
	DBConn *sqlx.DB
	Scheme string
}

func (a *DB) ConnectDB() error {

	cfg := a.Config.GetDBConfig()
	a.Scheme = cfg.Scheme
	var conf string
	if cfg.Driver == "pgx" {
		conf = fmt.Sprintf("host=%s user=%s password=%s database=%s  sslmode=disable",
			cfg.DBServer, cfg.Username, cfg.Password, cfg.DBName)

	}
	if cfg.Driver == "mysql" {

		conf = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&charset=utf8", cfg.Username, cfg.Password, cfg.DBServer, cfg.DBName)

	}

	conn, err := sqlx.Open(cfg.Driver, conf)

	if err != nil {
		log.MakeLog(1, err)
		return err
	}
	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetConnMaxLifetime(cfg.ConnMaxLifetime * time.Minute)
	conn.SetMaxIdleConns(cfg.MaxIdleConns)
	a.DBConn = conn
	return nil
}
func (a *DB) DBClose() {
	a.DBConn.Close()
}

type Case struct {
	Id               int
	Projectid        string `db:"project_id"`
	Casename         string `db:"case_name"`
	Used             bool
	PhoneNumber      string         `db:"phone_number"`
	Utc              sql.NullString `db:"utc"`
	Call_back        bool
	Allowed_start    string
	Allowed_stop     string
	Call_back_period string
}

func (a *DB) GetCases(dialerParams DialerParams) map[int]map[string]interface{} {

	casedata := make(map[int]map[string]interface{}, 500)
	defer a.MarkCaseAsUsed(dialerParams.ProjectId, dialerParams.CaseLimit)
	order := a.SortCases(dialerParams.Sort)
	limit := strconv.Itoa(dialerParams.CaseLimit)
	sqlStatement := "SELECT id,project_id,case_name,used,phone_number,utc,call_back,allowed_start,allowed_stop,call_back_period FROM " + a.Scheme + ".dialer_clients WHERE project_id='" + dialerParams.ProjectId + "' and used <> '1' order by " + order + " LIMIT " + limit
	rows, err := a.DBConn.Queryx(sqlStatement)
	if err != nil {
		log.MakeLog(1, err)
	}
	for rows.Next() {
		var bb Case
		rows.Scan(&bb.Id, &bb.Projectid, &bb.Casename, &bb.Used, &bb.PhoneNumber, &bb.Utc, &bb.Call_back, &bb.Allowed_start, &bb.Allowed_stop, &bb.Call_back_period)
		res := map[string]interface{}{"id": bb.Id, "project_id": bb.Projectid, "case_name": bb.Casename, "used": bb.Used, "phone_number": bb.PhoneNumber, "utc": bb.Utc, "call_back": bb.Call_back, "allowed_start": bb.Allowed_start, "allowed_stop": bb.Allowed_stop, "call_back_period": bb.Call_back}
		res["callback_count"] = 0
		res["dial_count"] = 0
		id := res["id"].(int)
		casedata[id] = res
	}

	log.MakeLog(2, "Get Cases for project"+dialerParams.ProjectId)
	log.MakeLog(3, sqlStatement)
	log.MakeLog(3, casedata)
	defer rows.Close()
	return casedata
}

func (a *DB) MarkCaseAsUsed(project_id string, limit int) {

	lim := strconv.Itoa(limit)
	sqlStatement := "UPDATE " + a.Scheme + ".dialer_clients SET used='1' where id in (select id from (select id from " + a.Scheme + ".dialer_clients where project_id='" + project_id + "' and used <> '1' order by priority DESC,id DESC LIMIT " + lim + ")tmp)"
	_, err := a.DBConn.Exec(sqlStatement)
	if err != nil {
		log.MakeLog(1, err)
	}
	log.MakeLog(3, sqlStatement)
}

type DialerParams struct {
	Id        int
	ProjectId string `db:"project_id"`
	Lines     int
	CallTime  string `db:"call_time"`
	CaseLimit int    `db:"case_limit"`
	Sort      string
	Type      string
	Exten     string
	Context   string
}

func (a *DB) GetProjectParams(projectid string) DialerParams {

	var params DialerParams
	sqlStatement := "SELECT * FROM " + a.Scheme + ".dialer_params WHERE project_id='" + projectid + "'"
	err := a.DBConn.Get(&params, sqlStatement)
	if err != nil {
		log.MakeLog(1, err)
	}
	log.MakeLog(3, sqlStatement)

	return params

}

func (a *DB) SortCases(sort string) string {

	sortType := map[int]string{

		1: "priority",
		2: "created",
		3: "utc",
	}
	sortOrder := map[int]string{

		1: "DESC",
		2: "ASC",
	}

	b := strings.Split(sort, ",")
	order := ""
	for _, value := range b {

		n := strings.Split(value, ":")
		st, _ := strconv.Atoi(n[0])
		so, _ := strconv.Atoi(n[1])
		order = order + "" + sortType[st] + " " + sortOrder[so] + ","

	}

	return strings.Trim(order, ",")

}

func (a *DB) SetDialCount(id int) {

	uuid := strconv.Itoa(id)
	sqlStatement := "UPDATE " + a.Scheme + ".dialer_stat SET dial_count=dial_count+'1' where id='" + uuid + "'"
	_, err := a.DBConn.Exec(sqlStatement)
	if err != nil {
		log.MakeLog(1, err)
	}
	log.MakeLog(3, sqlStatement)
}

func (a *DB) InsertDialCase(dialCase map[string]interface{}) {
	id := strconv.Itoa(dialCase["id"].(int))

	sqlStatement := "INSERT INTO " + a.Scheme + ".dialer_stat (id,case_name,project_id,phone_number,dial_count) VALUES" +
		"('" + id + "'," +
		"'" + dialCase["case_name"].(string) + "','" +
		"" + dialCase["project_id"].(string) + "','" +
		"" + dialCase["phone_number"].(string) + "','1')"
	_, err := a.DBConn.Exec(sqlStatement)
	if err != nil {
		log.MakeLog(1, err)
	}
	log.MakeLog(3, sqlStatement)
}

func (a *DB) SetStatusCall(id int, status string) {
	uuid := strconv.Itoa(id)
	sqlStatement := "UPDATE " + a.Scheme + ".dialer_stat SET state='" + status + "',ended=now() where id='" + uuid + "'"
	_, err := a.DBConn.Exec(sqlStatement)
	if err != nil {
		log.MakeLog(1, err)
	}
	log.MakeLog(3, sqlStatement)

}
func (a *DB) SetTimeEndCall(id int) {

	uuid := strconv.Itoa(id)
	sqlStatement := "UPDATE " + a.Scheme + ".dialer_stat SET ended=now() where id='" + uuid + "'"
	_, err := a.DBConn.Exec(sqlStatement)
	if err != nil {
		log.MakeLog(1, err)
	}
	log.MakeLog(3, sqlStatement)
}

func (a *DB) GetCountCases(projectid string) int {

	var countCase int
	sqlStatement := "SELECT count(id) FROM " + a.Scheme + ".dialer_clients WHERE project_id='" + projectid + "' and used <> '1'"
	err := a.DBConn.Get(&countCase, sqlStatement)
	if err != nil {
		log.MakeLog(1, err)
	}
	log.MakeLog(3, sqlStatement)
	return countCase

}

type DeferredCases struct {
	Id           int
	Casename     string         `db:"case_name"`
	PhoneNumber  string         `db:"phone_number"`
	Utc          sql.NullString `db:"utc"`
	DeferredTime string
}

func (a *DB) GetDeferredCases(dialerParams DialerParams) map[int]map[string]interface{} {

	casedata := make(map[int]map[string]interface{}, 500)
	order := a.SortCases(dialerParams.Sort)
	limit := strconv.Itoa(dialerParams.CaseLimit)
	sqlStatement := "SELECT id,case_name,phone_number,utc,deferred_time FROM " + a.Scheme + ".dialer_clients WHERE project_id='" + dialerParams.ProjectId + "' and deferred_time is NOT NULL and deferred_time !='' and deferred_done = false  order by " + order + " LIMIT " + limit
	rows, err := a.DBConn.Queryx(sqlStatement)
	if err != nil {
		log.MakeLog(1, err)
	}
	for rows.Next() {
		var bb DeferredCases
		rows.Scan(&bb.Id, &bb.Casename, &bb.PhoneNumber, &bb.Utc, &bb.DeferredTime)
		res := map[string]interface{}{"id": bb.Id, "case_name": bb.Casename, "phone_number": bb.PhoneNumber, "utc": bb.Utc, "deferredTime": bb.DeferredTime}
		res["callback_count"] = 0
		res["dial_count"] = 0
		id := res["id"].(int)
		casedata[id] = res
	}

	log.MakeLog(2, "Get Deferred Cases for project "+dialerParams.ProjectId)
	log.MakeLog(3, sqlStatement)
	log.MakeLog(3, casedata)
	defer rows.Close()
	return casedata

}

func (a *DB) SetDeferredDone(id int) {

	uuid := strconv.Itoa(id)
	sqlStatement := "UPDATE " + a.Scheme + ".dialer_clients SET deferred_done=true where id='" + uuid + "'"
	_, err := a.DBConn.Exec(sqlStatement)
	if err != nil {
		log.MakeLog(1, err)
	}
	log.MakeLog(3, sqlStatement)
}
