package user_role

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

type UserRoleSel struct {
	Id   int64
	Role string
}

type UsernameRoleSel struct {
	Role string
}

func ViewRoles4UserNotGranted(w http.ResponseWriter, r *http.Request) {
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

	sqlQuery := fmt.Sprintf(`SELECT Id, Role, coalesce(Descr,'') descr from monit_sch.Roles 
							WHERE id not in (select roleid from monit_sch.user_role where userid=%d)
							ORDER BY Role`, uid)
	rows, err := db.Db.Query(sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	user_role_not_granted := make([]*tbl.Roles, 0)
	for rows.Next() {
		m := new(tbl.Roles)
		err := rows.Scan(&m.Id, &m.Role, &m.Descr)
		log.PrintlnWithHttp(r, mwutils.FuncName(), "m="+m.Role)
		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		user_role_not_granted = append(user_role_not_granted, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	user_role_not_granted_Json, err := json.Marshal(user_role_not_granted)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", user_role_not_granted_Json)
}

func ViewRolesByUserId(w http.ResponseWriter, r *http.Request) {
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
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("id roleid= %d", uid))

	if !ok || len(uids[0]) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'rid' is missing")
		return
	}

	sqlQuery := fmt.Sprintf(`SELECT a.Id, role from monit_sch.user_role a, monit_sch.roles b 
								where a.roleid=b.id and userid=%d order by role`, uid)
	rows, err := db.Db.Query(sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select UserRole"
	user_role := make([]*UserRoleSel, 0)
	for rows.Next() {
		m := new(UserRoleSel)
		err := rows.Scan(&m.Id, &m.Role)
		log.PrintlnWithHttp(r, mwutils.FuncName(), "m="+m.Role)
		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		user_role = append(user_role, m)
	}

	if err = rows.Err(); err != nil {
		log.PrintlnWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	user_role_Json, err := json.Marshal(user_role)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", user_role_Json)
}

func ViewRolesByUserid(w http.ResponseWriter, r *http.Request) {
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

	uns, ok := r.URL.Query()["un"]
	un := uns[0]
	log.PrintlnWithHttp(r, "username=", un)
	if !ok || len(uns[0]) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'un' is missing")
		return
	}

	sqlQuery := fmt.Sprintf(`SELECT b.role from monit_sch.user_role a, monit_sch.roles b, monit_sch.users c 
								where a.roleid=b.id and c.id = a.userid and c.id ='%s' order by role`, un)
	log.PrintlnWithHttp(r, mwutils.FuncName(), "sqlQuery "+sqlQuery)
	rows, err := db.Db.Query(sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select UserRole"
	username_role := make([]*UsernameRoleSel, 0)
	for rows.Next() {
		m := new(UsernameRoleSel)
		err := rows.Scan(&m.Role)
		log.ErrorWithHttp(r, mwutils.FuncName(), "m "+m.Role)
		if err != nil {
			log.PrintlnWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		username_role = append(username_role, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	user_role_Json, err := json.Marshal(username_role)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", user_role_Json)
}

func CreateUserRole(w http.ResponseWriter, r *http.Request) {
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

	var user_role tbl.User_role
	err = json.Unmarshal(b, &user_role)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	log.PrintlnWithHttp(r, "crud insert", crud.CreateInsert(user_role, auth.GetUserID(ctx)))

	txn, _ := db.Db.Begin()

	var lastInsertId int64
	insq := crud.CreateInsert(user_role, auth.GetUserID(ctx))

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

func DeleteUserRole(w http.ResponseWriter, r *http.Request) {
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

	var user_role tbl.User_role
	err = json.Unmarshal(b, &user_role)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	mId := strconv.FormatInt(user_role.Id, 10)

	delq := crud.CreateDelete("user_role", mId, auth.GetUserID(ctx))

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

	fmt.Fprintf(w, "user_role deleted successfully (%d row affected)\n", rowsAffected)

}
