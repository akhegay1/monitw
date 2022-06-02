package resrc

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

func ViewResrcs(w http.ResponseWriter, r *http.Request) {
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

	var sqlResrcWhere string
	if sh != "" {
		sqlResrcWhere = sqlResrcWhere + " where resrc like '%" + sh + "%'"
	}

	sqlQuery := fmt.Sprintf(`SELECT Id, Resrc, Descr from (SELECT row_number() over (order by Id), Id, Resrc, coalesce(Descr,'') Descr from monit_sch.Resrcs
							%s 
							) x 
						%s order by resrc`, sqlResrcWhere, PnRpWhere)
	rows, err := db.Db.Query(sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error:"+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select Resrc")

	Resrc := make([]*tbl.Resrcs, 0)
	for rows.Next() {
		m := new(tbl.Resrcs)
		err := rows.Scan(&m.Id, &m.Resrc, &m.Descr)
		//log.PrintlnWithHttp(r, "m", m, err)
		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		Resrc = append(Resrc, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	ResrcJson, err := json.Marshal(Resrc)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", ResrcJson)
}

func CreateResrc(w http.ResponseWriter, r *http.Request) {
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

	var resrc tbl.Resrcs
	err = json.Unmarshal(b, &resrc)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	log.PrintlnWithHttp(r, "crud insert", crud.CreateInsert(resrc, auth.GetUserID(ctx)))

	txn, _ := db.Db.Begin()

	var lastInsertId int64
	insq := crud.CreateInsert(resrc, auth.GetUserID(ctx))

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

func UpdateResrc(w http.ResponseWriter, r *http.Request) {
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

	var resrc tbl.Resrcs
	err = json.Unmarshal(b, &resrc)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	vId := resrc.Id
	vResrc := resrc.Resrc
	vDescr := resrc.Descr

	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("Id resrc %d vResrc %s vDescr %s", vId, vResrc, vDescr))
	if vResrc == "" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(400), 400)
		log.ErrorWithHttp(r, mwutils.FuncName(), "vResrc is empty")
		return
	}

	txn, _ := db.Db.Begin()

	var lastUpdateId int64
	updq := crud.CreateUpdate(resrc, auth.GetUserID(ctx))
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
func DeleteResrc(w http.ResponseWriter, r *http.Request) {
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

	var resrc tbl.Resrcs
	err = json.Unmarshal(b, &resrc)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	mId := strconv.FormatInt(resrc.Id, 10)

	delq := crud.CreateDelete("Resrcs", mId, auth.GetUserID(ctx))

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

	fmt.Fprintf(w, "Resrc deleted successfully (%d row affected)\n", rowsAffected)

}

func SelResrcShrt(w http.ResponseWriter, r *http.Request) {
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

	sqlQuery := "SELECT Id, Resrc from monit_sch.Resrcs"
	rows, err := db.Db.Query(sqlQuery)
	//log.PrintlnWithHttp(r, mwutils.FuncName(), sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error:"+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select Resrc")
	type sResrc struct {
		Id    int64
		Resrc string
	}

	sResrcs := make([]*sResrc, 0)
	for rows.Next() {
		m := new(sResrc)
		err := rows.Scan(&m.Id, &m.Resrc)

		if err != nil {
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			return
		}
		sResrcs = append(sResrcs, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	sResrcJson, err := json.Marshal(sResrcs)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", sResrcJson)
}

func ViewResrc(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	//log.PrintlnWithHttp(r, "req", r)
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
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("id resrc= %d", id))
	if !ok || len(ids[0]) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'id' is missing")
		return
	}

	m := new(tbl.Resrcs)
	err := db.Db.QueryRow(`SELECT Id, Resrc, COALESCE(Descr,'') Descr  
								from monit_sch.Resrcs
							where id = $1`, id).Scan(&m.Id, &m.Resrc, &m.Descr)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	ResrcJson, err := json.Marshal(m)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", ResrcJson)
}
