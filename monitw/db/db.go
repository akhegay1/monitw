package db

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"monitw/charts"
	"monitw/hostnames"
	"monitw/metrics"
	"monitw/mwutils"
	"monitw/tmetrics"
	"monitw/vmetrics"
	"net/http"
	"os"
	"strings"
)

var Db *sql.DB

var (
	host     = ""
	port     = ""
	user     = ""
	password = ""
	dbname   = ""
)
var words []string
var Logger *log.Logger

func Connect() string {
	Logger.Println("db connect")
	key := mwutils.Key
	///////////!!!!!!!ENCRYPT PASSW!!!!!!!/////
	/*
		passw := "postgres"
		foo := Encrypt(key, passw)
		fmt.Println("foo", foo)
	*/

	////////////////FILE/////
	conf, err := os.Open("conf")
	if err != nil {
		Logger.Fatalf("failed opening file conf: %s", err)
		return fmt.Sprintf("failed opening file conf: %s", err)
	}
	defer conf.Close()

	sc := bufio.NewScanner(conf)

	for sc.Scan() {
		str := sc.Text() // GET the line string
		words = strings.Fields(str)
	}

	if err := sc.Err(); err != nil {
		Logger.Fatalf("scan file error: %v", err)
		return fmt.Sprintf("scan file error: %v", err)
	}

	host = words[0]
	port = words[1]
	user = words[2]
	password = words[3]
	dbname = words[4]

	user = Decrypt(key, user)
	password = Decrypt(key, password)

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	Db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	//defer Db.Close()

	err = Db.Ping()
	if err != nil {
		panic(err)
	}

	return "Successfully connected!"
}

///////////////////////ENCRYPT DECRYPT////////////////////////////////////////
var iv = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

func encodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func decodeBase64(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func Encrypt(key, text string) string {

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}
	plaintext := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)
	return encodeBase64(ciphertext)
}

func Decrypt(key, text string) string {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}
	ciphertext := decodeBase64(text)
	cfb := cipher.NewCFBEncrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	cfb.XORKeyStream(plaintext, ciphertext)
	return string(plaintext)
}

///////////////////////   API   ////////////////////
///////////////////TMetrics/////////////////////////
func ViewTmetrics(w http.ResponseWriter, r *http.Request) {
	tmetrics.ViewTmetrics(w, r, Db)
}
func ViewTmetric(w http.ResponseWriter, r *http.Request) {
	tmetrics.ViewTmetric(w, r, Db)
}
func CreateTmetric(w http.ResponseWriter, r *http.Request) {
	tmetrics.CreateTmetric(w, r, Db)
}
func UpdateTmetric(w http.ResponseWriter, r *http.Request) {
	tmetrics.UpdateTmetric(w, r, Db)
}
func DeleteTmetric(w http.ResponseWriter, r *http.Request) {
	tmetrics.DeleteTmetric(w, r, Db)
}
func SelTmetricsShrt(w http.ResponseWriter, r *http.Request) {
	tmetrics.SelTmetricsShrt(w, r, Db)
}

///////////////////Hostnames/////////////////////////

func ViewHostnames(w http.ResponseWriter, r *http.Request) {
	Logger.Println("db ViewHostnames")
	hostnames.ViewHostnames(w, r, Db)
}
func ViewHostname(w http.ResponseWriter, r *http.Request) {
	Logger.Println("db ViewHostname")
	hostnames.ViewHostname(w, r, Db)
}
func CreateHostname(w http.ResponseWriter, r *http.Request) {
	hostnames.CreateHostname(w, r, Db)
}
func UpdateHostname(w http.ResponseWriter, r *http.Request) {
	hostnames.UpdateHostname(w, r, Db)
}
func DeleteHostname(w http.ResponseWriter, r *http.Request) {
	hostnames.DeleteHostname(w, r, Db)
}
func SelHostnamesShrt(w http.ResponseWriter, r *http.Request) {
	hostnames.SelHostnamesShrt(w, r, Db)
}

///////////////////Metrics/////////////////////////
func ViewMetrics(w http.ResponseWriter, r *http.Request) {
	metrics.ViewMetrics(w, r, Db)
}
func ViewMetric(w http.ResponseWriter, r *http.Request) {
	metrics.ViewMetric(w, r, Db)
}
func CreateMetric(w http.ResponseWriter, r *http.Request) {
	metrics.CreateMetric(w, r, Db)
}
func UpdateMetric(w http.ResponseWriter, r *http.Request) {
	metrics.UpdateMetric(w, r, Db)
}
func DeleteMetric(w http.ResponseWriter, r *http.Request) {
	metrics.DeleteMetric(w, r, Db)
}

///////////////////Vmetrics/////////////////////////
func ViewVmetrics(w http.ResponseWriter, r *http.Request) {
	vmetrics.ViewVmetrics(w, r, Db)
}

///////////////////Charts/////////////////////////
func GetVmetricsBydt(w http.ResponseWriter, r *http.Request) {
	Logger.Println("AAA")
	charts.GetVmetricsBydt(w, r, Db)
}
