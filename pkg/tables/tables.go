package tables

type Metrics struct {
	Id                  int64 `keys:"pk"`
	Hostname            int64
	Port                int16
	Tmetric             int64
	Action              string
	Descr               string
	Warning             float64
	Error               float64
	Dbsid               string
	Username            string
	Password            string
	Startm              bool
	Intrvlnotcrucialhrs int16
	Mname               string
	Tresholdismin       bool
}

type Emails struct {
	Id    int64 `keys:"pk"`
	Email string
	Fio   string
	Descr string
}

type Metrics_emails struct {
	Id       int64 `keys:"pk"`
	Metricid int64
	Emailid  int64
}

type Users struct {
	Id       int64 `keys:"pk"`
	Fio      string
	Username string
	Passw    string
	DtStart  string
}

type Resrcs struct {
	Id    int64 `keys:"pk"`
	Resrc string
	Descr string
}

type User_resrc struct {
	Id      int64 `keys:"pk"`
	Usrid   int64
	Resrcid int64
}

type Roles struct {
	Id    int64 `keys:"pk"`
	Role  string
	Descr string
}

type Role_resrc struct {
	Id      int64 `keys:"pk"`
	Roleid  int64
	Resrcid int64
}

type User_role struct {
	Id     int64 `keys:"pk"`
	Userid int64
	Roleid int64
}

type Tmetrics struct {
	Id    int64 `keys:"pk"`
	Tname string
}

type Srvgrps struct {
	Id      int64 `keys:"pk"`
	Grpname string
}

type Hostnames struct {
	Id       int64 `keys:"pk"`
	Hostname string
	Descr    string
	Srvgrpid int64
}
