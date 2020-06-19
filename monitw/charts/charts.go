package charts

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"monitw/mwutils"
	"net/http"
	"time"
)

type VmParams struct {
	HostnameId int16
	StartTime  string
	FinishTime string
}

type ChVmetric struct {
	Hostmname string
	Vtime     string
	Value     string
}

var Logger *log.Logger

func GetVmetricsBydt(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("GetVmetricsBydt")
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

	var vmparams VmParams
	err = json.Unmarshal(b, &vmparams)

	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)
		return
	}
	//layout := "2006-01-02T15:04:05"
	mHostnameId := vmparams.HostnameId
	Logger.Println("vmparams.StartTime ", vmparams.StartTime)
	mStartTime, err := time.Parse(time.RFC3339, vmparams.StartTime)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	mFinishTime, err := time.Parse(time.RFC3339, vmparams.FinishTime)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	Logger.Println("mHostnameId", mHostnameId)
	Logger.Println("mStartTime", mStartTime)
	Logger.Println("mFinishTime", mFinishTime)

	Logger.Println("bef sel")
	sql := "select h.hostname || '/' || m.mname hostmname,to_char(vtime, 'hh24:mi') vtime,value from monit_sch.vmetrics v " +
		"join monit_sch.metrics m on m.id=v.metric " +
		"join monit_sch.hostnames h on m.hostname=h.id " +
		"join monit_sch.tmetrics t on m.tmetric=t.id " +
		"where m.intrvlnotcrucialhrs=0 and vtime > $1::timestamptz and vtime < $2::timestamptz and h.id=$3 " +
		"order by h.hostname,t.tname"
	Logger.Println(sql)

	rows, err := ldb.Query(sql, mStartTime, mFinishTime, mHostnameId)
	Logger.Println("aft query", rows)

	vmetrics := make([]*ChVmetric, 0)
	for rows.Next() {
		m := new(ChVmetric)
		err := rows.Scan(&m.Hostmname, &m.Vtime, &m.Value)

		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			Logger.Println("err", err)
			return
		}
		vmetrics = append(vmetrics, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		Logger.Println("err", err)
		return
	}

	VmetricsJson, err := json.Marshal(vmetrics)
	if err != nil {
		Logger.Fatal("Cannot encode to JSON ", err)
	}
	fmt.Fprintf(w, "%s", VmetricsJson)
}
