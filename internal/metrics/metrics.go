package metrics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"monitw/internal/auth"
	"monitw/internal/crypto"
	"monitw/internal/db"
	"monitw/pkg/crud"
	"monitw/pkg/mutils"
	"monitw/pkg/mwutils"
	"net/http"
	"strconv"

	log "monitw/internal/logger"

	tbl "monitw/pkg/tables"
)

type MetricsSel struct {
	Id                  int64 `json:"Id,omitempty"`
	Hid                 int64 `json:"Hid,omitempty"`
	Hostname            string
	Port                int16
	Tid                 int64 `json:"Tid,omitempty"`
	Tmetric             string
	Action              string
	Descr               string
	Warning             float64
	Error               float64
	Dbsid               string
	Username            string
	Password            string
	Startm              bool
	Intrvlnotcrucialhrs int16
	Mname               string
	Tresholdismin       bool
}

func ViewMetrics(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "OPTIONS")
		return
	}

	if r.Method != "GET" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(405), 405)
		log.PrintlnWithHttp(r, mwutils.FuncName(), http.StatusText(405))
		return
	}
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft GET check")
	rps, ok := r.URL.Query()["rp"] //rows per page
	// Query()["rp"] will return an array of items,
	// we only want the single item.
	rp, _ := strconv.Atoi(rps[0])
	if !ok || len(rps[0]) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'rp' is missing")
		return
	}

	pns, ok := r.URL.Query()["pn"] //page number
	pn, _ := strconv.Atoi(pns[0])
	if !ok || len(pns[0]) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'pn' is missing")
		return
	}

	shs, ok := r.URL.Query()["sh"] //search string
	var sh string = ""
	log.PrintlnWithHttp(r, mwutils.FuncName(), "shs")
	if len(shs) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'sh' is missing")
	} else {
		sh = shs[0]
	}
	sts, ok := r.URL.Query()["st"] //search string
	var st string = ""
	log.PrintlnWithHttp(r, mwutils.FuncName(), "sts")
	if len(sts) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'st' is missing")
	} else {
		st = sts[0]
	}
	sas, ok := r.URL.Query()["sa"] //search string
	var sa string = ""
	log.PrintlnWithHttp(r, mwutils.FuncName(), "sas")
	if len(sas) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'sa' is missing")
	} else {
		sa = sas[0]
	}

	var startRow, finishRow int
	startRow = (pn-1)*rp + 1
	finishRow = startRow + rp - 1
	var sqlHostWhere string
	if sh != "" {
		sqlHostWhere = " and h.hostname like '%" + sh + "%'"
	}

	sqlQuery := fmt.Sprintf(`SELECT m.Id, m.hid, m.Hostname, Port, t.id, t.Tname, Action, m.Descr, Warning, Error,	
								Dbsid, Username, Password, Startm, Intrvlnotcrucialhrs, mname, tresholdismin from 
							(SELECT row_number() over (order by m.Id), m.Id, h.id hid, h.Hostname, Port, Tmetric, Action, m.Descr, Warning, Error,	
								coalesce(Dbsid,'') Dbsid, Username, Password, Startm, Intrvlnotcrucialhrs, mname,tresholdismin from monit_sch.metrics m
								join monit_sch.vhostnames h on h.id=m.Hostname  %s
								) m
								join monit_sch.tmetrics t on t.id=m.Tmetric
								
							where row_number between %d and %d`, sqlHostWhere, startRow, finishRow)

	if st != "" {
		sqlQuery = sqlQuery + " and t.tname like '%" + st + "%'"
	}
	if sa != "" {
		sqlQuery = sqlQuery + " and action like '%" + sa + "%'"
	}

	rows, err := db.Db.Query(sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error:"+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	log.PrintlnWithHttp(r, mwutils.FuncName(), "after select metrics")

	metrics := make([]*MetricsSel, 0)
	for rows.Next() {

		m := new(MetricsSel)
		err := rows.Scan(&m.Id, &m.Hid, &m.Hostname, &m.Port, &m.Tid, &m.Tmetric, &m.Action, &m.Descr, &m.Warning, &m.Error,
			&m.Dbsid, &m.Username, &m.Password, &m.Startm, &m.Intrvlnotcrucialhrs, &m.Mname, &m.Tresholdismin)
		m.Password = crypto.Decrypt(mutils.Key, m.Password)
		//log.PrintlnWithHttp(r, "m.Password decr", m.Password)
		if err != nil {
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			log.PrintlnWithHttp(r, mwutils.FuncName(), err.Error())
			return
		}
		metrics = append(metrics, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	metricsJson, err := json.Marshal(metrics)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", metricsJson)
}

func CreateMetric(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	ctx := r.Context()
	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "OPTIONS")
		return
	}

	if r.Method != "POST" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(405), 405)
		log.ErrorWithHttp(r, mwutils.FuncName(), http.StatusText(405))
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	log.PrintlnWithHttp(r, mwutils.FuncName(), string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())

	}

	var metricsSel MetricsSel
	err = json.Unmarshal(b, &metricsSel)

	var metrics tbl.Metrics
	metrics.Id = metricsSel.Id
	metrics.Hostname = metricsSel.Hid
	metrics.Port = metricsSel.Port
	metrics.Tmetric = metricsSel.Tid
	metrics.Action = metricsSel.Action
	metrics.Descr = metricsSel.Descr
	metrics.Warning = metricsSel.Warning
	metrics.Error = metricsSel.Error
	metrics.Dbsid = metricsSel.Dbsid
	metrics.Username = metricsSel.Username
	metrics.Password = crypto.Encrypt(mutils.Key, metricsSel.Password)
	metrics.Startm = metricsSel.Startm
	metrics.Intrvlnotcrucialhrs = metricsSel.Intrvlnotcrucialhrs
	metrics.Mname = metricsSel.Mname
	metrics.Tresholdismin = metricsSel.Tresholdismin

	//fmt.Println(metrics)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	log.PrintlnWithHttp(r, mwutils.FuncName(), crud.CreateInsert(metrics, auth.GetUserID(ctx)))

	txn, _ := db.Db.Begin()

	var lastInsertId int64

	insq := crud.CreateInsert(metrics, auth.GetUserID(ctx))
	log.PrintlnWithHttp(r, mwutils.FuncName(), "insq "+insq)

	err = db.Db.QueryRow(insq).Scan(&lastInsertId)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(500)+" "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		txn.Rollback()
		return
	}

	err = txn.Commit()
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft insert commit")

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "%d", lastInsertId)

}

func UpdateMetric(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")
	ctx := r.Context()

	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "OPTIONS")
		return
	}

	if r.Method != "POST" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(405), 405)
		log.ErrorWithHttp(r, mwutils.FuncName(), http.StatusText(405))
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	log.PrintlnWithHttp(r, mwutils.FuncName(), string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
	}

	var metricsSel MetricsSel
	err = json.Unmarshal(b, &metricsSel)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	var metrics tbl.Metrics
	metrics.Id = metricsSel.Id
	metrics.Hostname = metricsSel.Hid
	metrics.Port = metricsSel.Port
	metrics.Tmetric = metricsSel.Tid
	metrics.Action = metricsSel.Action
	metrics.Descr = metricsSel.Descr
	metrics.Warning = metricsSel.Warning
	metrics.Error = metricsSel.Error
	metrics.Dbsid = metricsSel.Dbsid
	metrics.Username = metricsSel.Username
	metrics.Password = crypto.Encrypt(mutils.Key, metricsSel.Password)
	metrics.Startm = metricsSel.Startm
	metrics.Intrvlnotcrucialhrs = metricsSel.Intrvlnotcrucialhrs
	metrics.Mname = metricsSel.Mname
	metrics.Tresholdismin = metricsSel.Tresholdismin

	txn, _ := db.Db.Begin()

	var lastUpdateId int64
	updq := crud.CreateUpdate(metrics, auth.GetUserID(ctx))
	log.PrintlnWithHttp(r, "updq", updq)

	err = db.Db.QueryRow(updq).Scan(&lastUpdateId)

	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft update")
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(500)+" "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		txn.Rollback()
		return
	}

	err = txn.Commit()

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "%d", lastUpdateId)

}

func DeleteMetric(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	ctx := r.Context()

	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "OPTIONS")
		return
	}

	if r.Method != "POST" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(405), 405)
		log.ErrorWithHttp(r, mwutils.FuncName(), http.StatusText(405))
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	log.PrintlnWithHttp(r, mwutils.FuncName(), string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
	}

	var metric tbl.Metrics
	err = json.Unmarshal(b, &metric)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	mId := strconv.FormatInt(metric.Id, 10)

	delq := crud.CreateDelete("Metrics", mId, auth.GetUserID(ctx))

	txn, _ := db.Db.Begin()

	result, err := txn.Exec(delq)
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft delete ")
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(500)+" "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		txn.Rollback()
		return
	}
	err = txn.Commit()
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft Commit")

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "Metric deleted successfully (%d row affected)\n", rowsAffected)

}

func ViewMetric(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	//log.PrintlnWithHttp(r, "req", r)

	mwutils.SetupResponse(&w, r)

	if r.Method == "OPTIONS" {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "OPTIONS")
		return
	}

	if r.Method != "GET" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(405), 405)
		log.ErrorWithHttp(r, mwutils.FuncName(), http.StatusText(405))
		return
	}
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft GET check")

	ids, ok := r.URL.Query()["id"] //rows per page
	id, _ := strconv.Atoi(ids[0])
	if !ok || len(ids[0]) < 1 {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "Url Param 'id' is missing")
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'id' is missing")
		return
	}

	m := new(MetricsSel)
	err := db.Db.QueryRow(`SELECT m.Id, h.id, h.Hostname, 
								Port, t.id, t.Tname, Action, m.Descr, Warning, Error,	
								coalesce(Dbsid,'') Dbsid, Username, Password, Startm, Intrvlnotcrucialhrs, mname, Tresholdismin from monit_sch.metrics m
								join monit_sch.tmetrics t on t.id=m.Tmetric
								join monit_sch.vhostnames h on h.id=m.Hostname
							where m.id = $1`, id).Scan(&m.Id, &m.Hid, &m.Hostname, &m.Port,
		&m.Tid, &m.Tmetric, &m.Action, &m.Descr, &m.Warning,
		&m.Error, &m.Dbsid, &m.Username, &m.Password, &m.Startm, &m.Intrvlnotcrucialhrs, &m.Mname, &m.Tresholdismin)
	m.Password = crypto.Decrypt(mutils.Key, m.Password)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	TmetricJson, err := json.Marshal(m)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", TmetricJson)
}

//cmnt
