package main

import (
	"github.com/fjl/go-couchdb"
	"github.com/kpmy/ypk/halt"
	"log"
)

const dbUrl = "http://127.0.0.1:5984"
const dbName = "stats"
const docId = "total"

type CStatDoc struct {
	Total int
	Data  map[string]int
}

var db *couchdb.DB

func GetStat() (ret *CStatDoc, err error) {
	ret = &CStatDoc{}
	if err = db.Get(docId, ret, nil); err == nil {
		if ret.Data == nil {
			ret.Data = make(map[string]int)
		}
	} else if couchdb.NotFound(err) {
		if _, err = db.Put(docId, ret, ""); err == nil {
			ret, err = GetStat()
		}
	}
	return
}

func SetStat(old *CStatDoc) {
	if rev, err := db.Rev(docId); err == nil {
		if _, err = db.Put(docId, old, rev); err != nil {
			log.Println(err)
		}
	}
}

func IncStat(user string) {
	if s, err := GetStat(); err == nil {
		if _, ok := s.Data[user]; ok {
			s.Data[user] = s.Data[user] + 1
		} else {
			s.Data[user] = 1
		}
		s.Total++
		SetStat(s)
	}
}

func init() {
	if client, err := couchdb.NewClient(dbUrl, nil); err == nil {
		db, _ = client.CreateDB(dbName)
	} else {
		halt.As(100, err)
	}

}
