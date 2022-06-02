package tmetrics

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

type TmetricsSel struct {
	Id    int64
	Tname string
}

func ViewTmetrics(w http.ResponseWriter, r *http.Request) {
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

	sqlQuery := fmt.Sprintf(`SELECT Id, Tname from (SELECT row_number() over (order by Id), Id, Tname from monit_sch.Tmetrics) x
						where row_number between %d and %d`, startRow, finishRow)
	rows, err := db.Db.Query(sqlQuery)
	log.PrintlnWithHttp(r, mwutils.FuncName(), "sqlQuery"+sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	log.PrintlnWithHttp(r, mwutils.FuncName(), "after select Tmetrics")

	Tmetrics := make([]*TmetricsSel, 0)

	for rows.Next() {
		m := new(TmetricsSel)
		err := rows.Scan(&m.Id, &m.Tname)
		//log.PrintlnWithHttp(r, mwutils.FuncName(), "m "+m.Tname+err.Error())
		if err != nil {
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			return
		}
		Tmetrics = append(Tmetrics, m)
	}

	if err = rows.Err(); err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	TmetricsJson, err := json.Marshal(Tmetrics)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", TmetricsJson)
}

func CreateTmetric(w http.ResponseWriter, r *http.Request) {
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

	var tmetrics tbl.Tmetrics
	err = json.Unmarshal(b, &tmetrics)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	txn, _ := db.Db.Begin()

	var lastInsertId int64
	insq := crud.CreateInsert(tmetrics, auth.GetUserID(ctx))

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

func UpdateTmetric(w http.ResponseWriter, r *http.Request) {
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

	var tmetrics tbl.Tmetrics
	err = json.Unmarshal(b, &tmetrics)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	vId := tmetrics.Id
	vTname := tmetrics.Tname

	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("Id tmetrics %d %s", vId, vTname))
	if vTname == "" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(400), 400)
		log.ErrorWithHttp(r, mwutils.FuncName(), "vTname is empty")
		return
	}

	txn, _ := db.Db.Begin()

	var lastUpdateId int64
	updq := crud.CreateUpdate(tmetrics, auth.GetUserID(ctx))
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
func DeleteTmetric(w http.ResponseWriter, r *http.Request) {
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

	var tmetrics tbl.Tmetrics
	err = json.Unmarshal(b, &tmetrics)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	mId := strconv.FormatInt(tmetrics.Id, 10)

	delq := crud.CreateDelete("Tmetrics", mId, auth.GetUserID(ctx))

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

	fmt.Fprintf(w, "Tmetric deleted successfully (%d row affected)\n", rowsAffected)

}

func SelTmetricsShrt(w http.ResponseWriter, r *http.Request) {
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

	sqlQuery := "SELECT Id, Tname from monit_sch.Tmetrics"
	rows, err := db.Db.Query(sqlQuery)
	//log.PrintlnWithHttp(r, "sqlQuery", sqlQuery)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}
	defer rows.Close()

	//log.PrintlnWithHttp(r, "after select Tmetrics")
	type sTmetric struct {
		Id    int64
		Tname string
	}

	sTmetrics := make([]*sTmetric, 0)
	for rows.Next() {
		m := new(sTmetric)
		err := rows.Scan(&m.Id, &m.Tname)

		if err != nil {
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			return
		}
		sTmetrics = append(sTmetrics, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.PrintlnWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	sTmetricsJson, err := json.Marshal(sTmetrics)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", sTmetricsJson)
}

func ViewTmetric(w http.ResponseWriter, r *http.Request) {
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
	log.PrintlnWithHttp(r, mwutils.FuncName(), fmt.Sprintf("id tmetrics=%d", id))
	if !ok || len(ids[0]) < 1 {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Url Param 'id' is missing")
		return
	}

	m := new(tbl.Tmetrics)
	err := db.Db.QueryRow(`SELECT Id, coalesce(Tname,'') Tname  
								from monit_sch.Tmetrics
							where id = $1`, id).Scan(&m.Id, &m.Tname)

	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "select error: "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		return
	}

	TmetricJson, err := json.Marshal(m)
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), "Cannot encode to JSON "+err.Error())
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
	}
	fmt.Fprintf(w, "%s", TmetricJson)
}
