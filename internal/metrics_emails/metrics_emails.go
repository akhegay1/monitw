package metrics_emails

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"monitw/internal/auth"
	"monitw/internal/db"
	"monitw/pkg/mwutils"
	"net/http"
	"strconv"

	log "monitw/internal/logger"
	"monitw/pkg/crud"
	tbl "monitw/pkg/tables"
)

type MetricsEmailsSel struct {
	Id       int64
	MetricId int64
	EmailId  int64
	Fio      string
	Email    string
}

func ViewEmails4Metric(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	log.PrintlnWithHttp(r, mwutils.FuncName(), r.Method)

	mwutils.SetupResponse(&w, r)

	if r.Method == "OPTIONS" {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "OPTIONS")
		return
	}

	if r.Method != "GET" {
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+http.StatusText(405), 405)
		log.ErrorWithHttp(r, mwutils.FuncName(), http.StatusText(405))
		return
	}

	mids, ok := r.URL.Query()["mid"]
	mid, _ := strconv.Atoi(mids[0])
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("id email= %d", mid))
	if !ok || len(mids[0]) < 1 {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "Url Param 'id' is missing")
		return
	}

	sqlQuery := fmt.Sprintf(`SELECT Id, Email, Fio ||' '|| Email Fio, Descr from monit_sch.Emails 
							WHERE id not in (select emailid from monit_sch.metrics_emails where metricid=%d)
							ORDER BY Fio`, mid)
	rows, err := db.Db.Query(sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error:"+err.Error())
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select Emails")

	emails := make([]*tbl.Emails, 0)
	for rows.Next() {
		m := new(tbl.Emails)
		err := rows.Scan(&m.Id, &m.Email, &m.Fio, &m.Descr)

		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		emails = append(emails, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	EmailsJson, err := json.Marshal(emails)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", EmailsJson)
}

func ViewEmailsByMetricsId(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	mwutils.SetupResponse(&w, r)

	if r.Method == "OPTIONS" {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "OPTIONS")
		return
	}

	if r.Method != "GET" {
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+http.StatusText(405), 405)
		log.ErrorWithHttp(r, mwutils.FuncName(), http.StatusText(405))
		return
	}

	mids, ok := r.URL.Query()["mid"]
	mid, _ := strconv.Atoi(mids[0])
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("id email= %d", mid))
	if !ok || len(mids[0]) < 1 {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "Url Param 'id' is missing")
		return
	}

	sqlQuery := fmt.Sprintf(`SELECT a.Id, metricid, emailid, Fio ||' '|| Email Fio, email from monit_sch.metrics_emails a, monit_sch.emails b 
							where a.emailid=b.id and metricid=%d`, mid)
	rows, err := db.Db.Query(sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error:"+err.Error())
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select Emails")

	metrics_emails := make([]*MetricsEmailsSel, 0)
	for rows.Next() {
		m := new(MetricsEmailsSel)
		err := rows.Scan(&m.Id, &m.MetricId, &m.EmailId, &m.Fio, &m.Email)

		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		metrics_emails = append(metrics_emails, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	Metrics_emails_Json, err := json.Marshal(metrics_emails)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", Metrics_emails_Json)
}

func CreateMetricEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "OPTIONS")
		return
	}

	if r.Method != "POST" {
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+http.StatusText(405), 405)
		log.ErrorWithHttp(r, mwutils.FuncName(), http.StatusText(405))
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	log.PrintlnWithHttp(r, mwutils.FuncName(), string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())

	}

	var metrics_emails tbl.Metrics_emails
	err = json.Unmarshal(b, &metrics_emails)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	log.PrintlnWithHttp(r, "crud insert", crud.CreateInsert(metrics_emails, auth.GetUserID(ctx)))

	txn, _ := db.Db.Begin()

	var lastInsertId int64
	insq := crud.CreateInsert(metrics_emails, auth.GetUserID(ctx))

	err = db.Db.QueryRow(insq).Scan(&lastInsertId)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+http.StatusText(500)+" "+err.Error(), 500)
		txn.Rollback()
		return
	}

	err = txn.Commit()

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "%d", lastInsertId)
}

func DeleteMetricsEmails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "OPTIONS")
		return
	}

	if r.Method != "POST" {
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+http.StatusText(405), 405)
		log.ErrorWithHttp(r, mwutils.FuncName(), http.StatusText(405))
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	log.PrintlnWithHttp(r, mwutils.FuncName(), string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())

	}

	var metrics_emails tbl.Metrics_emails
	err = json.Unmarshal(b, &metrics_emails)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	mId := strconv.FormatInt(metrics_emails.Id, 10)

	delq := crud.CreateDelete("Metrics_emails", mId, auth.GetUserID(ctx))

	txn, _ := db.Db.Begin()

	result, err := txn.Exec(delq)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+http.StatusText(500)+" "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		txn.Rollback()
		return
	}
	err = txn.Commit()
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft Commit")

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "Metrics_emails deleted successfully (%d row affected)\n", rowsAffected)

}
