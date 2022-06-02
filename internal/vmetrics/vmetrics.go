package vmetrics

import (
	"encoding/json"
	"fmt"
	"monitw/internal/db"
	"monitw/pkg/mwutils"
	"net/http"
	"strconv"
	"time"

	log "monitw/internal/logger"

	"github.com/gorilla/websocket"
)

type Vmetric struct {
	Id            int64
	Hid           int64
	Hostname      string
	Tname         string
	Mname         string
	Vtime         time.Time
	Metric        int64
	Value         float64
	Lastm         int64
	Warning       float64
	Error         float64
	Execerr       string
	Tresholdismin bool
}

type VmetricAll struct {
	Id            int64
	Hid           int64
	Hostname      string
	Srvgrpid      int64
	Mname         string
	Vtime         time.Time
	Metric        int64
	Value         float64
	Lastm         int64
	Warning       float64
	Error         float64
	Execerr       string
	Tresholdismin bool
}

func ViewVmetrics(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

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
	//log.PrintlnWithHttp(r, "aft GET check")

	tms, ok := r.URL.Query()["tm"] //rows per page
	tm, _ := strconv.Atoi(tms[0])
	if !ok || len(tms[0]) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'tm' is missing")
		return
	}

	//sg - servergroups
	sgs, ok := r.URL.Query()["sg"] //rows per page
	sg, _ := strconv.Atoi(sgs[0])
	if !ok || len(sgs[0]) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'sg' is missing")
		return
	}

	//log.PrintlnWithHttp(r, "tm", tm)
	sqlQuery := fmt.Sprintf(`SELECT v.id,h.id hid, h.hostname,t.tname,m.mname,vtime,metric,value,lastm, warning, error, execerr,m.Tresholdismin from monit_sch.vmetrics v
								join monit_sch.metrics m on m.id=v.metric
								join monit_sch.vhostnames h on m.hostname=h.id
								join monit_sch.tmetrics t on m.tmetric=t.id
								join monit_sch.srvgrps s on h.srvgrpid=s.id
								where m.intrvlnotcrucialhrs=0
								and lastm=(SELECT max(lastm) FROM monit_sch.vmetrics)
								and vtime > clock_timestamp()- INTERVAL '5 minutes'
								and t.id=%d
								and s.id=%d
								order by m.mname, h.hostname`, tm, sg)
	rows, err := db.Db.Query(sqlQuery)
	//log.PrintlnWithHttp(r, "sqlQuery", sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select Vmetrics")

	Vmetrics := make([]*Vmetric, 0)
	for rows.Next() {
		m := new(Vmetric)
		err := rows.Scan(&m.Id, &m.Hid, &m.Hostname, &m.Tname, &m.Mname, &m.Vtime, &m.Metric, &m.Value, &m.Lastm, &m.Warning, &m.Error, &m.Execerr, &m.Tresholdismin)
		//log.PrintlnWithHttp(r, "m", m, err)
		if err != nil {
			log.PrintlnWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		Vmetrics = append(Vmetrics, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	VmetricsJson, err := json.Marshal(Vmetrics)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", VmetricsJson)
}

var upgrader = websocket.Upgrader{} // use default options

func ViewVmetricsAllWS(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	mwutils.SetupResponse(&w, r)
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	log.PrintlnWithHttp(r, mwutils.FuncName(), "after upgrader.CheckOrigin")

	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Error during connection upgradation:"+err.Error())
		return
	}
	defer conn.Close()

	// The event loop

	isOpen := true
	for isOpen {

		err = conn.WriteMessage(websocket.TextMessage, []byte(ViewVmetricsAll(w, r)))
		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), "Error during message writing:"+err.Error())
			isOpen = false
			break
		}
		time.Sleep(time.Second * 10)

	}

}

func ViewVmetricsAll(w http.ResponseWriter, r *http.Request) []byte {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "OPTIONS")
		return []byte("err")
	}

	if r.Method != "GET" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(405), 405)
		log.ErrorWithHttp(r, mwutils.FuncName(), http.StatusText(405))
		return []byte("err")
	}
	//log.PrintlnWithHttp(r, "aft GET check")

	sqlQuery := `SELECT v.id, h.id hid,h.hostname,h.srvgrpid, m.mname,vtime,metric,value,lastm, warning, error, execerr,m.Tresholdismin from monit_sch.vmetrics v
					join monit_sch.metrics m on m.id=v.metric
					join monit_sch.vhostnames h on m.hostname=h.id
					where m.intrvlnotcrucialhrs=0
					and lastm=(SELECT max(lastm) FROM monit_sch.vmetrics)
					and vtime > clock_timestamp()- INTERVAL '5 minutes'
					order by h.srvgrpid `
	rows, err := db.Db.Query(sqlQuery)
	//log.PrintlnWithHttp(r, "sqlQuery", sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return []byte("err")
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select Vmetrics")

	VmetricsAll := make([]*VmetricAll, 0)
	for rows.Next() {
		m := new(VmetricAll)
		err := rows.Scan(&m.Id, &m.Hid, &m.Hostname, &m.Srvgrpid, &m.Mname, &m.Vtime, &m.Metric, &m.Value, &m.Lastm, &m.Warning, &m.Error, &m.Execerr, &m.Tresholdismin)
		//log.PrintlnWithHttp(r, "m", m, err)
		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return []byte("err")
		}
		VmetricsAll = append(VmetricsAll, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return []byte("err")
	}

	VmetricsAllJson, err := json.Marshal(VmetricsAll)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	//fmt.Fprintf(w, "%s", VmetricsAllJson)
	return VmetricsAllJson
}
