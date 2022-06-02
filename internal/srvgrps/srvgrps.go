package srvgrps

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

type SrvgrpsSel struct {
	Id      int64
	Grpname string
}

func ViewSrvgrps(w http.ResponseWriter, r *http.Request) {
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
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'pn' is missing")
		return
	}
	//for pagination
	var startRow, finishRow int
	startRow = (pn-1)*rp + 1
	finishRow = startRow + rp - 1

	sqlQuery := fmt.Sprintf(`SELECT Id, Grpname from (SELECT row_number() over (order by Id), Id, Grpname from monit_sch.Srvgrps) x
						where row_number between %d and %d`, startRow, finishRow)
	rows, err := db.Db.Query(sqlQuery)
	log.PrintlnWithHttp(r, mwutils.FuncName(), sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	log.PrintlnWithHttp(r, mwutils.FuncName(), "after select Srvgrps")

	Srvgrps := make([]*SrvgrpsSel, 0)

	for rows.Next() {
		m := new(SrvgrpsSel)
		err := rows.Scan(&m.Id, &m.Grpname)
		log.PrintlnWithHttp(r, mwutils.FuncName(), "m "+m.Grpname)
		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		Srvgrps = append(Srvgrps, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	SrvgrpsJson, err := json.Marshal(Srvgrps)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", SrvgrpsJson)
}

func CreateSrvgrp(w http.ResponseWriter, r *http.Request) {
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

	var srvgrps tbl.Srvgrps
	err = json.Unmarshal(b, &srvgrps)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	log.PrintlnWithHttp(r, "CreateSrvgrp", crud.CreateInsert(srvgrps, auth.GetUserID(ctx)))

	txn, _ := db.Db.Begin()

	var lastInsertId int64
	insq := crud.CreateInsert(srvgrps, auth.GetUserID(ctx))

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

func UpdateSrvgrp(w http.ResponseWriter, r *http.Request) {
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

	var srvgrps tbl.Srvgrps
	err = json.Unmarshal(b, &srvgrps)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	vId := srvgrps.Id
	vGrpname := srvgrps.Grpname

	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("Id srvgrps %d vGrpname %s", vId, vGrpname))
	if vGrpname == "" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(400), 400)
		log.ErrorWithHttp(r, mwutils.FuncName(), "vGrpname is empty")
		return
	}

	txn, _ := db.Db.Begin()

	var lastUpdateId int64
	updq := crud.CreateUpdate(srvgrps, auth.GetUserID(ctx))
	log.PrintlnWithHttp(r, mwutils.FuncName(), "updq "+updq)

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
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "%d", lastUpdateId)

}
func DeleteSrvgrp(w http.ResponseWriter, r *http.Request) {
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

	var srvgrps tbl.Srvgrps
	err = json.Unmarshal(b, &srvgrps)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	mId := strconv.FormatInt(srvgrps.Id, 10)

	delq := crud.CreateDelete("Srvgrps", mId, auth.GetUserID(ctx))

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

	fmt.Fprintf(w, "Srvgrp deleted successfully (%d row affected)\n", rowsAffected)

}

func SelSrvgrpsShrt(w http.ResponseWriter, r *http.Request) {
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

	sqlQuery := "SELECT Id, Grpname from monit_sch.Srvgrps"
	rows, err := db.Db.Query(sqlQuery)
	//log.PrintlnWithHttp(r, mwutils.FuncName(), sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error:"+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select Srvgrps")
	type sSrvgrp struct {
		Id      int64
		Grpname string
	}

	sSrvgrps := make([]*sSrvgrp, 0)
	for rows.Next() {
		m := new(sSrvgrp)
		err := rows.Scan(&m.Id, &m.Grpname)

		if err != nil {
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			return
		}
		sSrvgrps = append(sSrvgrps, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	sSrvgrpsJson, err := json.Marshal(sSrvgrps)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", sSrvgrpsJson)
}

func ViewSrvgrp(w http.ResponseWriter, r *http.Request) {
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
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft GET check")

	ids, ok := r.URL.Query()["id"]
	id, _ := strconv.Atoi(ids[0])
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("id srvgrps= %d", id))
	if !ok || len(ids[0]) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'id' is missing")
		return
	}

	m := new(tbl.Srvgrps)
	err := db.Db.QueryRow(`SELECT Id, coalesce(Grpname,'') Grpname  
								from monit_sch.Srvgrps
							where id = $1`, id).Scan(&m.Id, &m.Grpname)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	SrvgrpJson, err := json.Marshal(m)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", SrvgrpJson)
}
