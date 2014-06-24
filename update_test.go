package main

import (
  "bytes"
  "encoding/json"
  //"fmt"
  "github.com/arkors/update/handler"
  "github.com/arkors/update/model"
  "github.com/go-martini/martini"
  "github.com/martini-contrib/binding"
  "github.com/martini-contrib/render"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  "net/http"
  "net/http/httptest"
)

var _ = Describe("Testing Update AppVersion Create", func() {
  It("POST '/v1/updates/:app' will returns a http.StatusCreated status code", func() {

    m := martini.Classic()
    m.Use(Db())
    m.Use(render.Renderer())
    // m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Post("/:app", binding.Json(model.Version{}), handler.CreateVersion)
    })

    response = httptest.NewRecorder()
    test := model.Version{Version: 4, Name: "0.3.3", Updated: "http.StatusCreated4-11-10 00:00:00", Changed: "1. New design application icon. \n 2. Fix some bugs.", Url: "http//file.arkors.com/releases/demo-lasttest.apk", Client: "Android", Compatible: "0.3.2 0.3.1 0.2.x 0.2.x"}
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
    m.Use(Db())
    // m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Post("/:app", binding.Json(model.Version{}), handler.CreateVersion)
    })

    response = httptest.NewRecorder()
    test := model.Version{Version: "3", Name: "0.3.3", Changed: "1. New design application icon. \n 2. Fix some bugs.", Url: "http//file.arkors.com/releases/demo-lasttest.apk", Client: "Android", Compatible: "0.3.2 0.3.1 0.2.x 0.2.x"}
    body, _ := json.Marshal(test)
    request, _ := http.NewRequest("POST", "/v1/updates/232", bytes.NewReader(body))
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
    // Ω(len(result.Error)).Should(BeNumerically(">"), 0)

    Expect(response.Code).To(Equal(http.StatusBadRequest))

  })

  It("POST '/v1/updates/:app' invalid application ID will  returns a http.StatusBadRequest status code", func() {

    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    // m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Post("/:app", binding.Json(model.Version{}), handler.CreateVersion)
    })

    response = httptest.NewRecorder()
    test := model.Version{Version: "3", Name: "0.3.3", Changed: "1. New design application icon. \n 2. Fix some bugs.", Url: "http//file.arkors.com/releases/demo-lasttest.apk", Client: "Android", Compatible: "0.3.2 0.3.1 0.2.x 0.2.x"}
    body, _ := json.Marshal(test)
    request, _ := http.NewRequest("POST", "/v1/updates/232updates", bytes.NewReader(body))
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
    //Ω(len(result.Error)).Should(BeNumerically(">"), 0)

    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })

  It("POST '/v1/updates/:app' with a invalid json body,  will returns a http.StatusBadRequest status code", func() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    // m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Post("/:app", binding.Json(model.Version{}), handler.CreateVersion)
    })

    response = httptest.NewRecorder()
    request, _ := http.NewRequest("POST", "/v1/updates/232", bytes.NewReader([]byte("{\"Version\"\"3\"}")))
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
    //Ω(len(result.Error)).Should(BeNumerically(">", 0))

    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })

  It("POST '/v1/updates/:app' with a invalid json field,  will returns a http.StatusBadRequest status code", func() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    // m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Post("/:app", binding.Json(model.Version{}), handler.CreateVersion)
    })

    response = httptest.NewRecorder()
    request, _ := http.NewRequest("POST", "/v1/updates/232", bytes.NewReader([]byte("{\"testVersion:\"\"3\"}")))
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
    //Ω(result.Error).Should(BeNumerically(">"), 0)

    Expect(response.Code).To(Equal(http.StatusBadRequest))

  })
})

var _ = Describe("Testing Update App Get Version Information", func() {
  It("Get '/v1/updates/:app/:version' will returns a http.StatusOK status code", func() {

    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    //m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Get("/:app/:version", binding.Json(model.Version{}), handler.GetVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("GET", "/v1/updates/233/3", nil)
    request.Header.Set("X-Arkors-Application-Id", "232")
    request.Header.Set("X-Arkors-Application-Token", "cb21df532c6647383af7efa0fd8405f2,1389085779854")
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "3ad3ce877d6c42b131580748603f8d6a,ANDROID")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    var result model.Version
    err := json.Unmarshal(response.Body.Bytes(), &result)
    Ω(err).Should(BeNil())
    Ω(result.App).Should(BeEquivalentTo(233))
    //Ω(len(result.Key)).Should(BeNumerically("==", 32))
    fmt.Println("result.Changed=" + result.Changed)
    Expect(response.Code).To(Equal(http.StatusOK))
  })
  It("GET '/v1/updates/:app/:version' with invalid appId or versionId  will returns a http.StatusBadRequest status code", func() {

    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    //m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Get("/:app/:version", binding.Json(model.Version{}), handler.GetVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("GET", "/v1/updates/232a/3b", nil)
    request.Header.Set("X-Arkors-Application-Id", "232")
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
    //Ω(len(result.Error)).Should(BeNumerically(">", 0))
    fmt.Println("result.Error====" + result.Error)
    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })

  It("GET '/v1/updates/:app/:version' with missing fields or etc  will returns a http.StatusBadRequest status code", func() {

    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    //m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Get("/:app/:version", binding.Json(model.Version{}), handler.GetVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("GET", "/v1/updates/232/3", nil)
    request.Header.Set("X-Arkors-Application-Id", "232")
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
    //Ω(len(result.Error)).Should(BeNumerically(">", 0))
    fmt.Println("result.Error====" + result.Error)
    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })
  It("GET '/v1/updates/:app/:version' with invalid fields or etc  will returns a http.StatusBadRequest status code", func() {

    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    //m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Get("/:app/:version", binding.Json(model.Version{}), handler.GetVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("GET", "/v1/updates/232/3", nil)
    //request.Header.Set("X-Arkors-Application-Id", "232a")
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
    //Ω(len(result.Error)).Should(BeNumerically(">", 0))
    fmt.Println("result.Error====" + result.Error)
    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })

  It("GET '/v1/updates/:app/:version' with  newest version information will returns a http.StatusNotFound status code", func() {

    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    //m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Get("/:app/:version", binding.Json(model.Version{}), handler.GetVersion)
    })

    p
    p
    response = httptest.NewRecorder()

    request, _ := http.NewRequest("GET", "/v1/updates/233/5", nil)
    request.Header.Set("X-Arkors-Application-Id", "235")
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
    fmt.Println("result.Error====" + result.Error)
    Ω(err).Should(BeNil())
    //Ω(len(result.Error)).Should(BeNumerically(">", 0))

    Expect(response.Code).To(Equal(http.StatusNotFound))
  })
})

var _ = Describe("Testing Updates App Update Application Infomation Put", func() {
  It("Put '/v1/updates/:app/:version' will returns a http.StatusOK status code", func() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    //m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Put("/:app/:version", binding.Json(model.Version{}), handler.UpdateApp)
    })

    response = httptest.NewRecorder()
    test := model.Version{Version: "3", Name: "0.3.3", Changed: "1. New design application icon. \n 2. Fix some bugs. 3. Add GCM push supported", Url: "http//file.arkors.com/releases/demo-lasttest.apk", Client: "Android", Compatible: "0.3.2 0.3.1 0.2.x 0.1.x"}
    body, err := json.Marshal(test)
    request, _ := http.NewRequest("PUT", "/v1/updates/233/3", bytes.NewReader(body))
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
    m.Use(render.Renderer())
    m.Use(Db())
    //m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Put("/:app/:version", binding.Json(model.Version{}), handler.UpdateApp)
    })

    response = httptest.NewRecorder()
    request, _ := http.NewRequest("PUT", "/v1/updates/232/5", bytes.NewReader([]byte("{\"Version\"\"3\"}")))
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
    //Ω(len(result.Error)).Should(BeNumerically(">", 0))
    fmt.Println("errorRRRR=====" + result.Error)
    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })
  It("PUT '/v1/updates/:app/:version' with missing fields or etc will returns a http.StatusBadRequest status code", func() {

    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    //m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Put("/:app/:version", binding.Json(model.Version{}), handler.UpdateApp)
    })

    response = httptest.NewRecorder()
    test := model.Version{Version: "3", Name: "0.3.3", Changed: "1. New design application icon. \n 2. Fix some bugs. 3. Add GCM push supported", Url: "http//file.arkors.com/releases/demo-lasttest.apk", Client: "Android", Compatible: "0.3.2 0.3.1 0.2.x 0.1.x"}
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
    //Ω(len(result.Error)).Should(BeNumerically(">", 0))

    Expect(response.Code).To(Equal(http.StatusBadRequest))
  })
  It("PUT '/v1/updates/:app/:version' with not found version infomation will returns a http.StatusNotFound status code", func() {

    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    //m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Put("/:app/:version", binding.Json(model.Version{}), handler.UpdateApp)
    })

    response = httptest.NewRecorder()
    test := model.Version{Version: "3", Name: "0.3.3", Changed: "1. New design application icon. \n 2. Fix some bugs. 3. Add GCM push supported", Url: "http//file.arkors.com/releases/demo-lasttest.apk", Client: "Android", Compatible: "0.3.2 0.3.1 0.2.x 0.1.x"}
    body, err := json.Marshal(test)
    request, _ := http.NewRequest("PUT", "/v1/updates/235/5", bytes.NewReader(body))
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
    //Ω(len(result.Error)).Should(BeNumerically(">", 0))
    fmt.Println(result.Error)
    Expect(response.Code).To(Equal(http.StatusNotFound))
  })
})

var _ = Describe("Testing Update App Delete Version", func() {
  It("DELETE '/v1/updates/:app/:version' will returns a http.StatusOK status code", func() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    //m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Delete("/:app/:version", binding.Json(model.Version{}), handler.DelVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("DELETE", "/v1/updates/233/3", nil)
    request.Header.Set("X-Arkors-Application-Log", "5024442115e7bd738354c1fac662aed5")
    request.Header.Set("X-Arkors-Application-Client", "3ad3ce877d6c42b131580748603f8d6a,ANDROID")
    request.Header.Set("Accept", "application/json")
    m.ServeHTTP(response, request)

    var result model.Version
    err := json.Unmarshal(response.Body.Bytes(), &result)
    Ω(err).Should(BeNil())
    Ω(result.App).Should(BeEquivalentTo(233))
    //Ω(len(result.Key)).Should(BeNumerically("==", 32))

    Expect(response.Code).To(Equal(http.StatusOK))

  })

  It("DELETE '/v1/updates/:app/:version' with invalid id will returns a http.StatusBadRequest status code", func() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Delete("/:app/:version", binding.Json(model.Version{}), handler.DelVersion)
    })

    response = httptest.NewRecorder()
    request, _ := http.NewRequest("DELETE", "/v1/updates/232abc/5", nil)
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
    //Ω(len(result.Error)).Should(BeNumerically(">", 0))
    Expect(response.Code).To(Equal(http.StatusBadRequest))

  })
  It("DELETE '/v1/updates/:app/:version' with invalid fields or etc will returns a http.StatusBadRequest status code", func() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    //m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Delete("/:app/:version", binding.Json(model.Version{}), handler.DelVersion)
    })

    response = httptest.NewRecorder()

    request, _ := http.NewRequest("DELETE", "/v1/updates/232/5", nil)
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
    //Ω(len(result.Error)).Should(BeNumerically(">", 0))

    Expect(response.Code).To(Equal(http.StatusBadRequest))

  })

  It("DELETE '/v1/updates/:app/:version' with no application Id will returns a http.StatusNotFound status code", func() {
    m := martini.Classic()
    m.Use(render.Renderer())
    m.Use(Db())
    //m.Use(Pool())
    m.Group("/v1/updates", func(r martini.Router) {
      r.Delete("/:app/:version", binding.Json(model.Version{}), handler.DelVersion)
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
    //Ω(len(result.Error)).Should(BeNumerically(">", 0))

    Expect(response.Code).To(Equal(http.StatusNotFound))

  })
})
