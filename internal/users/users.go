package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"monitw/internal/auth"
	"monitw/internal/crypto"
	"monitw/internal/db"
	"monitw/pkg/mutils"
	"monitw/pkg/mwutils"
	"net/http"
	"strconv"
	"strings"

	log "monitw/internal/logger"

	"monitw/pkg/crud"
	tbl "monitw/pkg/tables"
)

func ViewUsers(w http.ResponseWriter, r *http.Request) {
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
			log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'rp' is missing")
			return
		}

		pn, _ = strconv.Atoi(pns[0])
		if !ok || len(pns[0]) < 1 {
			log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'pn' is missing")
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
	log.PrintlnWithHttp(r, mwutils.FuncName(), "shs")
	log.PrintlnWithHttp(r, mwutils.FuncName(), strings.Join(shs, " "))
	if len(shs) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'sh' is missing")
	} else {
		sh = shs[0]
	}

	var sqlUsersWhere string
	if sh != "" {
		sqlUsersWhere = sqlUsersWhere + " where username like '%" + sh + "%'"
	}

	sqlQuery := fmt.Sprintf(`SELECT Id, Fio, Username, Passw, DtStart  from (SELECT row_number() over (order by Id), Id, coalesce(Fio,'') Fio, 
								Username, Passw, coalesce(TO_CHAR(dtstart , 'YYYY-MM-DD hh24:mi:ss'),'') DtStart from monit_sch.Users
							%s 
							) x 
						%s order by Username`, sqlUsersWhere, PnRpWhere)
	rows, err := db.Db.Query(sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error 1 : "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	log.PrintlnWithHttp(r, mwutils.FuncName(), "after select Users")

	users := make([]*tbl.Users, 0)
	for rows.Next() {
		m := new(tbl.Users)
		err := rows.Scan(&m.Id, &m.Fio, &m.Username, &m.Passw, &m.DtStart)

		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), "m="+m.Username+" err "+err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		users = append(users, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	UsersJson, err := json.Marshal(users)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", UsersJson)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
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

	var user tbl.Users
	err = json.Unmarshal(b, &user)
	log.PrintlnWithHttp(r, "user.Passw", user.Passw)
	user.Passw = crypto.Encrypt(mutils.Key, user.Passw)
	log.PrintlnWithHttp(r, "user.Passw", user.Passw)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	log.PrintlnWithHttp(r, "crud insert", crud.CreateInsert(user, auth.GetUserID(ctx)))

	txn, _ := db.Db.Begin()

	var lastInsertId int64
	insq := crud.CreateInsert(user, auth.GetUserID(ctx))

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
		log.PrintlnWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "%d", lastInsertId)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
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

	var user tbl.Users
	err = json.Unmarshal(b, &user)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	vId := user.Id
	vFio := user.Fio
	vUsername := user.Username
	vPassw := user.Passw

	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("Id user  %d vFio %s vUsername %s", vId, vFio, vUsername))

	if vUsername == "" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(400), 400)
		log.ErrorWithHttp(r, mwutils.FuncName(), "vUsername is empty")
		return
	}
	if vPassw == "" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(400), 400)
		log.ErrorWithHttp(r, mwutils.FuncName(), "vPassw is empty")
		return
	}

	//log.PrintlnWithHttp(r, "user.Passw", user.Passw)
	user.Passw = crypto.Encrypt(mutils.Key, user.Passw)
	//log.PrintlnWithHttp(r, "user.Passw", user.Passw)

	txn, _ := db.Db.Begin()

	var lastUpdateId int64
	updq := crud.CreateUpdate(user, auth.GetUserID(ctx))
	log.PrintlnWithHttp(r, "updq", updq)

	err = db.Db.QueryRow(updq).Scan(&lastUpdateId)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(500)+" "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		txn.Rollback()
		return
	}

	err = txn.Commit()

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.PrintlnWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "%d", lastUpdateId)

}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
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

	var user tbl.Users
	err = json.Unmarshal(b, &user)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	mId := strconv.FormatInt(user.Id, 10)

	delq := crud.CreateDelete("Users", mId, auth.GetUserID(ctx))

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
		return
	}

	fmt.Fprintf(w, "User deleted successfully (%d row affected)\n", rowsAffected)

}

func SelUsersShrt(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "OPTIONS")
		return
	}

	if r.Method != "GET" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(405), 405)
		return
	}
	//log.PrintlnWithHttp(r, "aft GET check")

	sqlQuery := "SELECT Id, Username from monit_sch.Users"
	rows, err := db.Db.Query(sqlQuery)
	//log.PrintlnWithHttp(r, "sqlQuery", sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select Users")
	type sUser struct {
		Id       int64
		Username string
	}

	sUsers := make([]*sUser, 0)
	for rows.Next() {
		m := new(sUser)
		err := rows.Scan(&m.Id, &m.Username)

		if err != nil {
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			return
		}
		sUsers = append(sUsers, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	sUsersJson, err := json.Marshal(sUsers)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", sUsersJson)
}

func ViewUser(w http.ResponseWriter, r *http.Request) {
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
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("id user= %d", id))
	if !ok || len(ids[0]) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'id' is missing")
		return
	}

	m := new(tbl.Users)
	err := db.Db.QueryRow(`SELECT Id, coalesce(Fio,'') Fio, Username, Passw, coalesce(TO_CHAR(dtstart , 'YYYY-MM-DD"T"HH24:MI:SS'),'') DtStart  
								from monit_sch.Users
							where id = $1`, id).Scan(&m.Id, &m.Fio, &m.Username, &m.Passw, &m.DtStart)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	UserJson, err := json.Marshal(m)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", UserJson)
}
