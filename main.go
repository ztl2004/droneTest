package main

import (
	"encoding/json"
	"github.com/arkors/update/handler"
	"github.com/arkors/update/model"
	"github.com/go-martini/martini"
	"github.com/go-xorm/xorm"
	"github.com/martini-contrib/render"
	"io/ioutil"
	"log"
	"net/http"
)

var db *xorm.Engine

func init() {
	var err error

  db, err = xorm.NewEngine("mysql", "arkors:arkors@/arkors_update?charset=utf8")
  db.ShowSQL = true

	if err != nil {
		log.Fatalf("Fail to create engine: %v\n", err)
	}

	if err = db.Sync(new(model.Version)); err != nil {
		log.Fatalf("Fail to sync database: %v\n", err)
	}
}

func Db() martini.Handler {
	return func(c martini.Context) {
		c.Map(db)
	}
}

func VerifyJSONBody() martini.Handler {
	return func(c martini.Context, w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if len(data) == 0 {
			if r.Method == "GET" || r.Method == "DELETE"{

			} else {
				return
			}
		} else {
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("{\"error\":\"Invalid request body.\"}"))
			}
			var version model.Version
			json2versionErr := json.Unmarshal(data, &version)
			if json2versionErr != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("{\"error\":\"can't trans to version\"}"))
			} else {
				c.Map(version)
			}
		}
	}

}

func VerifyHTTPHeader() martini.Handler {
	return func(c martini.Context, w http.ResponseWriter, r *http.Request) {
		_, log := r.Header["X-Arkors-Application-Log"]
		_, client := r.Header["X-Arkors-Application-Client"]
		if log != true || client != true {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{\"error\":\"Invalid request header, it should be include 'X-Arkors-Application-Log'  and 'X-Arkors-Application-Client'.\"}"))
		}
	}
}

func main() {
	m := martini.Classic()
	m.Use(Db())
	m.Use(VerifyJSONBody())
	m.Use(VerifyHTTPHeader())
	m.Use(render.Renderer())
	m.Group("/v1/updates", func(r martini.Router) {
		m.Get("/:app/:version", handler.GetVersion)
		m.Post("/:app", handler.CreateVersion)
		m.Put("/:app/:version", handler.UpdateVersion)
		m.Delete("/:app/:version", handler.DelVersion)
	})
	http.ListenAndServe(":3000", m)
}
