package logger

import (
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

var errorlog *os.File
var words []string
var app string = "MONITW"

func init() {

	// open a file

	f, err := os.OpenFile("monitw.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	//fmt.Println("after open logfile f=" + f.Name())
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}

	//Logger = log.New()
	//Logger.SetReportCaller(true)
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(f)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

}

func PrintlnWithHttp(r *http.Request, fname string, msg string) {
	ctx := r.Context()
	//fmt.Println("ctx", ctx)
	reqID := ctx.Value("requestID")
	//fmt.Println("reqID", reqID)
	userID := ctx.Value("userID")
	//fmt.Println("userID", userID, args)
	log.WithFields(log.Fields{
		"app":    app,
		"fname":  fname,
		"reqID":  reqID,
		"userID": userID,
	}).Info(msg)

	/*Logger.Println("app", app,
		"fname", fname,
		"reqID", reqID,
		"userID", userID,
		"msg", msg,
	)*/
}

func ErrorWithHttp(r *http.Request, fname string, msg string) {
	ctx := r.Context()
	//fmt.Println("ctx", ctx)
	reqID := ctx.Value("requestID")
	//fmt.Println("reqID", reqID)
	userID := ctx.Value("userID")
	//fmt.Println("userID", userID, args)
	log.WithFields(log.Fields{
		"app":    app,
		"fname":  fname,
		"reqID":  reqID,
		"userID": userID,
	}).Error(msg)
	/*Logger.Error("app", app,
		"fname", fname,
		"reqID", reqID,
		"userID", userID,
		"msg", msg,
	)*/
}

func Println(fname string, msg string) {
	log.WithFields(log.Fields{
		"app":    app,
		"fname":  fname,
		"reqID":  nil,
		"userID": nil,
	}).Info(msg)
}

func Error(fname string, msg string) {
	log.WithFields(log.Fields{
		"app":    app,
		"fname":  fname,
		"reqID":  nil,
		"userID": nil,
	}).Error(msg)
}

func PrintlnArgs(fname string, args ...interface{}) {
	log.WithFields(log.Fields{
		"app":    app,
		"fname":  fname,
		"reqID":  nil,
		"userID": nil,
	}).Info(args)
}
