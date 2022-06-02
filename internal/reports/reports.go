package reports

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"monitw/internal/db"
	log "monitw/internal/logger"
	"monitw/pkg/mwutils"
	"net/http"
	"os"
	"strings"
	"time"

	wkhtml "github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

type MaxCpuValue struct {
	Hostname    string
	MaxCpuValue float32
}

type MaxCpuValuesS struct {
	MaxCpuValuesAr []MaxCpuValue
}

var templpathparams []string

func F_rep_001(w http.ResponseWriter, r *http.Request) {

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

	dtStart := r.URL.Query().Get("dtStart")
	layout := "2006-01-02T15:04"
	t, err := time.Parse(layout, dtStart)
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}
	log.PrintlnWithHttp(r, "dtStart", t.String())

	dtFinish := r.URL.Query().Get("dtFinish")
	t, err = time.Parse(layout, dtFinish)
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}
	log.PrintlnWithHttp(r, "dtFinish", t.String())

	sql := "select h.hostname, max(value) MaxCpuValue from monit_sch.vmetrics v " +
		"join monit_sch.metrics m on m.id=v.metric " +
		"join monit_sch.hostnames h on m.hostname = h.id " +
		"where m.mname ='cpu' " +
		"and vtime >  $1::timestamptz and vtime < $2::timestamptz " +
		"group by m.mname,h.hostname " +
		"order by h.hostname "

	log.PrintlnWithHttp(r, mwutils.FuncName(), sql)
	dts := []interface{}{dtStart, dtFinish}
	rows, err := db.Db.Query(sql, dts...)
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}
	log.PrintlnWithHttp(r, mwutils.FuncName(), sql)

	MaxCpuValues := make([]MaxCpuValue, 0)
	log.PrintlnWithHttp(r, mwutils.FuncName(), "after make([]*MaxCpuValue")
	for rows.Next() {

		m := new(MaxCpuValue)
		err := rows.Scan(&m.Hostname, &m.MaxCpuValue)
		log.PrintlnWithHttp(r, mwutils.FuncName(), m.Hostname+"|"+fmt.Sprintf("%f", m.MaxCpuValue))

		if err != nil {
			http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
			log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
			return
		}
		MaxCpuValues = append(MaxCpuValues, *m)
	}

	//var MaxCpuValuesS1 MaxCpuValuesS

	if err = rows.Err(); err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	paramsfile, err := os.Open("params")
	if err != nil {
		log.Println(mwutils.FuncName(), "failed opening file params: "+err.Error())
	}
	defer paramsfile.Close()

	sc := bufio.NewScanner(paramsfile)

	for sc.Scan() {
		str := sc.Text() // GET the line string
		if len(str) == 0 {
			continue
		} else if str[0:1] == "#" {
			continue
		}
		parnam := str[0:strings.Index(str, "=")]
		if parnam == "templpath" {
			val := str[strings.Index(str, "=")+1:]
			templpathparams = append(templpathparams, val)
			break
		}
	}

	if err := sc.Err(); err != nil {
		log.Println(mwutils.FuncName(), "scan file error: "+err.Error())
	}

	//log.Println(mwutils.FuncName(), templpathparams[0])

	MaxCpuValuesVar := MaxCpuValuesS{
		MaxCpuValuesAr: MaxCpuValues,
	}

	tmpl := template.Must(template.ParseFiles(templpathparams[0] + "rep_001.html"))

	var b bytes.Buffer
	tmpl.Execute(&b, MaxCpuValuesVar)

	pdfg, err := wkhtml.NewPDFGenerator()
	if err != nil {
		fmt.Println(err)
		return
	}

	pdfg.AddPage(wkhtml.NewPageReader(strings.NewReader(b.String())))
	pdfg.Orientation.Set(wkhtml.OrientationLandscape)

	// Create PDF document in internal buffer
	err = pdfg.Create()
	if err != nil {
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
	}

	//Your Pdf Name
	/*err = pdfg.WriteFile("./Your_pdfname.pdf")
	if err != nil {
		log.Fatal(err)
	}*/

	w.Header().Set("Content-Disposition", "attachment; filename=kittens.pdf")
	w.Header().Set("Content-Type", "application/pdf")
	w.WriteHeader(http.StatusOK)
	w.Write(pdfg.Bytes())

	//fmt.Println("Done")

	//http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	//http.ListenAndServe(":3000", nil)
}
