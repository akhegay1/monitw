package auth

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"monitw/internal/crypto"
	"monitw/internal/db"
	"monitw/pkg/mutils"
	"monitw/pkg/mwutils"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"monitw/pkg/jwtserver"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	log "monitw/internal/logger"
)

type Allprivs struct {
	Key string
}

var AllprivsMap = make(map[string]struct{})

var jwthost string
var Auth bool
var jwtparams []string
var authparams []string
var conn *grpc.ClientConn

// AttachRequestID will attach a brand new request ID to a http request
func AssignUserID(ctx context.Context, user string) context.Context {
	log.Println(mwutils.FuncName(), "started")
	defer log.Println(mwutils.FuncName(), "finished")

	userID := user
	log.Println(mwutils.FuncName(), "userID="+userID)
	return context.WithValue(ctx, "userID", userID)
}

// GetRequestID will get reqID from a http request and return it as a string
func GetUserID(ctx context.Context) string {
	log.Println(mwutils.FuncName(), "started")
	defer log.Println(mwutils.FuncName(), "finished")

	userID := ctx.Value("userID")
	if ret, ok := userID.(string); ok {
		log.Println(mwutils.FuncName(), "ret="+ret)
		return ret
	}

	return ""
}

func init() {
	log.Println(mwutils.FuncName(), "started")
	defer log.Println(mwutils.FuncName(), "finished")
	////////////////FILE confJWT/////
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
		if parnam == "auth" {
			val := str[strings.Index(str, "=")+1:]
			authparams = append(authparams, val)
			break
		}
	}

	if err := sc.Err(); err != nil {
		log.Println(mwutils.FuncName(), "scan file error: "+err.Error())
	}

	Auth, _ = strconv.ParseBool(authparams[0])
	log.Println(mwutils.FuncName(), "Auth:"+strconv.FormatBool(Auth))

	//////////////////

	grpcConnect()
}

func grpcConnect() {
	log.Println(mwutils.FuncName(), "started")
	defer log.Println(mwutils.FuncName(), "finished")

	////////////////FILE confJWT/////
	conf, err := os.Open("params")
	if err != nil {
		log.Println(mwutils.FuncName(), "failed opening file confJWT:s"+err.Error())
	}
	defer conf.Close()

	sc := bufio.NewScanner(conf)

	for sc.Scan() {
		str := sc.Text() // GET the line string
		if len(str) == 0 {
			continue
		} else if str[0:1] == "#" {
			continue
		}
		// /fmt.Println("str", str[0:strings.Index(str, "=")])
		parnam := str[0:strings.Index(str, "=")]
		if parnam == "jwthost" {
			val := str[strings.Index(str, "=")+1:]
			jwtparams = append(jwtparams, val)
			break
		}
	}
	//fmt.Println("jwtparams", jwtparams)
	if err := sc.Err(); err != nil {
		log.Println(mwutils.FuncName(), "scan file error: "+err.Error())
	}

	jwthost = jwtparams[0]
	log.Println(mwutils.FuncName(), "jwthost:"+jwthost)

	//////////////////

	conn, err = grpc.Dial(jwthost+":9011", grpc.WithInsecure())
	if err != nil {
		log.Println(mwutils.FuncName(), "did not connect: "+err.Error())
	}
	//defer conn.Close()

	log.Println(mwutils.FuncName(), "aft grpc conect")

}

// TokenHandler is our handler to take a username and password and,
// if it's valid, return a token used for future requests.
func TokenHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(mwutils.FuncName(), "started")
	defer log.Println(mwutils.FuncName(), "finished")

	mwutils.SetupResponse(&w, r)
	w.Header().Add("Content-Type", "application/json")

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	//log.Logger.Println(string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Error(mwutils.FuncName(), err.Error())

	}
	type Login struct {
		Username string
		Password string
	}
	var login Login
	err = json.Unmarshal(b, &login)

	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Error(mwutils.FuncName(), err.Error())
		return
	}

	username := login.Username
	password := login.Password

	///////////////Get from DB////////
	var userid string
	var userdb string
	var passwdb string
	err = db.Db.QueryRow("SELECT id, username,passw FROM monit_sch.users WHERE username=$1", username).Scan(&userid, &userdb, &passwdb)
	log.Println(mwutils.FuncName(), "userid="+userid)
	log.Println(mwutils.FuncName(), "userdb="+userdb)
	log.Println(mwutils.FuncName(), "passwdb="+passwdb)
	passwdb = crypto.Decrypt(mutils.Key, passwdb)
	log.Println(mwutils.FuncName(), "passwdb="+passwdb)
	if err != nil || len(userdb) == 0 {
		log.Println(mwutils.FuncName(), err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":"invalid_credentials"}`)
		return
	}

	if username != userdb || password != passwdb {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":"invalid_credentials"}`)
		return
	}

	//////////////////grpc
	c := jwtserver.NewJwtServerServiceClient(conn)

	response, err := c.GetToken(context.Background(), &jwtserver.Reqtoken{User: userid})
	if err != nil {
		log.Println(mwutils.FuncName(), "Error when calling SayHello: "+err.Error())
	}

	log.Println(mwutils.FuncName(), "Response from server: "+response.TokenString)

	///
	//log.Logger.Println("tokenString",  response.TokenString)
	io.WriteString(w, `{"token":"`+response.TokenString+`"}`)
	return
}

// AuthMiddleware is our middleware to check our token is valid. Returning
// a 401 status to the client if it is not valid.
func AuthMiddleware(resrc string, next http.Handler) http.Handler {
	log.Println(mwutils.FuncName(), "started")
	defer log.Println(mwutils.FuncName(), "finished")

	if Auth == false {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { next.ServeHTTP(w, r) })
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		mwutils.SetupResponse(&w, r)
		if r.Method == "OPTIONS" {
			log.Println(mwutils.FuncName(), "OPTIONS")
			log.PrintlnArgs(mwutils.FuncName(), r.Header)
			mwutils.SetupResponse(&w, r)
			return
		}

		const BEARER_SCHEMA = "Bearer "
		authHeader := r.Header.Get("Authorization")
		log.Println(mwutils.FuncName(), "authHeader="+authHeader)

		if authHeader != "" {

			reqToken := authHeader[len(BEARER_SCHEMA):]
			log.Println(mwutils.FuncName(), "reqToken="+reqToken)

			c := jwtserver.NewJwtServerServiceClient(conn)
			response, err := c.CheckToken(context.Background(), &jwtserver.CheckAuth{TokenString: reqToken})

			str1 := fmt.Sprintf("Response from server: %t %s", response.Tokenvalid, response.User)
			log.Println(mwutils.FuncName(), str1)
			ctx2 := AssignUserID(ctx, response.User)
			r = r.WithContext(ctx2)
			if err != nil {
				log.Println(mwutils.FuncName(), "Error when calling SayHello: "+err.Error())
			}

			var hasPermission bool
			hasPermission = CheckPermission(response.User, resrc)
			log.Println(mwutils.FuncName(), "hasPermission="+strconv.FormatBool(hasPermission))
			if response.Tokenvalid && hasPermission {
				next.ServeHTTP(w, r)
			} else if !hasPermission {
				http.Error(w, "Permission denied ", 500)
			}

		} else {
			http.Error(w, "No Authorization Token provided", 500)
		}

	})

}

func CheckPermission(user string, resrc string) bool {
	log.Println(mwutils.FuncName(), "user="+string(user)+"resrc="+resrc)
	log.Println(mwutils.FuncName(), "user+resrc "+user+resrc)
	//эта проверка какие роли есть у юзера, она нужна всем, пойиг какие роли
	if resrc == "/user_rolesbyuserid" {
		return true
	}

	//log.Logger.Println(mwutils.FuncName(),"AllprivsMap="+ AllprivsMap)

	//log.PrintlnArgs(mwutils.FuncName(), AllprivsMap)

	if _, ok := AllprivsMap[user+resrc]; ok {
		return true
	}

	return false
}

func LoadPermissions() {
	log.Println(mwutils.FuncName(), "started")
	defer log.Println(mwutils.FuncName(), "finished")

	sqlQuery := `
	select u.id||b.resrc as "key" from monit_sch.user_resrc ur
	join monit_sch.resrcs b on ur.resrcid = b.id 
	join monit_sch.users u on ur.usrid = u.id 
	union
	select  u.id||b.resrc as "key" from monit_sch.user_role ur
	join monit_sch.roles c on ur.roleid = c.id
	join monit_sch.users u on ur.userid = u.id  
	join monit_sch.role_resrc rr on rr.roleid =c.id
	join monit_sch.resrcs b on rr.resrcid = b.id`

	rows, err := db.Db.Query(sqlQuery)
	if err != nil {
		log.Error(mwutils.FuncName(), err.Error())
		return
	}

	ml := sync.RWMutex{}

	for rows.Next() {
		m := new(Allprivs)
		err = rows.Scan(&m.Key)
		if err != nil {
			log.Println(mwutils.FuncName(), err.Error())
			return
		}
		//log.Logger.Println("LoadPermissions m", m)
		ml.RLock()
		AllprivsMap[m.Key] = struct{}{}
		ml.RUnlock()
	}
	//log.Logger.Println(mwutils.FuncName(), AllprivsArr)

	txn, _ := db.Db.Begin()
	_, err = db.Db.Exec("update monit_sch.privs_changed set is_changed=false")

	if err != nil {
		log.Error(mwutils.FuncName(), err.Error())
		txn.Rollback()
		return
	}

	err = txn.Commit()
	log.Println(mwutils.FuncName(), "aft update commit ")

	return
}

func CheckPermIsChanged() bool {
	var isChanged bool
	sqlStatement := `select is_changed from monit_sch.privs_changed;`
	err := db.Db.QueryRow(sqlStatement).Scan(&isChanged)

	//log.Logger.Println("CheckPermIsChanged AllprivsMap", AllprivsMap)

	if err != nil || !isChanged {
		return false
	}

	return true
}
