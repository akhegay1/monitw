package emails

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"monitw/internal/auth"
	"monitw/internal/db"
	"monitw/pkg/mwutils"
	"net/http"
	"strconv"
	"strings"

	log "monitw/internal/logger"

	"monitw/pkg/crud"
	tbl "monitw/pkg/tables"
)

func ViewEmails(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	log.PrintlnWithHttp(r, mwutils.FuncName(), r.Method)
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
	rps, ok := r.URL.Query()["rp"] //rows per page
	pns, ok := r.URL.Query()["pn"] //page number
	var rp int
	var pn int
	var PnRpWhere string

	if (rps != nil) && (pns != nil) {
		// Query()["rp"] will return an array of items,
		// we only want the single item.
		rp, _ = strconv.Atoi(rps[0])
		if !ok || len(rps[0]) < 1 {
			log.PrintlnWithHttp(r, mwutils.FuncName(), "Url Param 'rp' is missing")
			return
		}

		pn, _ = strconv.Atoi(pns[0])
		if !ok || len(pns[0]) < 1 {
			log.PrintlnWithHttp(r, mwutils.FuncName(), "Url Param 'pn' is missing")
			return
		}
		//for pagination
		var startRow, finishRow int
		startRow = (pn-1)*rp + 1
		finishRow = startRow + rp - 1

		PnRpWhere = fmt.Sprintf(`where row_number between %d and %d`, startRow, finishRow)
	}

	shs, ok := r.URL.Query()["sh"] //search string
	var sh string = ""
	log.PrintlnWithHttp(r, mwutils.FuncName(), "shs "+strings.Join(shs, ""))
	if len(shs) < 1 {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "Url Param 'sh' is missing")
	} else {
		sh = shs[0]
	}

	var sqlHostWhere string
	if sh != "" {
		sqlHostWhere = sqlHostWhere + " where email like '%" + sh + "%'"
	}

	sqlQuery := fmt.Sprintf(`SELECT Id, Email, Fio, Descr from (SELECT row_number() over (order by Id), Id, Email, Fio, Descr from monit_sch.Emails
							%s 
							) x 
						%s order by Fio`, sqlHostWhere, PnRpWhere)
	rows, err := db.Db.Query(sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	emails := make([]*tbl.Emails, 0)
	for rows.Next() {
		m := new(tbl.Emails)
		err := rows.Scan(&m.Id, &m.Email, &m.Fio, &m.Descr)
		log.PrintlnWithHttp(r, mwutils.FuncName(), m.Fio)
		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		emails = append(emails, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	EmailsJson, err := json.Marshal(emails)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON  "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", EmailsJson)
}

func CreateEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

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
	//log.PrintlnWithHttp(r, mwutils.FuncName(), string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
	}

	var email tbl.Emails
	err = json.Unmarshal(b, &email)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	log.PrintlnWithHttp(r, mwutils.FuncName(), crud.CreateInsert(email, auth.GetUserID(ctx)))

	txn, _ := db.Db.Begin()

	var lastInsertId int64
	insq := crud.CreateInsert(email, auth.GetUserID(ctx))

	err = db.Db.QueryRow(insq).Scan(&lastInsertId)

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

	fmt.Fprintf(w, "%d", lastInsertId)
}

func UpdateEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "fifnished")

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

	var email tbl.Emails
	err = json.Unmarshal(b, &email)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	vId := email.Id
	vEmail := email.Email
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("Id Email %d %s", vId, vEmail))
	if vEmail == "" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(400), 400)
		log.ErrorWithHttp(r, mwutils.FuncName(), "email is empty")
		return
	}

	txn, _ := db.Db.Begin()

	var lastUpdateId int64
	updq := crud.CreateUpdate(email, auth.GetUserID(ctx))
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
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft Commit")

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "%d", lastUpdateId)

}
func DeleteEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

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

	var email tbl.Emails
	err = json.Unmarshal(b, &email)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	mId := strconv.FormatInt(email.Id, 10)

	delq := crud.CreateDelete("Emails", mId, auth.GetUserID(ctx))

	txn, _ := db.Db.Begin()

	result, err := txn.Exec(delq)
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft delete err result ")
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

	fmt.Fprintf(w, "Email deleted successfully (%d row affected)\n", rowsAffected)

}

func SelEmailsShrt(w http.ResponseWriter, r *http.Request) {
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

	sqlQuery := "SELECT Id, Email from monit_sch.Emails"
	rows, err := db.Db.Query(sqlQuery)
	//log.PrintlnWithHttp(r, "sqlQuery", sqlQuery)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select Emails")
	type sEmail struct {
		Id    int64
		Email string
	}

	sEmails := make([]*sEmail, 0)
	for rows.Next() {
		m := new(sEmail)
		err := rows.Scan(&m.Id, &m.Email)

		if err != nil {
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			return
		}
		sEmails = append(sEmails, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	sEmailsJson, err := json.Marshal(sEmails)
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
	}
	fmt.Fprintf(w, "%s", sEmailsJson)
}

func ViewEmail(w http.ResponseWriter, r *http.Request) {
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
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft GET check")

	ids, ok := r.URL.Query()["id"]
	id, _ := strconv.Atoi(ids[0])
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("id email=%d", id))
	if !ok || len(ids[0]) < 1 {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "Url Param 'id' is missing")
		return
	}

	m := new(tbl.Emails)
	err := db.Db.QueryRow(`SELECT Id, Email, Fio,Descr  
								from monit_sch.Emails
							where id = $1`, id).Scan(&m.Id, &m.Email, &m.Fio, &m.Descr)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	EmailJson, err := json.Marshal(m)
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		return
	}
	fmt.Fprintf(w, "%s", EmailJson)
}
