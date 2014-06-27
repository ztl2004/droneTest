package main

import (
  "bytes"
  "encoding/json"
  "github.com/arkors/update/handler"
  "github.com/arkors/update/model"
  "github.com/go-martini/martini"
  "github.com/martini-contrib/render"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  "net/http"
  "net/http/httptest"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "github.com/hoisie/redis"
  "github.com/go-xorm/xorm"
  "log"
  "fmt"
)

var xormdb *xorm.Engine


func init(){
  test_db,_:=sql.Open("mysql","arkors_test:arkors_test@/arkors_update_test")
  dberr:=test_db.Ping()
  if dberr==nil && test_db!=nil {
     test_db.Close()
     return
  }
  test_db.Close()
  root_db, err := sql.Open("mysql", "root:@/")
  if(err!=nil){
      fmt.Println("fail to create database connnection %v",err)
      return
  }
  root_db.Exec("CREATE DATABASE `arkors_update_test` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci")
  root_db.Exec("insert into mysql.user(Host,User,Password) values('localhost','arkors_test',password('arkors_test'))")
  root_db.Exec("flush privileges")
  root_db.Exec("grant all privileges on arkors_update_test.* to arkors_test@localhost identified by 'arkors_test'")
  root_db.Close()

  arkors_db,errConn:=sql.Open("mysql","arkors_test:arkors_test@/arkors_update_test")
  if errConn != nil {
      fmt.Println("fail to create database connnection %v",errConn)
      return
  }
  arkors_db.Exec("CREATE TABLE `version` (`id` bigint(20) NOT NULL AUTO_INCREMENT,`app` bigint(20) DEFAULT NULL,`version` int(11) DEFAULT NULL,`name` text,`updated` date DEFAULT NULL,`changed` text,`url` text,`client` text,`compatible` text,PRIMARY KEY (`id`))")
  arkors_db.Exec("INSERT INTO `version` VALUES (1,233,1,'0.3.3','2013-11-10','1. New design application icon. \n 2. Fix some bugs.','http//file.arkors.com/releases/demo-lasttest.apk','Android','0.3.3 0.3.1 0.2.x 0.2.x')")
  arkors_db.Exec("INSERT INTO `version` VALUES (2,234,1,'0.3.3','2013-11-10','1. New design application icon. \n 2. Fix some bugs.','http//file.arkors.com/releases/demo-lasttest.apk','Android','0.3.3 0.3.1 0.2.x 0.2.x')")
  arkors_db.Exec("INSERT INTO `version` VALUES (3,234,2,'0.3.3','2014-11-10','1. New design application icon. \n 2. Fix some bugs.','http//file.arkors.com/releases/demo-lasttest.apk','Android','0.3.3 0.3.1 0.2.x 0.2.x')")
  arkors_db.Close()

  var client redis.Client
  client.Set("233@1",[]byte("{\"Id\":1,\"App\":233,\"Version\":1,\"Name\":\"0.3.3\",\"Updated\":\"2013-11-10 00:00:00\",\"Changed\":\"1. New design application icon. \\n 2. Fix some bugs. \\n \",\"Url\":\"http//file.arkors.com/releases/demo-lasttest.apk\",\"Client\":\"Android\",\"Compatible\":\"0.3.2 0.3.1 0.2.x 0.2.x\"}"))
  client.Set("234@1",[]byte("{\"Id\":2,\"App\":234,\"Version\":1,\"Name\":\"0.3.3\",\"Updated\":\"2013-11-10 00:00:00\",\"Changed\":\"1. New design application icon. \\n 2. Fix some bugs. \\n \",\"Url\":\"http//file.arkors.com/releases/demo-lasttest.apk\",\"Client\":\"Android\",\"Compatible\":\"0.3.2 0.3.1 0.2.x 0.2.x\"}"))
  client.Set("234@2",[]byte("{\"Id\":3,\"App\":234,\"Version\":2,\"Name\":\"0.3.3\",\"Updated\":\"2014-11-10 00:00:00\",\"Changed\":\"1. New design application icon. \\n 2. Fix some bugs. \\n \",\"Url\":\"http//file.arkors.com/releases/demo-lasttest.apk\",\"Client\":\"Android\",\"Compatible\":\"0.3.2 0.3.1 0.2.x 0.2.x\"}"))

  var xormErr error
  xormdb,xormErr = xorm.NewEngine("mysql","arkors_test:arkors_test@/arkors_update_test?charset=utf8")
  if xormErr != nil {
     log.Fatalf("Fail to create test engine: %v\n", xormErr)
  }

  if xormErr = xormdb.Sync(new(model.Version)); xormErr!= nil {
    log.Fatalf("Fail to sync test database: %v\n", xormErr)
  }
}

func useDb() martini.Handler {
    return func(c martini.Context){
    c.Map(xormdb)
    }
}


var _ = Describe("Testing Update AppVersion Create", func() {
  It("POST '/v1/updates/:app' will returns a http.StatusCreated status code", func() {

    m := martini.Classic()
    m.Use(useDb())
    m.Use(VerifyJSONBody())
    m.Use(VerifyHTTPHeader())
    m.Use(render.Renderer())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Post("/:app",handler.CreateVersion)
    })

    response = httptest.NewRecorder()
    test := model.Version{Version: 2, Name: "0.3.2", Updated: "2013-11-10 00:00:00", Changed: "1. New design application icon. \n 2. Fix some bugs.", Url: "http//file.arkors.com/releases/demo-lasttest.apk", Client: "Android", Compatible: "0.3.2 0.3.1 0.2.x 0.2.x"}
    body, _ := json.Marshal(test)
    request, _ := http.NewRequest("POST", "/v1/updates/233", bytes.NewReader(body))
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "127.0.0.1,BOARD")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    var result model.Version

    err := json.Unmarshal(response.Body.Bytes(), &result)

    Ω(err).Should(BeNil())
    Ω(result.App).Should(BeEquivalentTo(233))

    Expect(response.Code).To(Equal(http.StatusCreated))
  })

  It("POST '/v1/updates/:app' repetition application ID will  returns a http.StatusBadRequest status code", func() {

    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Use(VerifyJSONBody())
    m.Use(VerifyHTTPHeader())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Post("/:app",handler.CreateVersion)
    })

    response = httptest.NewRecorder()
    test := model.Version{Version: 2, Name: "0.3.2",Updated:"2013-11-10 00:00:00",Changed: "1. New design application icon. \n 2. Fix some bugs.", Url: "http//file.arkors.com/releases/demo-lasttest.apk", Client: "Android", Compatible: "0.3.2 0.3.1 0.2.x 0.2.x"}
    body, _ := json.Marshal(test)
    request, _ := http.NewRequest("POST", "/v1/updates/233", bytes.NewReader(body))
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "127.0.0.1,BOARD")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }
    var result Result
    err2 := json.Unmarshal(response.Body.Bytes(), &result)

    Ω(err2).Should(BeNil())

    Expect(response.Code).To(Equal(http.StatusBadRequest))

  })

  It("POST '/v1/updates/:app' invalid application ID will  returns a http.StatusBadRequest status code", func() {

    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(VerifyJSONBody())
    m.Use(VerifyHTTPHeader())
    m.Use(useDb())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Post("/:app", handler.CreateVersion)
    })

    response = httptest.NewRecorder()
    test := model.Version{Version: 3, Name: "0.3.2", Updated:"2013-11-10 00:00:00",Changed: "1. New design application icon. \n 2. Fix some bugs.", Url: "http//file.arkors.com/releases/demo-lasttest.apk", Client: "Android", Compatible: "0.3.2 0.3.1 0.2.x 0.2.x"}
    body, _ := json.Marshal(test)
    request, _ := http.NewRequest("POST", "/v1/updates/233updates", bytes.NewReader(body))
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "127.0.0.1,BOARD")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }
    var result Result
    err2 := json.Unmarshal(response.Body.Bytes(), &result)
    Ω(err2).Should(BeNil())

    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })

  It("POST '/v1/updates/:app' with a invalid json body,  will returns a http.StatusBadRequest status code", func() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Use(VerifyJSONBody())
    m.Use(VerifyHTTPHeader())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Post("/:app", handler.CreateVersion)
    })

    response = httptest.NewRecorder()
    test := model.Version{Version: 3, Name: "0.3.2", Updated:"2013-11-10 00:00:00",Changed: "1. New design application icon. \n 2. Fix some bugs.", Url: "http//file.arkors.com/releases/demo-lasttest.apk", Client: "Android", Compatible: "0.3.2 0.3.1 0.2.x 0.2.x"}
    body, _ := json.Marshal(test)
    request, _ := http.NewRequest("POST", "/v1/updates/233updates", bytes.NewReader(body))
    //request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "127.0.0.1,BOARD")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }

    var result Result
    err := json.Unmarshal(response.Body.Bytes(), &result)

    Ω(err).Should(BeNil())

    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })

  It("POST '/v1/updates/:app' with a invalid json field,  will returns a http.StatusBadRequest status code", func() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Use(VerifyJSONBody())
    m.Use(VerifyHTTPHeader())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Post("/:app",handler.CreateVersion)
    })

    response = httptest.NewRecorder()
    request, _ := http.NewRequest("POST", "/v1/updates/233", bytes.NewReader([]byte("{\"testVersion:\"\"3\"}")))
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "127.0.0.1,BOARD")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }
    var result Result
    err3 := json.Unmarshal(response.Body.Bytes(), &result)

    Ω(err3).Should(BeNil())

    Expect(response.Code).To(Equal(http.StatusBadRequest))

  })
})

var _ = Describe("Testing Update App Get Version Information", func() {
  It("Get '/v1/updates/:app/:version' will returns a http.StatusOK status code", func() {

    m := martini.Classic()
    //m.Use(VerifyJSONBody())
    m.Use(VerifyHTTPHeader())
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Get("/:app/:version",handler.GetVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("GET", "/v1/updates/233/1", nil)
    request.Header.Set("X-Arkors-Application-Id", "233")
    request.Header.Set("X-Arkors-Application-Token", "cb21df532c6647383af7efa0fd8405f2,1389085779854")
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "3ad3ce877d6c42b131580748603f8d6a,ANDROID")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    var result model.Version
    err := json.Unmarshal(response.Body.Bytes(), &result)
    fmt.Println(result)
    fmt.Println(err)
    Ω(err).Should(BeNil())
    Ω(result.App).Should(BeEquivalentTo(233))
    Expect(response.Code).To(Equal(http.StatusOK))
  })
  It("GET '/v1/updates/:app/:version' with invalid appId or versionId  will returns a http.StatusBadRequest status code", func() {

    m := martini.Classic()
    m.Use(useDb())
    //m.Use(VerifyJSONBody())
    m.Use(VerifyHTTPHeader())
    m.Use(render.Renderer())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Get("/:app/:version",handler.GetVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("GET", "/v1/updates/233d/3b", nil)
    request.Header.Set("X-Arkors-Application-Id", "233")
    request.Header.Set("X-Arkors-Application-Token", "cb21df532c6647383af7efa0fd8405f2,1389085779854")
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "3ad3ce877d6c42b131580748603f8d6a,ANDROID")
    request.Header.Set("Accept", "application/json")

    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }

    var result Result
    err := json.Unmarshal(response.Body.Bytes(), &result)

    Ω(err).Should(BeNil())
    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })

  It("GET '/v1/updates/:app/:version' with missing fields or etc  will returns a http.StatusBadRequest status code", func() {

    m := martini.Classic()
    //m.Use(VerifyJSONBody())
    m.Use(VerifyHTTPHeader())
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Get("/:app/:version",handler.GetVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("GET", "/v1/updates/233/1", nil)
    request.Header.Set("X-Arkors-Application-Id", "233")
    //request.Header.Set("X-Arkors-Application-Token","cb21df532c6647383af7efa0fd8405f2,1389085779854")
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "3ad3ce877d6c42b131580748603f8d6a,ANDROID")
    request.Header.Set("Accept", "application/json")

    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }

    var result Result
    err := json.Unmarshal(response.Body.Bytes(), &result)

    Ω(err).Should(BeNil())
    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })
  It("GET '/v1/updates/:app/:version' not in right upgrade time  will returns a http.StatusBadRequest status code", func() {

    m := martini.Classic()
//m.Use(VerifyJSONBody())
    m.Use(VerifyHTTPHeader())
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Get("/:app/:version",handler.GetVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("GET", "/v1/updates/234/1", nil)
    request.Header.Set("X-Arkors-Application-Id", "234")
    request.Header.Set("X-Arkors-Application-Token", "cb21df532c6647383af7efa0fd8405f2,1389085779854")
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "3ad3ce877d6c42b131580748603f8d6a,ANDROID")
    request.Header.Set("Accept", "application/json")

    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }

    var result Result
    err := json.Unmarshal(response.Body.Bytes(), &result)

    Ω(err).Should(BeNil())
    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })

  It("GET '/v1/updates/:app/:version' with  newest version information will returns a http.StatusNotFound status code", func() {

    m := martini.Classic()
//m.Use(VerifyJSONBody())
    m.Use(VerifyHTTPHeader())
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Get("/:app/:version",handler.GetVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("GET", "/v1/updates/233/2", nil)
    request.Header.Set("X-Arkors-Application-Id", "233")
    request.Header.Set("X-Arkors-Application-Token", "cb21df532c6647383af7efa0fd8405f2,1389085779854")
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "3ad3ce877d6c42b131580748603f8d6a,ANDROID")
    request.Header.Set("Accept", "application/json")

    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }

    var result Result
    err := json.Unmarshal(response.Body.Bytes(), &result)
    Ω(err).Should(BeNil())

    Expect(response.Code).To(Equal(http.StatusNotFound))
  })
})

var _ = Describe("Testing Updates App Update Application Infomation Put", func() {
  It("Put '/v1/updates/:app/:version' will returns a http.StatusOK status code", func() {
    m := martini.Classic()
    m.Use(VerifyJSONBody())
    m.Use(VerifyHTTPHeader())
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Put("/:app/:version",handler.UpdateVersion)
    })

    response = httptest.NewRecorder()
    test := model.Version{Version:2, Name: "0.3.3", Changed: "1. New design application icon. \n 2. Fix some bugs. 3. Add GCM push supported", Url: "http//file.arkors.com/releases/demo-lasttest.apk", Client: "Android", Compatible: "0.3.2 0.3.1 0.2.x 0.1.x"}
    body, err := json.Marshal(test)
    request, _ := http.NewRequest("PUT", "/v1/updates/233/2", bytes.NewReader(body))
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "127.0.0.1,BOARD")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    var version model.Version

    err = json.Unmarshal(response.Body.Bytes(), &version)
    Ω(err).Should(BeNil())
    Ω(version.App).Should(BeEquivalentTo(233))

    Expect(response.Code).To(Equal(http.StatusOK))
  })
  It("PUT '/v1/updates/:app/:version' with invalid json will returns a http.StatusBadRequest status code", func() {

    m := martini.Classic()
    m.Use(VerifyJSONBody())
    m.Use(VerifyHTTPHeader())
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Put("/:app/:version",handler.UpdateVersion)
    })

    response = httptest.NewRecorder()
    request, _ := http.NewRequest("PUT", "/v1/updates/233/2", bytes.NewReader([]byte("{\"Version\"\"3\"}")))
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "127.0.0.1,BOARD")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }

    var result Result
    err := json.Unmarshal(response.Body.Bytes(), &result)

    Ω(err).Should(BeNil())
    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })
  It("PUT '/v1/updates/:app/:version' with missing fields or etc will returns a http.StatusBadRequest status code", func() {

    m := martini.Classic()
    m.Use(VerifyHTTPHeader())
    m.Use(VerifyJSONBody())
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Put("/:app/:version", handler.UpdateVersion)
    })

    response = httptest.NewRecorder()
    test := model.Version{Version: 3, Name: "0.3.3", Changed: "1. New design application icon. \n 2. Fix some bugs. 3. Add GCM push supported", Url: "http//file.arkors.com/releases/demo-lasttest.apk", Client: "Android", Compatible: "0.3.2 0.3.1 0.2.x 0.1.x"}
    body, err := json.Marshal(test)
    request, _ := http.NewRequest("PUT", "/v1/updates/232/3", bytes.NewReader(body))
    // request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "127.0.0.1,BOARD")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }

    var result Result
    err = json.Unmarshal(response.Body.Bytes(), &result)
    Ω(err).Should(BeNil())

    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })
  It("PUT '/v1/updates/:app/:version' with not found version infomation will returns a http.StatusNotFound status code", func() {

    m := martini.Classic()
    m.Use(VerifyHTTPHeader())
    m.Use(VerifyJSONBody())
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Put("/:app/:version",handler.UpdateVersion)
    })

    response = httptest.NewRecorder()
    test := model.Version{Version: 3, Name: "0.3.3", Changed: "1. New design application icon. \n 2. Fix some bugs. 3. Add GCM push supported", Url: "http//file.arkors.com/releases/demo-lasttest.apk", Client: "Android", Compatible: "0.3.2 0.3.1 0.2.x 0.1.x"}
    body, err := json.Marshal(test)
    request, _ := http.NewRequest("PUT", "/v1/updates/239/5", bytes.NewReader(body))
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "127.0.0.1,BOARD")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }
    var result Result
    err = json.Unmarshal(response.Body.Bytes(), &result)
    Ω(err).Should(BeNil())
    Expect(response.Code).To(Equal(http.StatusNotFound))
  })
})


var _ = Describe("Testing Update App Delete Version", func() {
  It("DELETE '/v1/updates/:app/:version' will returns a http.StatusOK status code", func() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Delete("/:app/:version",handler.DelVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("DELETE", "/v1/updates/233/1", nil)
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "3ad3ce877d6c42b131580748603f8d6a,ANDROID")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    var result model.Version
    err := json.Unmarshal(response.Body.Bytes(), &result)
    Ω(err).Should(BeNil())
    Ω(result.App).Should(BeEquivalentTo(233))

    Expect(response.Code).To(Equal(http.StatusOK))

  })

  It("DELETE '/v1/updates/:app/:version' with invalid id will returns a http.StatusBadRequest status code", func() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Delete("/:app/:version", handler.DelVersion)
    })

    response = httptest.NewRecorder()
    request, _ := http.NewRequest("DELETE", "/v1/updates/233abc/1", nil)
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "3ad3ce877d6c42b131580748603f8d6a,ANDROID")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }

    var result Result
    err := json.Unmarshal(response.Body.Bytes(), &result)

    Ω(err).Should(BeNil())
    Expect(response.Code).To(Equal(http.StatusBadRequest))

  })
  It("DELETE '/v1/updates/:app/:version' with invalid fields or etc will returns a http.StatusBadRequest status code", func() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Delete("/:app/:version", handler.DelVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("DELETE", "/v1/updates/233/1", nil)
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    //request.Header.Set("X-Arkors-Application-Client", "3ad3ce877d6c42b131580748603f6a,ANDROID")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }

    var result Result
    err := json.Unmarshal(response.Body.Bytes(), &result)

    Ω(err).Should(BeNil())

    Expect(response.Code).To(Equal(http.StatusBadRequest))

  })

  It("DELETE '/v1/updates/:app/:version' with no application Id will returns a http.StatusNotFound status code", func() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(useDb())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Delete("/:app/:version", handler.DelVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("DELETE", "/v1/updates/235/6", nil)
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "3ad3ce877d6c42b131580748603f6a,ANDROID")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    type Result struct {
      Error string
    }

    var result Result
    err := json.Unmarshal(response.Body.Bytes(), &result)

    Ω(err).Should(BeNil())

    Expect(response.Code).To(Equal(http.StatusNotFound))

  })
})

