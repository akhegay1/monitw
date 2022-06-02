package changeMyPwd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"monitw/internal/crypto"
	"monitw/internal/db"
	"monitw/pkg/mutils"
	"monitw/pkg/mwutils"
	"net/http"

	log "monitw/internal/logger"
)

type ChangePwd struct {
	NewPwd1 string
	NewPwd2 string
}

func ChangeMyPwd(w http.ResponseWriter, r *http.Request) {
	log.PrintlnWithHttp(r, mwutils.FuncName(), "started")
	defer log.PrintlnWithHttp(r, mwutils.FuncName(), "finished")

	mwutils.SetupResponse(&w, r)
	if r.Method == "OPTIONS" {
		log.PrintlnWithHttp(r, mwutils.FuncName(), "OPTIONS")
		return
	}

	if r.Method != "POST" {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(405), 405)
		log.ErrorWithHttp(r, mwutils.FuncName(), http.StatusText(405))
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	log.PrintlnWithHttp(r, mwutils.FuncName(), string(b))

	defer r.Body.Close()
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())

	}

	var chgPwd ChangePwd
	err = json.Unmarshal(b, &chgPwd)

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	//log.PrintlnWithHttp(r, "chgPwd.NewPwd1", chgPwd.NewPwd1, "chgPwd.NewPwd2", chgPwd.NewPwd2)
	if chgPwd.NewPwd1 == "" {
		http.Error(w, mwutils.FuncName()+": "+"NewPwd1 is empty", 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), "chgPwd.NewPwd1 is empty")
		return
	}
	if chgPwd.NewPwd2 == "" {
		http.Error(w, mwutils.FuncName()+": "+"NewPwd2 is empty", 400)
		log.ErrorWithHttp(r, mwutils.FuncName(), "chgPwd.NewPwd2 is empty")
		return
	}
	if chgPwd.NewPwd1 != chgPwd.NewPwd2 {
		http.Error(w, mwutils.FuncName()+": "+"Entered passwords are different", 400)
		log.ErrorWithHttp(r, mwutils.FuncName(), "Entered passwords are different")
		return
	}

	//log.PrintlnWithHttp(r, mwutils.FuncName(), "chgPwd.NewPwd " + chgPwd.NewPwd1)
	chgPwd.NewPwd1 = crypto.Encrypt(mutils.Key, chgPwd.NewPwd1)
	//log.PrintlnWithHttp(r, mwutils.FuncName(), "chgPwd.NewPwd "+chgPwd.NewPwd1)

	ctx := r.Context()
	userID := ctx.Value("userID")

	txn, _ := db.Db.Begin()

	var lastUpdateId int64
	updq := fmt.Sprintf(`update monit_sch.Users set passw = '%s' 
							where id = %s returning id`, chgPwd.NewPwd1, userID)

	log.PrintlnWithHttp(r, "updq", updq)

	err = db.Db.QueryRow(updq).Scan(&lastUpdateId)

	log.PrintlnWithHttp(r, mwutils.FuncName(), "after update ")
	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+http.StatusText(500)+" "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		txn.Rollback()
		return
	}

	err = txn.Commit()
	log.PrintlnWithHttp(r, mwutils.FuncName(), "aft Commit")

	if err != nil {
		http.Error(w, mwutils.FuncName()+": "+err.Error(), 500)
		log.ErrorWithHttp(r, mwutils.FuncName(), err.Error())
		return
	}

	fmt.Fprintf(w, "%d", lastUpdateId)

}
