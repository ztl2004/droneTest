package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/arkors/log/logmodel"
	"github.com/arkors/oauth/utils"
	"github.com/arkors/update/handler"
	"github.com/arkors/update/model"
	"github.com/go-martini/martini"
	"github.com/go-xorm/xorm"
	"github.com/hoisie/redis"
	"github.com/martini-contrib/render"
)

var db *xorm.Engine
var redisClient redis.Client
var logChanP *chan string
var sys_log_level string

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

	logChan := make(chan string, 1024)
	logChanP = &logChan

	//读取日志配置文件
	var config map[string]string
	config, err = utils.LoadConfig("update.conf")
	if err != nil {
		log.Fatalf("Fail to read configuration : %v\n", err)
	}
	log_level_config := config["LOG_LEVEL"]
	sys_log_level = log_level_config
}

func Db() martini.Handler {
	return func(c martini.Context) {
		c.Map(db)
	}
}

func RedisDb() martini.Handler {
	return func(c martini.Context) {
		redisClient.Addr = "127.0.0.1:6379"
		c.Map(redisClient)
	}
}

func InitParams() martini.Handler {
	return func(c martini.Context) {
		c.Map(logChanP)
		c.Map(sys_log_level)
	}
}

func VerifyJSONBody() martini.Handler {
	return func(c martini.Context, w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if len(data) == 0 {
			if r.Method == "GET" || r.Method == "DELETE" {

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
	//启动发送日志进程
	go logmodel.Sendlog(logChanP)

	m := martini.Classic()
	m.Use(Db())
	m.Use(VerifyJSONBody())
	m.Use(VerifyHTTPHeader())
	m.Use(RedisDb())
	m.Use(InitParams())
	m.Use(render.Renderer())
	m.Group("/v1/updates", func(r martini.Router) {
		m.Get("/:app/:version", handler.GetVersion)
		m.Post("/:app", handler.CreateVersion)
		m.Put("/:app/:version", handler.UpdateVersion)
		m.Delete("/:app/:version", handler.DelVersion)
	})
	http.ListenAndServe(":3001", m)
}
