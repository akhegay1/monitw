package hostnames

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"monitw/internal/db"
	"monitw/pkg/mwutils"
	"net/http"
	"strconv"
	"strings"

	auth "monitw/internal/auth"
	log "monitw/internal/logger"

	"monitw/pkg/crud"
	tbl "monitw/pkg/tables"
)

type HostnamesSel struct {
	Id       int64
	Hostname string
	Descr    string
	Grpname  string
	Srvgrpid int64
}

func ViewHostnames(w http.ResponseWriter, r *http.Request) {
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
	rps, ok := r.URL.Query()["rp"] //rows per page
	// Query()["rp"] will return an array of items,
	// we only want the single item.
	rp, _ := strconv.Atoi(rps[0])
	if !ok || len(rps[0]) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'rp' is missing")
		return
	}

	pns, ok := r.URL.Query()["pn"] //page number
	pn, _ := strconv.Atoi(pns[0])
	if !ok || len(pns[0]) < 1 {
		http.Error(w, mwutils.FuncName()+": "+"Url Param 'pn' is missing", 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'pn' is missing")
		return
	}

	shs, ok := r.URL.Query()["sh"] //search string
	var sh string = ""
	log.PrintlnWithHttp(r, mwutils.FuncName(), "shs")
	log.PrintlnWithHttp(r, mwutils.FuncName(), strings.Join(shs, " "))
	if len(shs) < 1 {
		//http.Error(w, mwutils.FuncName() +  ": " +  "Url Param 'sh' is missing", 500)
		log.PrintlnWithHttp(r, mwutils.FuncName(), "Url Param 'sh' is missing")
	} else {
		sh = shs[0]
	}

	var startRow, finishRow int
	startRow = (pn-1)*rp + 1
	finishRow = startRow + rp - 1
	var sqlHostWhere string
	if sh != "" {
		sqlHostWhere = sqlHostWhere + " where hostname like '%" + sh + "%'"
	}

	sqlQuery := fmt.Sprintf(`SELECT x.Id, Hostname, Descr, s.grpname, Srvgrpid from monit_sch.Srvgrps s,
									(SELECT row_number() over (order by Id), Id, Hostname, COALESCE(Descr,'') Descr, Srvgrpid from monit_sch.Hostnames
									%s
									) x
							where s.id=x.Srvgrpid and row_number between %d and %d`, sqlHostWhere, startRow, finishRow)
	log.PrintlnWithHttp(r, mwutils.FuncName(), "sh sql"+sh)

	rows, err := db.Query(sqlQuery)
	//Logger.Println("sqlQuery", sqlQuery)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		return
	}
	defer rows.Close()

	//Logger.Println("after select Hosts")

	Hosts := make([]*HostnamesSel, 0)
	for rows.Next() {
		m := new(HostnamesSel)
		err := rows.Scan(&m.Id, &m.Hostname, &m.Descr, &m.Grpname, &m.Srvgrpid)

		if err != nil {
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			return
		}
		Hosts = append(Hosts, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	HostsJson, err := json.Marshal(Hosts)
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
	}
	fmt.Fprintf(w, "%s", HostsJson)
}

func CreateHostname(w http.ResponseWriter, r *http.Request) {
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

	var hostnames tbl.Hostnames
	err = json.Unmarshal(b, &hostnames)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	log.PrintlnWithHttp(r, "crud insert", crud.CreateInsert(hostnames, auth.GetUserID(ctx)))

	txn, _ := db.Db.Begin()

	var lastInsertId int64

	insq := crud.CreateInsert(hostnames, auth.GetUserID(ctx))

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

func UpdateHostname(w http.ResponseWriter, r *http.Request) {
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

	var hostnames tbl.Hostnames
	err = json.Unmarshal(b, &hostnames)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	vId := hostnames.Id
	vHostname := hostnames.Hostname
	vSrvgrpid := hostnames.Srvgrpid

	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("Id hostnames  %d vHostname %s vSrvgrpid %d", vId, vHostname, vSrvgrpid))
	if vHostname == "" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(400), 400)
		log.ErrorWithHttp(r, mwutils.FuncName(), "vHostname is empty")
		return
	}

	txn, _ := db.Db.Begin()

	var lastUpdateId int64
	updq := crud.CreateUpdate(hostnames, auth.GetUserID(ctx))
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
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft Commit ")

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "%d", lastUpdateId)

}
func DeleteHostname(w http.ResponseWriter, r *http.Request) {
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

	var hostnames tbl.Hostnames
	err = json.Unmarshal(b, &hostnames)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	mId := strconv.FormatInt(hostnames.Id, 10)

	delq := crud.CreateDelete("Hostnames", mId, auth.GetUserID(ctx))
	log.PrintlnWithHttp(r, "delq=", delq)

	txn, _ := db.Db.Begin()
	log.PrintlnWithHttp(r, mwutils.FuncName(), "bef exec")
	result, err := txn.Exec(delq)
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(500)+" "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		txn.Rollback()
		return
	}

	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(500)+" "+" no records deleted ", 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("delete rowsAffected= %d", rowsAffected))
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	err = txn.Commit()
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft Commit ")

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "Hostname deleted successfully (%d row affected)\n", rowsAffected)

}

func SelHostnamesShrt(w http.ResponseWriter, r *http.Request) {
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

	sqlQuery := "SELECT Id, Hostname from monit_sch.Hostnames"
	rows, err := db.Db.Query(sqlQuery)
	//log.PrintlnWithHttp(r, "sqlQuery", sqlQuery)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select Hostnames")
	type sHostname struct {
		Id       int64
		Hostname string
	}

	sHostnames := make([]*sHostname, 0)
	for rows.Next() {
		m := new(sHostname)
		err := rows.Scan(&m.Id, &m.Hostname)

		if err != nil {
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			return
		}
		sHostnames = append(sHostnames, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	sHostnamesJson, err := json.Marshal(sHostnames)
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
	}
	fmt.Fprintf(w, "%s", sHostnamesJson)
}

func ViewHostname(w http.ResponseWriter, r *http.Request) {
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
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("id hostnames= %d", id))
	if !ok || len(ids[0]) < 1 {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "Url Param 'id' is missing")
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'id' is missin")
		return
	}

	m := new(tbl.Hostnames)
	err := db.Db.QueryRow(`SELECT Id, coalesce(Hostname,'') Hostname, coalesce(Descr,'') Descr, Srvgrpid  
								from monit_sch.Hostnames
							where id = $1`, id).Scan(&m.Id, &m.Hostname, &m.Descr, &m.Srvgrpid)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		return
	}

	Hostnamejson, err := json.Marshal(m)
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
	}
	fmt.Fprintf(w, "%s", Hostnamejson)
}
