package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"monitw/charts"
	"monitw/db"
	"monitw/hostnames"
	"monitw/metrics"
	"monitw/mwutils"
	"monitw/tmetrics"
	"monitw/vmetrics"
	"net/http"
	"os"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
)

var errorlog *os.File
var logger *log.Logger
var words []string

func init() {
	errorlog, err := os.OpenFile("monitw.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Printf("error opening file: %v", err)
		os.Exit(1)
	}
	//defer errorlog.Close()
	logger = log.New(errorlog, "applog: ", log.Lshortfile|log.LstdFlags)
	logger.Println("main")

	db.Logger = logger
	metrics.Logger = logger
	hostnames.Logger = logger
	tmetrics.Logger = logger
	vmetrics.Logger = logger
	charts.Logger = logger

	c := db.Connect()
	logger.Println("connect ", c)
}

const (
	APP_KEY = "golangcode.com"
)

func main() {
	logger.Println("Start ")
	http.HandleFunc("/metrics", db.ViewMetrics)
	http.HandleFunc("/metric", db.ViewMetric)
	http.HandleFunc("/metrics/create", db.CreateMetric)
	http.HandleFunc("/metrics/update", db.UpdateMetric)
	http.HandleFunc("/metrics/delete", db.DeleteMetric)
	http.HandleFunc("/hostnames", db.ViewHostnames)
	http.HandleFunc("/hostname", db.ViewHostname)
	http.HandleFunc("/hostnames/create", db.CreateHostname)
	http.HandleFunc("/hostnames/update", db.UpdateHostname)
	http.HandleFunc("/hostnames/delete", db.DeleteHostname)
	http.HandleFunc("/shostnames", db.SelHostnamesShrt)
	http.HandleFunc("/tmetrics", db.ViewTmetrics)
	http.HandleFunc("/tmetric", db.ViewTmetric)
	http.HandleFunc("/tmetrics/create", db.CreateTmetric)
	http.HandleFunc("/tmetrics/update", db.UpdateTmetric)
	http.HandleFunc("/tmetrics/delete", db.DeleteTmetric)
	http.HandleFunc("/stmetrics", db.SelTmetricsShrt)
	http.HandleFunc("/vmetrics", db.ViewVmetrics)
	http.HandleFunc("/vmetrics/getbydt", db.GetVmetricsBydt)
	/*
		!!!Надо добавить методы для single row !!!!
		http.HandleFunc("/auth", TokenHandler)
		http.Handle("/metrics", AuthMiddleware(http.HandlerFunc(db.ViewMetrics)))
		http.Handle("/metrics/create", AuthMiddleware(http.HandlerFunc(db.CreateMetric)))
		http.Handle("/metrics/update", AuthMiddleware(http.HandlerFunc(db.UpdateMetric)))
		http.Handle("/metrics/delete", AuthMiddleware(http.HandlerFunc(db.DeleteMetric)))
		http.Handle("/hostnames", AuthMiddleware(http.HandlerFunc(db.ViewHostnames)))
		http.Handle("/hostnames/create", AuthMiddleware(http.HandlerFunc(db.CreateHostname)))
		http.Handle("/hostnames/update", AuthMiddleware(http.HandlerFunc(db.UpdateHostname)))
		http.Handle("/hostnames/delete", AuthMiddleware(http.HandlerFunc(db.DeleteHostname)))
		http.Handle("/shostnames", AuthMiddleware(http.HandlerFunc(db.SelHostnamesShrt)))
		http.Handle("/tmetrics", AuthMiddleware(http.HandlerFunc(db.ViewTmetrics)))
		http.Handle("/tmetrics/create", AuthMiddleware(http.HandlerFunc(db.CreateTmetric)))
		http.Handle("/tmetrics/update", AuthMiddleware(http.HandlerFunc(db.UpdateTmetric)))
		http.Handle("/tmetrics/delete", AuthMiddleware(http.HandlerFunc(db.DeleteTmetric)))
		http.Handle("/stmetrics", AuthMiddleware(http.HandlerFunc(db.SelTmetricsShrt)))
		http.Handle("/vmetrics", AuthMiddleware(http.HandlerFunc(db.ViewVmetrics)))
		http.Handle("/vmetrics/getbydt", AuthMiddleware(http.HandlerFunc(db.GetVmetricsBydt)))*/
	http.ListenAndServe(":3010", nil)
	logger.Println("Aft Listen ")
}

// TokenHandler is our handler to take a username and password and,
// if it's valid, return a token used for future requests.
func TokenHandler(w http.ResponseWriter, r *http.Request) {

	mwutils.EnableCors(&w)
	mwutils.SetupResponse(&w, r)
	w.Header().Add("Content-Type", "application/json")

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	//Logger.Println(string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		logger.Println("err", err)

	}
	type Login struct {
		Username string
		Password string
	}
	var login Login
	err = json.Unmarshal(b, &login)

	if err != nil {
		http.Error(w, err.Error(), 500)
		logger.Println("err", err)
		return
	}

	username := login.Username
	password := login.Password

	///////////////Get from DB////////
	var userdb string
	var passwdb string
	err = db.Db.QueryRow("SELECT username,passw FROM monit_sch.users WHERE username=$1", username).Scan(&userdb, &passwdb)
	logger.Println("userdb", len(userdb))
	//logger.Println("passwdb", passwdb)
	if err != nil || len(userdb) == 0 {
		logger.Println("err", err)
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":"invalid_credentials"}`)
		return
	}

	if username != userdb || password != passwdb {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":"invalid_credentials"}`)
		return
	}

	// We are happy with the credentials, so build a token. We've given it
	// an expiry of 1 hour.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": username,
		"exp":  time.Now().Add(time.Hour * time.Duration(4320)).Unix(),
		"iat":  time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(APP_KEY))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"error":"token_generation_failed"}`)
		return
	}
	io.WriteString(w, `{"token":"`+tokenString+`"}`)
	return
}

// AuthMiddleware is our middleware to check our token is valid. Returning
// a 401 status to the client if it is not valid.
func AuthMiddleware(next http.Handler) http.Handler {
	logger.Println("AuthMiddleware")
	if len(APP_KEY) == 0 {
		log.Fatal("HTTP server unable to start, expected an APP_KEY for JWT auth")
	}
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(APP_KEY), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
	logger.Println("AuthMiddleware bef next", next)
	return jwtMiddleware.Handler(next)
}
