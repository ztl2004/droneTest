package handler

import (
  "github.com/arkors/update/model"
  "github.com/go-martini/martini"
  "github.com/go-xorm/xorm"
  "github.com/hoisie/redis"
  "github.com/martini-contrib/render"
  "net/http"
  "sort"
  "strconv"
  "strings"
  "time"
)

func CreateVersion(db *xorm.Engine, params martini.Params, version model.Version, r render.Render, res *http.Request) {
  appId, err := strconv.ParseInt(params["app"], 0, 64)
  versionId, _ := strconv.Itoa(version.Version)
  if err != nil {
    r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
    return
  }
  if version.Version == 0 || version.Name == "" || version.Changed == "" || version.Url == "" || version.Client == "" || version.Compatible == "" || version.Updated == "" {
    r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid json body "})
    return
  }
  versionBeforeInsert := new(model.Version)
  has, errDb := db.Where("app=? and version=?", params["app"], version.Version).Get(versionBeforeInsert)
  if has && errDb == nil {
    r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id already exist"})
    return
  } else {
    version.App = appId
    _, err2 := db.Insert(version)
    if err2 != nil {
      r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Database error"})
      return
    } else {
      //写入redis内存库
      var client redis.Client
      versionJson, err := json.Marshal(version)
      if err != nil {
        r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "version struct to json error"})
        return
      }
      client.Set(params["app"]+"@"+strconv.Itoa(version.Version), versionJson)
      r.JSON(http.StatusCreated, version)
      return
    }
  }
}
func GetVersion(db *xorm.Engine, params martini.Params, r render.Render, res *http.Request) {
  appId, errAppId := strconv.ParseInt(params["app"], 0, 64)
  versionNumber, versionErr := strconv.Atoi(params["version"])
  if errAppId != nil && versionErr != nil {
    r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
    return
  }
  _, id := res.Header["X-Arkors-Application-Id"]
  _, Token := res.Header["X-Arkors-Application-Token"]
  if id != true || Token != true {
    r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid request header,it should be include 'id' and 'Token'"})
    return
  }
  var client redis.Client
  //获取内存库中所有的key值
  keyAll, redisErr := client.Keys(params["app"] + "@*")
  if redisErr != nil {
    r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "redis error"})
    return
  }
  //内存库key值不存在，查询mysql确认是否存在，如果存在，把数据重新插入到redis中
  if keyAll == nil {
    versionInMysql := new(model.Version)
    has, errDb := db.Where("app=? and version=?", params["app"], versionNumber).Get(versionInMysql)
    if errDb == nil && has {
      versionJson, err := json.Marshal(versionInMysql)
      if err != nil {
        r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "version struct to json error"})
        return
      }
      client.Set(params["app"]+"@"+strconv.Itoa(versionNumber), versionJson)
    }
    r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "the verison is newst!"})
    return
  }
  var versionArray = make([]int, len(keyAll))
  for i, _ := range versionArray {
    versionArray[i], _ = strconv.Atoi(strings.Split(keyAll[i], "@")[1])
  }
  sort.Ints(versionArray)

  if versionArray[len(keyAll)-1] == versionNumber {
    r.JSON(http.StatusNotFound, map[string]interface{}{"error": "The application version is newest!"})
    return
  }

  var versionModelNew model.Version
  result, err := client.Get(params["app"] + "@" + strconv.Itoa(versionArray[len(keyAll)-1]))
  if err == nil && result != nil {
    json2versionErr := json.Unmarshal(result, &versionModelNew)
    if json2versionErr != nil {
      r.JSON(http.StatusNotFound, map[string]interface{}{"error": "json trans struct error"})
    }
  }
  var versionModelOld model.Version
  result, err = client.Get(params["app"] + "@" + params["version"])
  if err == nil && result != nil {
    json2versionErr := json.Unmarshal(result, &versionModelOld)
    if json2versionErr != nil {
      r.JSON(http.StatusNotFound, map[string]interface{}{"error": "json trans struct error"})
    }
  }
  upgrade := false
  appNameArr := strings.Split(versionModelNew.Compatible, " ")
  for _, nameValue := range appNameArr {
    if nameValue == versionModelOld.Name {
      upgrade = true
    }
  }
  if upgrade {
    t2 := time.Parse("2006-01-02 15:04:05", versionModelNew.Updated)
    t1 := time.Now()
    timeResult := t1.After(t2)
    if timeResult {
      r.JSON(http.StatusOK, versionModelNew)
      return
    } else {
      r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The App version can't in right time to upgrade"})
      return
    }
  } else {
    r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The App version can't upgrade"})
    return
  }
}

func UpdateApp(db *xorm.Engine, params martini.Params, version model.Version, r render.Render, res *http.Request) {
  appId, errAppId := strconv.ParseInt(params["app"], 0, 64)
  if errAppId != nil {
    r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
    return
  }
  if version.Version == 0 || version.Name == "" || version.Changed == "" || version.Url == "" || version.Client == "" || version.Compatible == "" {
    r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid json body "})
    return
  }
  version.App = appId
  version.Version, _ = strconv.Atoi(params["version"])
  has, err := db.In("App", appId).Update(version)
  if err != nil {
    r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Database Error"})
    return
  }
  if has == 0 {
    r.JSON(http.StatusNotFound, map[string]interface{}{"error": "Not found any version records"})
    return
  } else {
    var client redis.Client
    versionJson, err := json.Marshal(version)
    if err != nil {
      r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "version struct to json error"})
      return
    }
    client.Set(params["app"]+"@"+strconv.Itoa(version.Version), versionJson)
    r.JSON(http.StatusOK, version)
    return
  }
}

func DelVersion(db *xorm.Engine, params martini.Params, version model.Version, r render.Render, res *http.Request) {
  _, err := strconv.ParseInt(params["app"], 0, 64)
  if err != nil {
    r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
    return
  }
  result := new(model.Version)
  has, err := db.Where("app=? and version=?", params["app"], params["version"]).Get(result)
  if err != nil {
    r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Datebase Error"})
    return
  }
  if has {
    deleteVersion := new(model.Version)
    affect, err = db.Where("app=? and version=?", params["app"], params["version"]).Delete(deleteVersion)
    if affect == 1 && err == nil {
      var client redis.Client
      client.Del(params["app"] + "@" + params["version"])
      r.JSON(http.StatusOK, result)
      return
    }
  } else {
    r.JSON(http.StatusNotFound, map[string]interface{}{"error": "Application's ID is not exist!"})
    return
  }
}
