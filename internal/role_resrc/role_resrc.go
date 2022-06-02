package role_resrc

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

type RoleResrcSel struct {
	Id    int64
	Resrc string
}

func ViewResrcs4RoleNotGranted(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), r.Method)
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

	rids, ok := r.URL.Query()["rid"]
	rid, _ := strconv.Atoi(rids[0])
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("id role= %d", rid))
	if !ok || len(rids[0]) < 1 {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "Url Param 'rid' is missing")
		return
	}

	sqlQuery := fmt.Sprintf(`SELECT Id, Resrc, coalesce(Descr,'') descr from monit_sch.Resrcs 
							WHERE id not in (select resrcid from monit_sch.role_resrc where roleid=%d)
							ORDER BY id`, rid)
	rows, err := db.Db.Query(sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select role_resrc")

	role_resrc_not_granted := make([]*tbl.Resrcs, 0)
	for rows.Next() {
		m := new(tbl.Resrcs)
		err := rows.Scan(&m.Id, &m.Resrc, &m.Descr)
		//log.PrintlnWithHttp(r, "m", m, err)
		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		role_resrc_not_granted = append(role_resrc_not_granted, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	role_resrc_not_granted_Json, err := json.Marshal(role_resrc_not_granted)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", role_resrc_not_granted_Json)
}

func ViewResrcsByRoleId(w http.ResponseWriter, r *http.Request) {
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

	rids, ok := r.URL.Query()["rid"]
	rid, _ := strconv.Atoi(rids[0])
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("id roleid= %d", rid))

	if !ok || len(rids[0]) < 1 {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "Url Param 'rid' is missing")
		return
	}

	sqlQuery := fmt.Sprintf(`SELECT a.Id, resrc from monit_sch.role_resrc a, monit_sch.resrcs b 
								where a.resrcid=b.id and roleid=%d order by resrc`, rid)
	rows, err := db.Db.Query(sqlQuery)

	if err != nil {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select RoleResrc")
	role_resrc := make([]*RoleResrcSel, 0)
	for rows.Next() {
		m := new(RoleResrcSel)
		err := rows.Scan(&m.Id, &m.Resrc)
		//log.PrintlnWithHttp(r, "m", m, err)
		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		role_resrc = append(role_resrc, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	role_resrc_Json, err := json.Marshal(role_resrc)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", role_resrc_Json)
}

func CreateRoleResrc(w http.ResponseWriter, r *http.Request) {
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

	var role_resrc tbl.Role_resrc
	err = json.Unmarshal(b, &role_resrc)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	log.PrintlnWithHttp(r, "crud insert", crud.CreateInsert(role_resrc, auth.GetUserID(ctx)))

	txn, _ := db.Db.Begin()

	var lastInsertId int64
	insq := crud.CreateInsert(role_resrc, auth.GetUserID(ctx))

	err = db.Db.QueryRow(insq).Scan(&lastInsertId)
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft insert ")
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(500)+" "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		txn.Rollback()
		return
	}

	err = txn.Commit()
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft insert commit")

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "%d", lastInsertId)
}

func DeleteRoleResrc(w http.ResponseWriter, r *http.Request) {
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

	var role_resrc tbl.Role_resrc
	err = json.Unmarshal(b, &role_resrc)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	mId := strconv.FormatInt(role_resrc.Id, 10)

	delq := crud.CreateDelete("role_resrc", mId, auth.GetUserID(ctx))

	txn, _ := db.Db.Begin()

	result, err := txn.Exec(delq)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(500)+" "+err.Error(), 500)
		log.PrintlnWithHttp(r, mwutils.FuncName(), err.Error())
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

	fmt.Fprintf(w, "role_resrc deleted successfully (%d row affected)\n", rowsAffected)

}
