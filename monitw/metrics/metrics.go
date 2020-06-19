package metrics

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"monitw/crypto"
	"monitw/mwutils"
	"net/http"
	"strconv"
)

type Metric struct {
	Id                  int16 `json:"Id,omitempty"`
	Hid                 int16 `json:"Hid,omitempty"`
	Hostname            string
	Port                int16
	Tid                 int16 `json:"Tid,omitempty"`
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
}

var Logger *log.Logger

func ViewMetrics(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("ViewMetrics")
	mwutils.EnableCors(&w)
	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		Logger.Println("OPTIONS")
		return
	}

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	Logger.Println("aft GET check")
	rps, ok := r.URL.Query()["rp"] //rows per page
	// Query()["rp"] will return an array of items,
	// we only want the single item.
	rp, _ := strconv.Atoi(rps[0])
	if !ok || len(rps[0]) < 1 {
		Logger.Println("Url Param 'rp' is missing")
		return
	}

	pns, ok := r.URL.Query()["pn"] //page number
	pn, _ := strconv.Atoi(pns[0])
	if !ok || len(pns[0]) < 1 {
		Logger.Println("Url Param 'pn' is missing")
		return
	}

	shs, ok := r.URL.Query()["sh"] //search string
	var sh string = ""
	Logger.Println("shs")
	if len(shs) < 1 {
		Logger.Println("Url Param 'sh' is missing")
	} else {
		sh = shs[0]
	}
	sts, ok := r.URL.Query()["st"] //search string
	var st string = ""
	Logger.Println("sts")
	if len(sts) < 1 {
		Logger.Println("Url Param 'st' is missing")
	} else {
		st = sts[0]
	}
	sas, ok := r.URL.Query()["sa"] //search string
	var sa string = ""
	Logger.Println("sas")
	if len(sas) < 1 {
		Logger.Println("Url Param 'sa' is missing")
	} else {
		sa = sas[0]
	}

	var startRow, finishRow int
	startRow = (pn-1)*rp + 1
	finishRow = startRow + rp - 1

	sqlQuery := fmt.Sprintf(`SELECT m.Id, h.id, h.Hostname, Port, t.id, t.Tname, Action, m.Descr, Warning, Error,	
								Dbsid, Username, Password, Startm, Intrvlnotcrucialhrs, mname from 
							(SELECT row_number() over (order by Id), Id, Hostname, Port, Tmetric, Action, Descr, Warning, Error,	
								Dbsid, Username, Password, Startm, Intrvlnotcrucialhrs, mname from monit_sch.metrics) m
								join monit_sch.tmetrics t on t.id=m.Tmetric
								join monit_sch.hostnames h on h.id=m.Hostname
							where row_number between %d and %d`, startRow, finishRow)

	if sh != "" {
		sqlQuery = sqlQuery + " and h.hostname like '%" + sh + "%'"
	}
	if st != "" {
		sqlQuery = sqlQuery + " and t.tname like '%" + st + "%'"
	}
	if sa != "" {
		sqlQuery = sqlQuery + " and action like '%" + sa + "%'"
	}

	rows, err := ldb.Query(sqlQuery)

	if err != nil {
		Logger.Fatalf("select error: %v", err)
		return
	}
	defer rows.Close()

	Logger.Println("after select metrics")

	metrics := make([]*Metric, 0)
	for rows.Next() {

		m := new(Metric)
		err := rows.Scan(&m.Id, &m.Hid, &m.Hostname, &m.Port, &m.Tid, &m.Tmetric, &m.Action, &m.Descr, &m.Warning, &m.Error,
			&m.Dbsid, &m.Username, &m.Password, &m.Startm, &m.Intrvlnotcrucialhrs, &m.Mname)
		m.Password = crypto.Decrypt(mwutils.Key, m.Password)
		Logger.Println("m.Password decr", m.Password)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			Logger.Println("err", err)
			return
		}
		metrics = append(metrics, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		Logger.Println("err", err)
		return
	}

	metricsJson, err := json.Marshal(metrics)
	if err != nil {
		Logger.Fatal("Cannot encode to JSON ", err)
	}
	fmt.Fprintf(w, "%s", metricsJson)
}

func CreateMetric(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("CreateMetric")
	mwutils.EnableCors(&w)
	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		Logger.Println("OPTIONS")
		return
	}

	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	Logger.Println(string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)

	}

	var metric Metric
	err = json.Unmarshal(b, &metric)

	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)
		return
	}

	mHid := metric.Hid
	mPort := metric.Port
	mTid := metric.Tid
	mAction := metric.Action
	mDescr := metric.Descr
	mWarning := metric.Warning
	mError := metric.Error
	mDbsid := metric.Dbsid
	mUsername := metric.Username
	mPassword := crypto.Encrypt(mwutils.Key, metric.Password)
	mStartm := metric.Startm
	mIntrvlnotcrucialhrs := metric.Intrvlnotcrucialhrs
	mMname := metric.Mname

	txn, _ := ldb.Begin()

	var lastInsertId int16
	err = ldb.QueryRow("INSERT INTO monit_sch.Metrics (Hostname, Port, Tmetric, Action, Descr, Warning, Error,Dbsid, Username, Password, Startm, Intrvlnotcrucialhrs, Mname)"+
		" VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) returning id", mHid, mPort, mTid, mAction, mDescr, mWarning, mError, mDbsid, mUsername, mPassword, mStartm, mIntrvlnotcrucialhrs, mMname).Scan(&lastInsertId)
	Logger.Println("aft insert err", err)
	if err != nil {
		http.Error(w, http.StatusText(500)+" "+err.Error(), 500)
		txn.Rollback()
		return
	}

	err = txn.Commit()
	Logger.Println("aft insert commit err", err)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	fmt.Fprintf(w, "%d", lastInsertId)

}

func UpdateMetric(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("UpdateHostname")
	mwutils.EnableCors(&w)
	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		Logger.Println("OPTIONS")
		return
	}

	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	Logger.Println(string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)

	}

	var metric Metric
	err = json.Unmarshal(b, &metric)

	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)
		return
	}

	Logger.Println("metric.Password ", metric.Password)
	mId := metric.Id
	mHid := metric.Hid
	//mHostname := metric.Hostname
	mPort := metric.Port
	mTid := metric.Tid
	//mTmetric := metric.Tmetric
	mAction := metric.Action
	mDescr := metric.Descr
	mWarning := metric.Warning
	mError := metric.Error
	mDbsid := metric.Dbsid
	mUsername := metric.Username
	mPassword := crypto.Encrypt(mwutils.Key, metric.Password)
	mStartm := metric.Startm
	mmIntrvlnotcrucialhrs := metric.Intrvlnotcrucialhrs
	mMname := metric.Mname

	txn, _ := ldb.Begin()
	var lastUpdatedId int16
	err = ldb.QueryRow("update monit_sch.Metrics set Hostname=$1, Port=$2, Tmetric=$3, Action=$4, Descr=$5, Warning=$6, Error=$7, Dbsid=$8, Username=$9, Password=$10, Startm=$11, Intrvlnotcrucialhrs=$12, Mname=$13 where id=$14 returning id",
		mHid, mPort, mTid, mAction, mDescr, mWarning, mError, mDbsid, mUsername, mPassword, mStartm, mmIntrvlnotcrucialhrs, mMname, mId).Scan(&lastUpdatedId)
	Logger.Println("aft update err", err)
	if err != nil {
		http.Error(w, http.StatusText(500)+" "+err.Error(), 500)
		txn.Rollback()
		return
	}

	err = txn.Commit()
	Logger.Println("aft Commit err", err)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	fmt.Fprintf(w, "%d", lastUpdatedId)

}

func DeleteMetric(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("DeleteMetric")
	mwutils.EnableCors(&w)
	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		Logger.Println("OPTIONS")
		return
	}

	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	Logger.Println(string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)

	}

	var metric Metric
	err = json.Unmarshal(b, &metric)

	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)
		return
	}

	mId := metric.Id

	txn, _ := ldb.Begin()
	result, err := txn.Exec("delete from monit_sch.Metrics where id= $1", mId)
	Logger.Println("aft delete err result", err, result)
	if err != nil {
		http.Error(w, http.StatusText(500)+" "+err.Error(), 500)
		txn.Rollback()
		return
	}
	err = txn.Commit()
	Logger.Println("aft Commit err", err)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	fmt.Fprintf(w, "Metric deleted successfully (%d row affected)\n", rowsAffected)

}

func ViewMetric(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("ViewMmetric")
	//Logger.Println("req", r)

	mwutils.EnableCors(&w)
	mwutils.SetupResponse(&w, r)

	if r.Method == "OPTIONS" {
		Logger.Println("OPTIONS")
		return
	}

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	Logger.Println("aft GET check")

	ids, ok := r.URL.Query()["id"] //rows per page
	id, _ := strconv.Atoi(ids[0])
	if !ok || len(ids[0]) < 1 {
		Logger.Println("Url Param 'id' is missing")
		return
	}

	m := new(Metric)
	err := ldb.QueryRow(`SELECT m.Id, h.id, h.Hostname, Port, t.id, t.Tname, Action, m.Descr, Warning, Error,	
								Dbsid, Username, Password, Startm, Intrvlnotcrucialhrs, mname from monit_sch.metrics m
								join monit_sch.tmetrics t on t.id=m.Tmetric
								join monit_sch.hostnames h on h.id=m.Hostname
							where m.id = $1`, id).Scan(&m.Id, &m.Hid, &m.Hostname, &m.Port,
		&m.Tid, &m.Tmetric, &m.Action, &m.Descr, &m.Warning,
		&m.Error, &m.Dbsid, &m.Username, &m.Password, &m.Startm, &m.Intrvlnotcrucialhrs, &m.Mname)
	m.Password = crypto.Decrypt(mwutils.Key, m.Password)

	if err != nil {
		Logger.Fatalf("select error: %v", err)
		return
	}

	TmetricJson, err := json.Marshal(m)
	if err != nil {
		Logger.Fatal("Cannot encode to JSON ", err)
	}
	fmt.Fprintf(w, "%s", TmetricJson)
}
