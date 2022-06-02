package crud

import (
	"fmt"
	log "monitw/internal/logger"
	"reflect"
	"strings"
	"time"
)

func CreateInsert(q interface{}, user string) string {
	log.Println("CreateInsert", "userID from CreateInsert="+user)
	var fnames, fvalues string
	if reflect.ValueOf(q).Kind() == reflect.Struct {
		t := reflect.TypeOf(q).Name()
		nam := reflect.Indirect(reflect.ValueOf(q))

		v := reflect.ValueOf(q)
		for i := 0; i < v.NumField(); i++ {
			if reflect.TypeOf(q).Field(i).Tag.Get("keys") != "pk" {
				fnames = fnames + ", " + nam.Type().Field(i).Name
			}
		}
		fnames = "(" + fnames[2:] + " ,author" + ")"

		query := fmt.Sprintf("insert into monit_sch.%s %s values(", t, fnames)

		for i := 0; i < v.NumField(); i++ {
			switch v.Field(i).Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if reflect.TypeOf(q).Field(i).Tag.Get("keys") != "pk" {
					fvalues = fmt.Sprintf("%s, %d", fvalues, v.Field(i).Int())
				}
			case reflect.String:
				if v.Field(i).String() != "" {
					fvalues = fmt.Sprintf("%s, '%s'", fvalues, strings.Replace(v.Field(i).String(), "'", "''", -1))
				} else {
					fvalues = fmt.Sprintf("%s, %s", fvalues, "null")
				}

			case reflect.Bool:
				fvalues = fmt.Sprintf("%s, %v", fvalues, v.Field(i).Interface().(bool))
			case reflect.Float32, reflect.Float64:
				fvalues = fmt.Sprintf("%s, %.2f", fvalues, v.Field(i).Interface().(float64))
			default:
				return "Unsupported type"
			}
		}

		fvalues = fvalues[2:] + ", " + user
		query = fmt.Sprintf("%s %s) returning id", query, fvalues)
		//fmt.Println(query)
		return query

	}
	return "unsupported type"
}

func CreateUpdate(q interface{}, user string) string {

	if reflect.ValueOf(q).Kind() == reflect.Struct {
		t := reflect.TypeOf(q).Name()
		nam := reflect.Indirect(reflect.ValueOf(q))

		v := reflect.ValueOf(q)

		query := fmt.Sprintf("update monit_sch.%s set ", t)
		ustr := ""

		var wher string
		for i := 0; i < v.NumField(); i++ {

			if reflect.TypeOf(q).Field(i).Tag.Get("keys") == "pk" {
				wher = fmt.Sprintf(" where %s=", nam.Type().Field(i).Name)
				wher = fmt.Sprintf("%s%d", wher, v.Field(i).Int())
			} else {
				switch v.Field(i).Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					ustr = fmt.Sprintf("%s, %s=%d", ustr, nam.Type().Field(i).Name, v.Field(i).Int())
				case reflect.String:
					if v.Field(i).String() != "" {
						ustr = fmt.Sprintf("%s, %s='%s'", ustr, nam.Type().Field(i).Name, strings.Replace(v.Field(i).String(), "'", "''", -1))
					} else {
						ustr = fmt.Sprintf("%s, %s=%s", ustr, nam.Type().Field(i).Name, "null")
					}
				case reflect.Bool:
					ustr = fmt.Sprintf("%s, %s=%v", ustr, nam.Type().Field(i).Name, v.Field(i).Interface().(bool))
				case reflect.Float32, reflect.Float64:
					ustr = fmt.Sprintf("%s, %s=%.2f", ustr, nam.Type().Field(i).Name, v.Field(i).Interface().(float64))
				default:
					return "Unsupported type"
				}
			}
		}

		ustr = ustr[2:] + ", lastupdatedauthor=" + user + ", lastupdates=" + "'" +
			time.Now().UTC().Format("2006-01-02T15:04:05-0700") + "'"
		wher = wher + " and (author=" + "'" + user + "'" + " or " + user +
			" in (select userid from monit_sch.user_role where roleid=13) " + ")" //role admin
		query = fmt.Sprintf("%s%s%s returning id", query, ustr, wher)

		return query

	}
	return "unsupported type"
}

func CreateDelete(tabl string, id string, user string) string {

	query := fmt.Sprintf("delete from monit_sch.%s where id=%s"+
		" and (author="+"'"+user+"'"+" or "+user+
		" in (select userid from monit_sch.user_role where roleid=13)) ", tabl, id)

	return query

}
