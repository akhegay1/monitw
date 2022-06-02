package user_resrc

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

type UserResrcSel struct {
	Id    int64
	Resrc string
}

func ViewResrcs4UserNotGranted(w http.ResponseWriter, r *http.Request) {
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

	uids, ok := r.URL.Query()["uid"]
	uid, _ := strconv.Atoi(uids[0])
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("id user= %d", uid))

	if !ok || len(uids[0]) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'uid' is missing")
		return
	}

	sqlQuery := fmt.Sprintf(`SELECT Id, Resrc, coalesce(Descr,'') descr from monit_sch.Resrcs 
							WHERE id not in (select resrcid from monit_sch.user_resrc where usrid=%d)
							ORDER BY Resrc`, uid)
	rows, err := db.Db.Query(sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select user_resrc")

	user_resrc_not_granted := make([]*tbl.Resrcs, 0)
	for rows.Next() {
		m := new(tbl.Resrcs)
		err := rows.Scan(&m.Id, &m.Resrc, &m.Descr)
		log.PrintlnWithHttp(r, mwutils.FuncName(), "m="+m.Resrc)
		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		user_resrc_not_granted = append(user_resrc_not_granted, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	user_resrc_not_granted_Json, err := json.Marshal(user_resrc_not_granted)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", user_resrc_not_granted_Json)
}

func ViewResrcsByUserId(w http.ResponseWriter, r *http.Request) {
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

	uids, ok := r.URL.Query()["uid"]
	uid, _ := strconv.Atoi(uids[0])
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("id userid= %d", uid))

	if !ok || len(uids[0]) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'uid' is missing")
		return
	}

	sqlQuery := fmt.Sprintf(`SELECT a.Id, resrc from monit_sch.user_resrc a, monit_sch.resrcs b 
								where a.resrcid=b.id and usrid=%d order by resrc`, uid)
	rows, err := db.Db.Query(sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error:"+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select UserResrc"
	user_resrc := make([]*UserResrcSel, 0)
	for rows.Next() {
		m := new(UserResrcSel)
		err := rows.Scan(&m.Id, &m.Resrc)
		log.PrintlnWithHttp(r, mwutils.FuncName(), "m="+m.Resrc)
		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		user_resrc = append(user_resrc, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	user_resrc_Json, err := json.Marshal(user_resrc)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", user_resrc_Json)
}

func CreateUserResrc(w http.ResponseWriter, r *http.Request) {
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

	var user_resrc tbl.User_resrc
	err = json.Unmarshal(b, &user_resrc)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	log.PrintlnWithHttp(r, "crud insert", crud.CreateInsert(user_resrc, auth.GetUserID(ctx)))

	txn, _ := db.Db.Begin()

	var lastInsertId int64
	insq := crud.CreateInsert(user_resrc, auth.GetUserID(ctx))

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

func DeleteUserResrc(w http.ResponseWriter, r *http.Request) {
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
		log.PrintlnWithHttp(r, mwutils.FuncName(), http.StatusText(405))
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

	var user_resrc tbl.User_resrc
	err = json.Unmarshal(b, &user_resrc)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	mId := strconv.FormatInt(user_resrc.Id, 10)

	delq := crud.CreateDelete("user_resrc", mId, auth.GetUserID(ctx))

	txn, _ := db.Db.Begin()

	result, err := txn.Exec(delq)
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

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "user_resrc deleted successfully (%d row affected)\n", rowsAffected)

}
