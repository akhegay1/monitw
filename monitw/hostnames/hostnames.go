package hostnames

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

type Hostname struct {
	Id       int16
	Hostname string
	Descr    string
}

var Logger *log.Logger

func ViewHostnames(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("ViewHostnames")
	Logger.Println("req", r)

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

	shs, ok := r.URL.Query()["sh"] //search string
	var sh string = ""
	Logger.Println("shs")
	Logger.Println(shs)
	if len(shs) < 1 {
		Logger.Println("Url Param 'sh' is missing")
	} else {
		sh = shs[0]
	}

	var startRow, finishRow int
	startRow = (pn-1)*rp + 1
	finishRow = startRow + rp - 1

	sqlQuery := fmt.Sprintf(`SELECT Id, Hostname, Descr from (SELECT row_number() over (order by Id), Id, Hostname, COALESCE(Descr,'') Descr from monit_sch.Hostnames) x
						where row_number between %d and %d`, startRow, finishRow)
	Logger.Println("sh sql" + sh)
	if sh != "" {
		sqlQuery = sqlQuery + " and hostname like '%" + sh + "%'"
	}
	rows, err := ldb.Query(sqlQuery)
	//Logger.Println("sqlQuery", sqlQuery)

	if err != nil {
		Logger.Fatalf("select error: %v", err)
		return
	}
	defer rows.Close()

	//Logger.Println("after select Hosts")

	Hosts := make([]*Hostname, 0)
	for rows.Next() {
		m := new(Hostname)
		err := rows.Scan(&m.Id, &m.Hostname, &m.Descr)

		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			Logger.Println("err", err)
			return
		}
		Hosts = append(Hosts, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		Logger.Println("err", err)
		return
	}

	HostsJson, err := json.Marshal(Hosts)
	if err != nil {
		Logger.Fatal("Cannot encode to JSON ", err)
	}
	fmt.Fprintf(w, "%s", HostsJson)
}

func CreateHostname(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("CreateHostname")
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
	//Logger.Println(string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)

	}

	var hostname Hostname
	err = json.Unmarshal(b, &hostname)

	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)
		return
	}

	hHostname := hostname.Hostname
	hDescr := hostname.Descr
	if hHostname == "" {
		http.Error(w, http.StatusText(400), 400)
		Logger.Println("Hostname is empty")
		return
	}

	txn, _ := ldb.Begin()

	var lastInsertId int16
	err = ldb.QueryRow("INSERT INTO monit_sch.Hostnames (hostname, descr)"+
		" VALUES($1, $2) returning id", hHostname, hDescr).Scan(&lastInsertId)
	Logger.Println("aft insert err", err)
	if err != nil {
		http.Error(w, http.StatusText(500)+" "+err.Error(), 500)
		txn.Rollback()
		return
	}
	err = txn.Commit()
	Logger.Println("aft insert commit err", err)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	Logger.Println("lastInsertId", lastInsertId)
	fmt.Fprintf(w, "%d", lastInsertId)

}

func UpdateHostname(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("UpdateHostname")
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

	var hostname Hostname
	err = json.Unmarshal(b, &hostname)

	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)
		return
	}

	hId := hostname.Id
	hHostname := hostname.Hostname
	hDescr := hostname.Descr

	txn, _ := ldb.Begin()
	var lastUpdatedId int16
	err = ldb.QueryRow("update monit_sch.Hostnames set hostname=$1, Descr=$2 where id=$3 returning id", hHostname, hDescr, hId).Scan(&lastUpdatedId)
	Logger.Println("aft update err", err)
	if err != nil {
		http.Error(w, http.StatusText(500)+" "+err.Error(), 500)
		txn.Rollback()
		return
	}

	err = txn.Commit()
	Logger.Println("aft update commit err", err)

	if err != nil {
		http.Error(w, http.StatusText(500)+err.Error(), 500)
		return
	}

	fmt.Fprintf(w, "%d", lastUpdatedId)

}

func DeleteHostname(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("DeleteHostname")
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

	var Hostname Hostname
	err = json.Unmarshal(b, &Hostname)

	if err != nil {
		http.Error(w, err.Error(), 500)
		Logger.Println("err", err)
		return
	}

	hId := Hostname.Id
	Logger.Println("Id", hId)

	txn, _ := ldb.Begin()
	result, err := txn.Exec("delete from monit_sch.Hostnames where id= $1", hId)
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

	fmt.Fprintf(w, "Hostname deleted successfully (%d row affected)\n", rowsAffected)

}

func SelHostnamesShrt(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
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

	sqlQuery := "SELECT Id, Hostname from monit_sch.Hostnames"
	rows, err := ldb.Query(sqlQuery)
	//Logger.Println("sqlQuery", sqlQuery)

	if err != nil {
		Logger.Fatalf("select error: %v", err)
		return
	}
	defer rows.Close()

	//Logger.Println("after select Hosts")
	type sHostname struct {
		Id       int16
		Hostname string
	}

	sHosts := make([]*sHostname, 0)
	for rows.Next() {
		m := new(sHostname)
		err := rows.Scan(&m.Id, &m.Hostname)

		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			Logger.Println("err", err)
			return
		}
		sHosts = append(sHosts, m)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		Logger.Println("err", err)
		return
	}

	sHostsJson, err := json.Marshal(sHosts)
	if err != nil {
		Logger.Fatal("Cannot encode to JSON ", err)
	}
	fmt.Fprintf(w, "%s", sHostsJson)
}

func ViewHostname(w http.ResponseWriter, r *http.Request, ldb *sql.DB) {
	Logger.Println("ViewHostname")
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

	m := new(Hostname)
	err := ldb.QueryRow(`SELECT Id, Hostname, Descr 
								from monit_sch.Hostnames
							where id = $1`, id).Scan(&m.Id, &m.Hostname, &m.Descr)

	if err != nil {
		Logger.Fatalf("select error: %v", err)
		return
	}

	HostJson, err := json.Marshal(m)
	if err != nil {
		Logger.Fatal("Cannot encode to JSON ", err)
	}
	fmt.Fprintf(w, "%s", HostJson)
}
