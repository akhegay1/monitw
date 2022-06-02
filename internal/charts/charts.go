package charts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"monitw/internal/db"
	log "monitw/internal/logger"
	"monitw/pkg/mwutils"
	"net/http"
	"time"
)

type VmParams struct {
	HostnameId int64
	StartTime  string
	FinishTime string
}

type ChVmetric struct {
	Hostmname string
	Vtime     string
	Value     string
}

func GetVmetricsBydt(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "method OPTIONS")
		return
	}

	if r.Method != "POST" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(405), 405)
		log.ErrorWithHttp(r, mwutils.FuncName(), http.StatusText(405))
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)

	defer r.Body.Close()
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
	}

	var vmparams VmParams
	err = json.Unmarshal(b, &vmparams)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}
	//layout := "2006-01-02T15:04:05"
	mHostnameId := vmparams.HostnameId
	log.PrintlnWithHttp(r, mwutils.FuncName(), "vmparams.StartTime "+vmparams.StartTime)

	mStartTime, err := time.Parse(time.RFC3339, vmparams.StartTime)
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 400)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}
	mFinishTime, err := time.Parse(time.RFC3339, vmparams.FinishTime)
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 400)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("mHostnameId %d", mHostnameId))
	log.PrintlnWithHttp(r, mwutils.FuncName(), "mStartTime "+mStartTime.String())
	log.PrintlnWithHttp(r, mwutils.FuncName(), "mFinishTime "+mFinishTime.String())

	sql := "select h.hostname || '/' || m.mname hostmname,to_char(vtime, 'hh24:mi') vtime,value from monit_sch.vmetrics v " +
		"join monit_sch.metrics m on m.id=v.metric " +
		"join monit_sch.vhostnames h on m.hostname=h.id " +
		"join monit_sch.tmetrics t on m.tmetric=t.id " +
		"where m.intrvlnotcrucialhrs=0 and vtime > $1::timestamptz and vtime < $2::timestamptz and h.id=$3 " +
		"order by h.hostname,t.tname"
	log.PrintlnWithHttp(r, mwutils.FuncName(), sql)

	rows, err := db.Db.Query(sql, mStartTime, mFinishTime, mHostnameId)
	log.PrintlnWithHttp(r, mwutils.FuncName(), sql)

	vmetrics := make([]*ChVmetric, 0)
	for rows.Next() {
		m := new(ChVmetric)
		err := rows.Scan(&m.Hostmname, &m.Vtime, &m.Value)

		if err != nil {
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			return
		}
		vmetrics = append(vmetrics, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	VmetricsJson, err := json.Marshal(vmetrics)
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
	}
	fmt.Fprintf(w, "%s", VmetricsJson)
}
