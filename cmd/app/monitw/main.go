package main

import (
	"monitw/internal/apilist"
	"monitw/internal/auth"
	"monitw/internal/db"
	req "monitw/internal/requestid"
	"monitw/pkg/mwutils"
	"net/http"
	"os"

	"time"

	log "monitw/internal/logger"

	_ "net/http/pprof"

	_ "github.com/lib/pq"
)

var errorlog *os.File
var words []string

//ppp
func init() {
	log.Println(mwutils.FuncName(), "started")
	defer log.Println(mwutils.FuncName(), "finished")

	c := db.Connect()
	log.PrintlnArgs(mwutils.FuncName(), "connect ", c)
	auth.LoadPermissions()

}

type Middleware func(http.Handler) http.Handler

func MultipleMiddleware(h http.Handler, m ...Middleware) http.Handler {

	if len(m) < 1 {
		return h
	}

	wrapped := h

	// loop in reverse to preserve middleware order
	for i := len(m) - 1; i >= 0; i-- {
		wrapped = m[i](wrapped)
	}

	return wrapped

}

func checkDbChanges() {
	//log.Logger.Println("Start checkDbChanges")
	if auth.CheckPermIsChanged() {
		auth.LoadPermissions()
	}
}

func main() {
	log.Println(mwutils.FuncName(), "started")
	defer log.Println(mwutils.FuncName(), "finished")

	quit := make(chan bool, 1)

	go func() {
		checkDbChanges()

		for {
			select {
			case <-quit:
				return
			case <-time.After(10 * time.Second):
				checkDbChanges()
			}
		}
	}()

	//authorization page, moddleware is not neccessary
	http.HandleFunc("/auth", auth.TokenHandler)
	//arr of middleware functions
	commonMiddleware := []Middleware{
		req.ReqIDMiddleware1,
		//AuthMiddleware,
	}

	//map for urls and handlers
	endpoints := map[string]http.HandlerFunc{
		"/metrics":        apilist.ViewMetrics,
		"/metric":         apilist.ViewMetric,
		"/metrics/create": apilist.CreateMetric,
		"/metrics/update": apilist.UpdateMetric,
		"/metrics/delete": apilist.DeleteMetric,

		"/hostnames":        apilist.ViewHostnames,
		"/hostname":         apilist.ViewHostname,
		"/hostnames/create": apilist.CreateHostname,
		"/hostnames/update": apilist.UpdateHostname,
		"/hostnames/delete": apilist.DeleteHostname,
		"/shostnames":       apilist.SelHostnamesShrt,

		"/tmetrics":        apilist.ViewTmetrics,
		"/tmetric":         apilist.ViewTmetric,
		"/tmetrics/create": apilist.CreateTmetric,
		"/tmetrics/update": apilist.UpdateTmetric,
		"/tmetrics/delete": apilist.DeleteTmetric,
		"/stmetrics":       apilist.SelTmetricsShrt,

		"/srvgrps":        apilist.ViewSrvgrps,
		"/srvgrp":         apilist.ViewSrvgrp,
		"/srvgrps/create": apilist.CreateSrvgrp,
		"/srvgrps/update": apilist.UpdateSrvgrp,
		"/srvgrps/delete": apilist.DeleteSrvgrp,
		"/ssrvgrps":       apilist.SelSrvgrpsShrt,

		"/emails":        apilist.ViewEmails,
		"/email":         apilist.ViewEmail,
		"/emails/create": apilist.CreateEmail,
		"/emails/update": apilist.UpdateEmail,
		"/emails/delete": apilist.DeleteEmail,
		"/semails":       apilist.SelEmailsShrt,

		"/vmetrics": apilist.ViewVmetrics,

		"/vmetrics/getbydt": apilist.GetVmetricsBydt,

		"/metrics_emails":        apilist.ViewEmailsByMetricsId,
		"/metrics_emails/emails": apilist.ViewEmails4Metric,
		"/metrics_emails/create": apilist.CreateMetricEmail,
		"/metrics_emails/delete": apilist.DeleteMetricsEmails,

		"/users":        apilist.ViewUsers,
		"/user":         apilist.ViewUser,
		"/users/create": apilist.CreateUser,
		"/users/update": apilist.UpdateUser,
		"/users/delete": apilist.DeleteUser,

		"/resrcs":        apilist.ViewResrcs,
		"/resrc":         apilist.ViewResrc,
		"/resrcs/create": apilist.CreateResrc,
		"/resrcs/update": apilist.UpdateResrc,
		"/resrcs/delete": apilist.DeleteResrc,

		"/user_resrcs_notgranted": apilist.ViewResrcs4UserNotGranted,
		"/user_resrcsbyid":        apilist.ViewResrcsByUserId,
		"/user_resrc/create":      apilist.CreateUserResrc,
		"/user_resrc/delete":      apilist.DeleteUserResrc,

		"/roles":        apilist.ViewRoles,
		"/role":         apilist.ViewRole,
		"/roles/create": apilist.CreateRole,
		"/roles/update": apilist.UpdateRole,
		"/roles/delete": apilist.DeleteRole,

		"/role_resrcs_notgranted": apilist.ViewResrcs4RoleNotGranted,
		"/role_resrcsbyid":        apilist.ViewResrcsByRoleId,
		"/role_resrc/create":      apilist.CreateRoleResrc,
		"/role_resrc/delete":      apilist.DeleteRoleResrc,

		"/user_role_notgranted": apilist.ViewRoless4UserNotGranted,
		"/user_rolesbyid":       apilist.ViewRolesByUserId,
		"/user_rolesbyuserid":   apilist.ViewRolesByUserid,
		"/user_role/create":     apilist.CreateUserRole,
		"/user_role/delete":     apilist.DeleteUserRole,
		"/changeMyPwd":          apilist.ChangeMyPwd,
		"/rep_001":              apilist.F_rep_001,
	}

	for endpoint, f := range endpoints {
		http.Handle(endpoint, auth.AuthMiddleware(endpoint, MultipleMiddleware(f, commonMiddleware...)))
	}
	http.HandleFunc("/vmetricsAll", apilist.ViewVmetricsAllWS)
	//http.HandleFunc("/rep_001", apilist.F_rep_001)

	server := &http.Server{
		Addr:         ":3010",
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	//http.ListenAndServe(":3010", nil)
	server.ListenAndServe()
	log.Println(mwutils.FuncName(), "Aft Listen ")
}
