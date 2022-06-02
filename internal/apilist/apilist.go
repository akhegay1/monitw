package apilist

import (
	"monitw/internal/changeMyPwd"
	"monitw/internal/charts"
	"monitw/internal/emails"
	"monitw/internal/hostnames"
	"monitw/internal/metrics"
	"monitw/internal/metrics_emails"
	"monitw/internal/reports"
	"monitw/internal/resrc"
	"monitw/internal/role_resrc"
	"monitw/internal/roles"
	"monitw/internal/srvgrps"
	"monitw/internal/tmetrics"
	"monitw/internal/user_resrc"
	"monitw/internal/user_role"
	"monitw/internal/users"
	"monitw/internal/vmetrics"
	"net/http"
)

///////////////////////   API   ////////////////////
///////////////////TMetrics/////////////////////////
func ViewTmetrics(w http.ResponseWriter, r *http.Request) {
	tmetrics.ViewTmetrics(w, r)
}
func ViewTmetric(w http.ResponseWriter, r *http.Request) {
	tmetrics.ViewTmetric(w, r)
}
func CreateTmetric(w http.ResponseWriter, r *http.Request) {
	tmetrics.CreateTmetric(w, r)
}
func UpdateTmetric(w http.ResponseWriter, r *http.Request) {
	tmetrics.UpdateTmetric(w, r)
}
func DeleteTmetric(w http.ResponseWriter, r *http.Request) {
	tmetrics.DeleteTmetric(w, r)
}
func SelTmetricsShrt(w http.ResponseWriter, r *http.Request) {
	tmetrics.SelTmetricsShrt(w, r)
}

///////////////////Srvgrps/////////////////////////
func ViewSrvgrps(w http.ResponseWriter, r *http.Request) {
	srvgrps.ViewSrvgrps(w, r)
}
func ViewSrvgrp(w http.ResponseWriter, r *http.Request) {
	srvgrps.ViewSrvgrp(w, r)
}
func CreateSrvgrp(w http.ResponseWriter, r *http.Request) {
	srvgrps.CreateSrvgrp(w, r)
}
func UpdateSrvgrp(w http.ResponseWriter, r *http.Request) {
	srvgrps.UpdateSrvgrp(w, r)
}
func DeleteSrvgrp(w http.ResponseWriter, r *http.Request) {
	srvgrps.DeleteSrvgrp(w, r)
}
func SelSrvgrpsShrt(w http.ResponseWriter, r *http.Request) {
	srvgrps.SelSrvgrpsShrt(w, r)
}

///////////////////Hostnames/////////////////////////

func ViewHostnames(w http.ResponseWriter, r *http.Request) {
	hostnames.ViewHostnames(w, r)
}
func ViewHostname(w http.ResponseWriter, r *http.Request) {
	hostnames.ViewHostname(w, r)
}
func CreateHostname(w http.ResponseWriter, r *http.Request) {
	hostnames.CreateHostname(w, r)
}
func UpdateHostname(w http.ResponseWriter, r *http.Request) {
	hostnames.UpdateHostname(w, r)
}
func DeleteHostname(w http.ResponseWriter, r *http.Request) {
	hostnames.DeleteHostname(w, r)
}
func SelHostnamesShrt(w http.ResponseWriter, r *http.Request) {
	hostnames.SelHostnamesShrt(w, r)
}

///////////////////Metrics/////////////////////////
func ViewMetrics(w http.ResponseWriter, r *http.Request) {
	metrics.ViewMetrics(w, r)
}
func ViewMetric(w http.ResponseWriter, r *http.Request) {
	metrics.ViewMetric(w, r)
}
func CreateMetric(w http.ResponseWriter, r *http.Request) {
	metrics.CreateMetric(w, r)
}
func UpdateMetric(w http.ResponseWriter, r *http.Request) {
	metrics.UpdateMetric(w, r)
}
func DeleteMetric(w http.ResponseWriter, r *http.Request) {
	metrics.DeleteMetric(w, r)
}

///////////////////Emails/////////////////////////

func ViewEmails(w http.ResponseWriter, r *http.Request) {
	emails.ViewEmails(w, r)
}
func ViewEmail(w http.ResponseWriter, r *http.Request) {
	emails.ViewEmail(w, r)
}
func CreateEmail(w http.ResponseWriter, r *http.Request) {
	emails.CreateEmail(w, r)
}
func UpdateEmail(w http.ResponseWriter, r *http.Request) {
	emails.UpdateEmail(w, r)
}
func DeleteEmail(w http.ResponseWriter, r *http.Request) {
	emails.DeleteEmail(w, r)
}
func SelEmailsShrt(w http.ResponseWriter, r *http.Request) {
	emails.SelEmailsShrt(w, r)
}

///////////////////Vmetrics/////////////////////////
func ViewVmetrics(w http.ResponseWriter, r *http.Request) {
	vmetrics.ViewVmetrics(w, r)
}

func ViewVmetricsAllWS(w http.ResponseWriter, r *http.Request) {
	vmetrics.ViewVmetricsAllWS(w, r)
}

///////////////////Charts/////////////////////////
func GetVmetricsBydt(w http.ResponseWriter, r *http.Request) {
	charts.GetVmetricsBydt(w, r)
}

///////////////////MertricsEmails/////////////////////////

func ViewEmailsByMetricsId(w http.ResponseWriter, r *http.Request) {
	metrics_emails.ViewEmailsByMetricsId(w, r)
}
func ViewEmails4Metric(w http.ResponseWriter, r *http.Request) {
	metrics_emails.ViewEmails4Metric(w, r)
}
func CreateMetricEmail(w http.ResponseWriter, r *http.Request) {
	metrics_emails.CreateMetricEmail(w, r)
}
func DeleteMetricsEmails(w http.ResponseWriter, r *http.Request) {
	metrics_emails.DeleteMetricsEmails(w, r)
}

///////////////////Users/////////////////////////

func ViewUsers(w http.ResponseWriter, r *http.Request) {
	users.ViewUsers(w, r)
}
func ViewUser(w http.ResponseWriter, r *http.Request) {
	users.ViewUser(w, r)
}
func CreateUser(w http.ResponseWriter, r *http.Request) {
	users.CreateUser(w, r)
}
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	users.UpdateUser(w, r)
}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	users.DeleteUser(w, r)
}

///////////////////Resrc/////////////////////////

func ViewResrcs(w http.ResponseWriter, r *http.Request) {
	resrc.ViewResrcs(w, r)
}
func ViewResrc(w http.ResponseWriter, r *http.Request) {
	resrc.ViewResrc(w, r)
}
func CreateResrc(w http.ResponseWriter, r *http.Request) {
	resrc.CreateResrc(w, r)
}
func UpdateResrc(w http.ResponseWriter, r *http.Request) {
	resrc.UpdateResrc(w, r)
}
func DeleteResrc(w http.ResponseWriter, r *http.Request) {
	resrc.DeleteResrc(w, r)
}

///////////////////User_Resrc/////////////////////////

func ViewResrcs4UserNotGranted(w http.ResponseWriter, r *http.Request) {
	user_resrc.ViewResrcs4UserNotGranted(w, r)
}
func ViewResrcsByUserId(w http.ResponseWriter, r *http.Request) {
	user_resrc.ViewResrcsByUserId(w, r)
}
func CreateUserResrc(w http.ResponseWriter, r *http.Request) {
	user_resrc.CreateUserResrc(w, r)
}
func DeleteUserResrc(w http.ResponseWriter, r *http.Request) {
	user_resrc.DeleteUserResrc(w, r)
}

///////////////////Roles/////////////////////////

func ViewRoles(w http.ResponseWriter, r *http.Request) {
	roles.ViewRoles(w, r)
}
func ViewRole(w http.ResponseWriter, r *http.Request) {
	roles.ViewRole(w, r)
}
func CreateRole(w http.ResponseWriter, r *http.Request) {
	roles.CreateRole(w, r)
}
func UpdateRole(w http.ResponseWriter, r *http.Request) {
	roles.UpdateRole(w, r)
}
func DeleteRole(w http.ResponseWriter, r *http.Request) {
	roles.DeleteRole(w, r)
}

///////////////////Role_Resrc/////////////////////////

func ViewResrcs4RoleNotGranted(w http.ResponseWriter, r *http.Request) {
	role_resrc.ViewResrcs4RoleNotGranted(w, r)
}
func ViewResrcsByRoleId(w http.ResponseWriter, r *http.Request) {
	role_resrc.ViewResrcsByRoleId(w, r)
}
func CreateRoleResrc(w http.ResponseWriter, r *http.Request) {
	role_resrc.CreateRoleResrc(w, r)
}
func DeleteRoleResrc(w http.ResponseWriter, r *http.Request) {
	role_resrc.DeleteRoleResrc(w, r)
}

///////////////////User_role/////////////////////////

func ViewRoless4UserNotGranted(w http.ResponseWriter, r *http.Request) {
	user_role.ViewRoles4UserNotGranted(w, r)
}
func ViewRolesByUserId(w http.ResponseWriter, r *http.Request) {
	user_role.ViewRolesByUserId(w, r)
}
func ViewRolesByUserid(w http.ResponseWriter, r *http.Request) {
	user_role.ViewRolesByUserid(w, r)
}
func CreateUserRole(w http.ResponseWriter, r *http.Request) {
	user_role.CreateUserRole(w, r)
}
func DeleteUserRole(w http.ResponseWriter, r *http.Request) {
	user_role.DeleteUserRole(w, r)
}

///////////////////ChangeMyPwd/////////////////////////
func ChangeMyPwd(w http.ResponseWriter, r *http.Request) {
	changeMyPwd.ChangeMyPwd(w, r)
}

///////////////////Reports/////////////////////////
func F_rep_001(w http.ResponseWriter, r *http.Request) {
	reports.F_rep_001(w, r)
}
