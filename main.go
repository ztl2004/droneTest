package main

import (
  "github.com/arkors/update/handler"
  "github.com/arkors/update/model"
  "github.com/go-martini/martini"
  "github.com/go-xorm/xorm"
  "github.com/martini-contrib/binding"
  "github.com/martini-contrib/render"
  "log"
  "net/http"
)

const url = "/v1/updates/"

var db *xorm.Engine

func init() {
  var err error
  db, err = xorm.NewEngine("mysql", "arkors:arkors@/arkors_update?charset=utf8")
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

/*func Pool() martini.Handler{
  return func(c martini.Context){
  c.Map(pool)
  }
}*/
func main() {
  m := martini.Classic()
  m.Use(Db())
  m.Use(render.Renderer())
  m.Group("/v1/updates", func(r martini.Router) {
    m.Get("/:app/:version", handler.GetVersion)
    m.Post("/:app", binding.Json(model.Version{}), handler.CreateVersion)
    m.Put("/:app/:version", handler.PutVersion)
    m.Delete("/:app/:version", handler.DelVersion)
  })
  http.ListenAndServe(":3000", m)
}
