package vmetrics

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"monitw/mwutils"
	"net/http"
	"strconv"
	"time"
)

type Vmetric struct {
	Id       int16
	Hid      int16
	Hostname string
	Tname    string
	Mname    string
	Vtime    time.Time
	Metric   int16
	Value    float64
	Lastm    int16
	Warning  float64
	Error    float64
	Execerr  string
}

var Logger *log.Logger

func ViewVmetrics(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println(r.Method)
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
	//Logger.Println("aft GET check")

	//Logger.Println("aft GET check")
	tms, ok := r.URL.Query()["tm"] //rows per page
	tm, _ := strconv.Atoi(tms[0])
	if !ok || len(tms[0]) < 1 {
		Logger.Println("Url Param 'tm' is missing")
		return
	}
	//Logger.Println("tm", tm)
	sqlQuery := fmt.Sprintf(`SELECT v.id,h.id hid, h.hostname,t.tname,m.mname,vtime,metric,value,lastm, warning, error, execerr from monit_sch.vmetrics v
								join monit_sch.metrics m on m.id=v.metric
								join monit_sch.hostnames h on m.hostname=h.id
								join monit_sch.tmetrics t on m.tmetric=t.id
								where m.intrvlnotcrucialhrs=0
								and lastm=(SELECT max(lastm) FROM monit_sch.vmetrics)
								and vtime > clock_timestamp()- INTERVAL '5 minutes'
								and t.id=%d
								order by m.mname, h.hostname`, tm)
	rows, err := ldb.Query(sqlQuery)
	//Logger.Println("sqlQuery", sqlQuery)

	if err != nil {
		Logger.Fatalf("select error: %v", err)
		return
	}
	defer rows.Close()

	//Logger.Println("after select Vmetrics")

	Vmetrics := make([]*Vmetric, 0)
	for rows.Next() {
		m := new(Vmetric)
		err := rows.Scan(&m.Id, &m.Hid, &m.Hostname, &m.Tname, &m.Mname, &m.Vtime, &m.Metric, &m.Value, &m.Lastm, &m.Warning, &m.Error, &m.Execerr)
		Logger.Println("m", m, err)
		if err != nil {
			Logger.Println("err", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		Vmetrics = append(Vmetrics, m)
	}

	if err = rows.Err(); err != nil {
		Logger.Println("err", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	VmetricsJson, err := json.Marshal(Vmetrics)
	if err != nil {
		Logger.Fatal("Cannot encode to JSON ", err)
	}
	fmt.Fprintf(w, "%s", VmetricsJson)
}
