package tmetrics

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"monitw/mwutils"
	"net/http"
	"strconv"
)

type Tmetric struct {
	Id    int16
	Tname string
}

var Logger *log.Logger

func ViewTmetrics(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println(r.Method)
	mwutils.EnableCors(&w)
	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		Logger.Println("OPTIONS")
		return
	}

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	Logger.Println("aft GET check")
	rps, ok := r.URL.Query()["rp"] //rows per page
	// Query()["rp"] will return an array of items,
	// we only want the single item.
	rp, _ := strconv.Atoi(rps[0])
	if !ok || len(rps[0]) < 1 {
		Logger.Println("Url Param 'rp' is missing")
		return
	}

	pns, ok := r.URL.Query()["pn"] //page number
	pn, _ := strconv.Atoi(pns[0])
	if !ok || len(pns[0]) < 1 {
		Logger.Println("Url Param 'pn' is missing")
		return
	}

	var startRow, finishRow int
	startRow = (pn-1)*rp + 1
	finishRow = startRow + rp - 1

	sqlQuery := fmt.Sprintf(`SELECT Id, Tname from (SELECT row_number() over (order by Id), Id, Tname from monit_sch.Tmetrics) x
						where row_number between %d and %d`, startRow, finishRow)
	rows, err := ldb.Query(sqlQuery)
	Logger.Println("sqlQuery", sqlQuery)

	if err != nil {
		Logger.Fatalf("select error: %v", err)
		return
	}
	defer rows.Close()

	Logger.Println("after select Hosts")

	Hosts := make([]*Tmetric, 0)
	for rows.Next() {
		m := new(Tmetric)
		err := rows.Scan(&m.Id, &m.Tname)
		Logger.Println("m", m, err)
		if err != nil {
			Logger.Println("err", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		Hosts = append(Hosts, m)
	}

	if err = rows.Err(); err != nil {
		Logger.Println("err", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	HostsJson, err := json.Marshal(Hosts)
	if err != nil {
		Logger.Fatal("Cannot encode to JSON ", err)
	}
	fmt.Fprintf(w, "%s", HostsJson)
}

func CreateTmetric(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("CreateTmetric")
	mwutils.EnableCors(&w)
	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		Logger.Println("OPTIONS")
		return
	}

	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	Logger.Println(string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)

	}

	var tmetric Tmetric
	err = json.Unmarshal(b, &tmetric)

	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)
		return
	}

	tname := tmetric.Tname
	if tname == "" {
		http.Error(w, http.StatusText(400), 400)
		Logger.Println("tmetric is empty")
		return
	}

	txn, _ := ldb.Begin()

	var lastInsertId int16
	err = ldb.QueryRow("INSERT INTO monit_sch.tmetrics (tname) VALUES($1) returning id", tname).Scan(&lastInsertId)
	Logger.Println("aft insert err", err)
	if err != nil {
		http.Error(w, http.StatusText(500)+" "+err.Error(), 500)
		txn.Rollback()
		return
	}

	err = txn.Commit()
	Logger.Println("aft insert Commit")

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	Logger.Println("lastInsertId", lastInsertId)
	fmt.Fprintf(w, "%d", lastInsertId)
}

func UpdateTmetric(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("UpdateTmetric")
	mwutils.EnableCors(&w)
	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		Logger.Println("OPTIONS")
		return
	}

	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	Logger.Println(string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)

	}

	var tmetric Tmetric
	err = json.Unmarshal(b, &tmetric)

	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)
		return
	}

	tId := tmetric.Id
	tname := tmetric.Tname
	Logger.Println("Id Tname", tId, tname)
	if tname == "" {
		http.Error(w, http.StatusText(400), 400)
		Logger.Println("tmetric is empty")
		return
	}

	txn, _ := ldb.Begin()
	var lastUpdatedId int16
	err = ldb.QueryRow("update monit_sch.tmetrics set tname=$1 where id=$2 returning id", tname, tId).Scan(&lastUpdatedId)
	Logger.Println("aft update err", err)
	if err != nil {
		http.Error(w, http.StatusText(500)+" "+err.Error(), 500)
		txn.Rollback()
		return
	}

	if err != nil {
		http.Error(w, http.StatusText(500)+" "+err.Error(), 500)
		return
	}
	err = txn.Commit()
	Logger.Println("aft Commit err", err)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	fmt.Fprintf(w, "%d", lastUpdatedId)

}

func DeleteTmetric(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("DeleteTmetric")
	mwutils.EnableCors(&w)
	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		Logger.Println("OPTIONS")
		return
	}

	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	Logger.Println(string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)

	}

	var tmetric Tmetric
	err = json.Unmarshal(b, &tmetric)

	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)
		return
	}

	tId := tmetric.Id
	Logger.Println("Id", tId)

	txn, _ := ldb.Begin()
	result, err := txn.Exec("delete from monit_sch.tmetrics where id= $1", tId)
	Logger.Println("aft delete err result", err, result)
	if err != nil {
		http.Error(w, http.StatusText(500)+" "+err.Error(), 500)
		txn.Rollback()
		return
	}
	err = txn.Commit()
	Logger.Println("aft Commit err", err)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	fmt.Fprintf(w, "Tname deleted successfully (%d row affected)\n", rowsAffected)

}

func SelTmetricsShrt(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	//Logger.Println("ViewHostnames")
	mwutils.EnableCors(&w)
	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		Logger.Println("OPTIONS")
		return
	}

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	//Logger.Println("aft GET check")

	sqlQuery := "SELECT Id, Tname from monit_sch.Tmetrics"
	rows, err := ldb.Query(sqlQuery)
	//Logger.Println("sqlQuery", sqlQuery)

	if err != nil {
		Logger.Fatalf("select error: %v", err)
		return
	}
	defer rows.Close()

	//Logger.Println("after select Hosts")
	type sTmetric struct {
		Id    int16
		Tname string
	}

	sTmetrics := make([]*sTmetric, 0)
	for rows.Next() {
		m := new(sTmetric)
		err := rows.Scan(&m.Id, &m.Tname)

		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			Logger.Println("err", err)
			return
		}
		sTmetrics = append(sTmetrics, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		Logger.Println("err", err)
		return
	}

	sTmetricsJson, err := json.Marshal(sTmetrics)
	if err != nil {
		Logger.Fatal("Cannot encode to JSON ", err)
	}
	fmt.Fprintf(w, "%s", sTmetricsJson)
}

func ViewTmetric(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("ViewTmetric")
	//Logger.Println("req", r)

	mwutils.EnableCors(&w)
	mwutils.SetupResponse(&w, r)

	if r.Method == "OPTIONS" {
		Logger.Println("OPTIONS")
		return
	}

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	Logger.Println("aft GET check")

	ids, ok := r.URL.Query()["id"] //rows per page
	id, _ := strconv.Atoi(ids[0])
	if !ok || len(ids[0]) < 1 {
		Logger.Println("Url Param 'id' is missing")
		return
	}

	m := new(Tmetric)
	err := ldb.QueryRow(`SELECT Id, Tname  
								from monit_sch.Tmetrics
							where id = $1`, id).Scan(&m.Id, &m.Tname)

	if err != nil {
		Logger.Fatalf("select error: %v", err)
		return
	}

	TmetricJson, err := json.Marshal(m)
	if err != nil {
		Logger.Fatal("Cannot encode to JSON ", err)
	}
	fmt.Fprintf(w, "%s", TmetricJson)
}
