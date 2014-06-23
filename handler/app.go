package handler

import (
  "bytes"
  "github.com/arkors/update/model"
  "github.com/go-martini/martini"
  "github.com/go-xorm/xorm"
  //"github.com/codegangsta/martini-contrib/binding"
  "fmt"
  "github.com/hoisie/redis"
  "github.com/martini-contrib/render"
  "net/http"
  "sort"
  "strconv"
  "strings"
  "time"
)

func CreateVersion(db *xorm.Engine, params martini.Params, version model.Version, r render.Render, res *http.Request) {
  fmt.Println(version.Updated)
  appId, err := strconv.ParseInt(params["app"], 0, 64)
  versionId, _ := strconv.Itoa(version.Version)
  if err != nil {
    r.JSON(400, map[string]interface{}{"error": "The application's id must be numrical"})
    return
  }
  _, log := res.Header["X-Arkors-Application-Log"]
  _, client := res.Header["X-Arkors-Application-Client"]
  if log != true || client != true {
    r.JSON(400, map[string]interface{}{"error": "Invalid request header,it should be include 'X-Arkors-Application-log' and 'X-Arkors-Application-Client'."})
    return
  }
  if version.Version == 0 || version.Name == "" || version.Changed == "" || version.Url == "" || version.Client == "" || version.Compatible == "" || version.Updated == "" {
    r.JSON(400, map[string]interface{}{"error": "Invalid json body "})
    return
  }
  sql := "select * from version where app=" + params["app"] + " and version=" +
    fmt.Println(sql)
  results, err := db.Query(sql)
  //fmt.Println(results)
  if results != nil {
    //fmt.Println("testRepeatid")
    r.JSON(400, map[string]interface{}{"error": "The application's id already exist"})
    return
  } else {
    version.App = appId
    _, err2 := db.Insert(version)
    if err2 != nil {
      r.JSON(400, map[string]interface{}{"error": "Database error"})
      return
    } else {
      //Indert into redis
      //sql := "select * from version where app=" + params["app"]
      //result := new(model.Version)
      //has, err := db.Sql(sql).Get(result)
      //if err == nil && has {
      var client redis.Client
      var buffer bytes.Buffer
      //idTrans := strconv.FormatInt(result.Id, 10)
      //appTrans := strconv.FormatInt(result.App, 10)
      //buffer.WriteString("Id=" + idTrans + "|")
      //buffer.WriteString("App=" + appTrans + "|")
      //buffer.WriteString("Version=" + result.Version + "|")
      //buffer.WriteString("Name=" + result.Name + "|")
      //buffer.WriteString("Updated=" + String(result.Updated) + "|")
      //buffer.WriteString("Changed=" + result.Changed + "|")
      //buffer.WriteString("Client=" + result.Client + "|")
      //buffer.WriteString("Url=" + result.Url + "|")
      //buffer.WriteString("Compatible=" + result.Compatible)
      //client.Set(appTrans+"@"+result.Version, []byte(buffer.String()))
      buffer.WriteString("App=" + params["app"] + "|")
      buffer.WriteString("Version=" + strconv.Itoa(version.Version) + "|")
      buffer.WriteString("Name=" + version.Name + "|")
      buffer.WriteString("Updated=" + version.Updated + "|")
      buffer.WriteString("Changed=" + version.Changed + "|")
      buffer.WriteString("Client=" + version.Client + "|")
      buffer.WriteString("Url=" + version.Url + "|")
      buffer.WriteString("Compatible=" + version.Compatible)
      client.Set(params["app"]+"@"+strconv.Itoa(version.Version), []byte(buffer.String()))
      r.JSON(201, version)
      fmt.Println(version)
      return
      // }
    }
  }
}
func GetVersion(db *xorm.Engine, params martini.Params, r render.Render, res *http.Request) {
  appId, errAppId := strconv.ParseInt(params["app"], 0, 64)
  versionNumber, versionErr := strconv.Atoi(params["version"])
  if errAppId != nil && versionErr != nil {
    r.JSON(400, map[string]interface{}{"error": "The application's id must be numrical"})
    return
  }
  _, log := res.Header["X-Arkors-Application-Log"]
  _, id := res.Header["X-Arkors-Application-Id"]
  _, Token := res.Header["X-Arkors-Application-Token"]
  _, clientHeader := res.Header["X-Arkors-Application-Client"]
  if log != true || clientHeader != true || id != true || Token != true {
    r.JSON(400, map[string]interface{}{"error": "Invalid request header,it should be include 'X-Arkors-Application-log' and 'X-Arkors-Application-Client'."})
    return
  }
  var client redis.Client
  //get all keys
  keyAll, _ := client.Keys(params["app"] + "@*")
  var versionArray = make([]int, len(keyAll))
  for i, _ := range versionArray {
    versionArray[i], _ = strconv.Atoi(strings.Split(keyAll[i], "@")[1])
  }
  sort.Ints(versionArray)

  if versionArray[len(keyAll)-1] == versionNumber {
    r.JSON(404, map[string]interface{}{"error": "The application version is newest!"})
    return
  }

  versionModelNew := new(model.Version)
  result, err := client.Get(params["app"] + "@" + strconv.Itoa(versionArray[len(keyAll)-1]))
  if err == nil && result != nil {
    fmt.Println(string(result))
    versionStringArray := strings.Split(string(result), "|")
    for i, _ := range versionStringArray {
      colum := strings.Split(versionStringArray[i], "=")[0]
      value := strings.Split(versionStringArray[i], "=")[1]
      fmt.Println(colum + "=" + value)
      if colum == "Id" {
        idTrans, _ := strconv.ParseInt(value, 0, 64)
        versionModelNew.Id = idTrans
      } else if colum == "App" {
        versionModelNew.App = appId
      } else if colum == "Version" {
        versionModelNew.Version, _ = strconv.Atoi(value)
      } else if colum == "Name" {
        versionModelNew.Name = value
      } else if colum == "Updated" {
        versionModelNew.Updated = value
      } else if colum == "Changed" {
        versionModelNew.Changed = value
      } else if colum == "Client" {
        versionModelNew.Client = value
      } else if colum == "Url" {
        versionModelNew.Url = value
      } else if colum == "Compatible" {
        versionModelNew.Compatible = value
      }
    }
  }
  versionModelOld := new(model.Version)
  result, err = client.Get(params["app"] + "@" + params["version"])
  if err == nil && result != nil {
    fmt.Println(string(result))
    versionStringArray := strings.Split(string(result), "|")
    for i, _ := range versionStringArray {
      colum := strings.Split(versionStringArray[i], "=")[0]
      value := strings.Split(versionStringArray[i], "=")[1]
      fmt.Println(colum + "=" + value)
      if colum == "Id" {
        idTrans, _ := strconv.ParseInt(value, 0, 64)
        versionModelOld.Id = idTrans
      } else if colum == "App" {
        versionModelOld.App = appId
      } else if colum == "Version" {
        versionModelOld.Version, _ = strconv.Atoi(value)
      } else if colum == "Name" {
        versionModelOld.Name = value
      } else if colum == "Updated" {
        versionModelOld.Updated = value
      } else if colum == "Changed" {
        versionModelOld.Changed = value
      } else if colum == "Client" {
        versionModelOld.Client = value
      } else if colum == "Url" {
        versionModelOld.Url = value
      } else if colum == "Compatible" {
        versionModelOld.Compatible = value
      }
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
    timeArr := strings.Split(versionModelNew.Updated, " ")
    year, _ := strconv.Atoi(strings.Split(timeArr[0], "-")[0])
    month, _ := strconv.Atoi(strings.Split(timeArr[0], "-")[1])
    var goMonth time.Month
    switch month {
    case 1:
      goMonth = time.January
    case 2:
      goMonth = time.February
    case 3:
      goMonth = time.March
    case 4:
      goMonth = time.April
    case 5:
      goMonth = time.May
    case 6:
      goMonth = time.June
    case 7:
      goMonth = time.July
    case 8:
      goMonth = time.August
    case 9:
      goMonth = time.September
    case 10:
      goMonth = time.October
    case 11:
      goMonth = time.November
    case 12:
      goMonth = time.December
    }
    day, _ := strconv.Atoi(strings.Split(timeArr[0], "-")[2])
    hour, _ := strconv.Atoi(strings.Split(timeArr[1], ":")[0])
    min, _ := strconv.Atoi(strings.Split(timeArr[1], ":")[1])
    ss, _ := strconv.Atoi(strings.Split(timeArr[1], ":")[2])
    t2 := time.Date(year, goMonth, day, hour, min, ss, 0, time.UTC)
    fmt.Println(t2.String())
    t1 := time.Now()
    timeResult := t1.After(t2)
    if timeResult {
      r.JSON(200, versionModelNew)
      return
    } else {
      r.JSON(400, map[string]interface{}{"error": "The App version can't in right time to upgrade"})
      return
    }
  } else {
    r.JSON(400, map[string]interface{}{"error": "The App version can't upgrade"})
    return
  }
}

func UpdateApp(db *xorm.Engine, params martini.Params, version model.Version, r render.Render, res *http.Request) {
  appId, errAppId := strconv.ParseInt(params["app"], 0, 64)
  if errAppId != nil {
    r.JSON(400, map[string]interface{}{"error": "The application's id must be numrical"})
    return
  }
  _, log := res.Header["X-Arkors-Application-Log"]
  _, clientHeader := res.Header["X-Arkors-Application-Client"]
  if log != true || clientHeader != true {
    fmt.Println("Invalid request header,it should be include 'X-Arkors-Application-log' and 'X-Arkors-Application-Client'.")
    r.JSON(400, map[string]interface{}{"error": "Invalid request header,it should be include 'X-Arkors-Application-log' and 'X-Arkors-Application-Client'."})
    return
  }
  if version.Version == 0 || version.Name == "" || version.Changed == "" || version.Url == "" || version.Client == "" || version.Compatible == "" {
    r.JSON(400, map[string]interface{}{"error": "Invalid json body "})
    return
  }
  version.App = appId
  version.Version, _ = strconv.Atoi(params["version"])
  fmt.Println("changed=======" + version.Changed)
  has, err := db.In("App", appId).Update(version)
  if err != nil {
    r.JSON(400, map[string]interface{}{"error": "Database Error"})
    return
  }
  if has == 0 {
    r.JSON(404, map[string]interface{}{"error": "Not found any version records"})
    return
  } else {
    var client redis.Client
    var buffer bytes.Buffer
    buffer.WriteString("App=" + params["app"] + "|")
    buffer.WriteString("Version=" + params["version"] + "|")
    buffer.WriteString("Name=" + version.Name + "|")
    //buffer.WriteString("Updated=" + String(result.Updated) + "|")
    buffer.WriteString("Changed=" + version.Changed + "|")
    buffer.WriteString("Client=" + version.Client + "|")
    buffer.WriteString("Url=" + version.Url + "|")
    buffer.WriteString("Compatible=" + version.Compatible)
    client.Set(params["app"]+"@"+params["version"], []byte(buffer.String()))
    val, _ := client.Get(params["app"] + "@" + params["version"])
    key := params["app"] + "@" + params["version"]
    fmt.Println("key=="+key, "val==="+string(val))
    r.JSON(200, version)
    return
  }
}

func DelVersion(db *xorm.Engine, params martini.Params, version model.Version, r render.Render, res *http.Request) {
  _, err := strconv.ParseInt(params["app"], 0, 64)
  if err != nil {
    r.JSON(400, map[string]interface{}{"error": "The application's id must be numrical"})
    return
  }
  _, log := res.Header["X-Arkors-Application-Log"]
  _, clientHeader := res.Header["X-Arkors-Application-Client"]
  if log != true || clientHeader != true {
    r.JSON(400, map[string]interface{}{"error": "Invalid request header,it should be include 'X-Arkors-Application-log' and 'X-Arkors-Application-Client'."})
    return
  }
  sql := "select * from version where app=" + params["app"] + " and version='" + params["version"] + "'"
  result := new(model.Version)
  has, err := db.Sql(sql).Get(result)
  if err != nil {
    r.JSON(400, map[string]interface{}{"error": "Datebase Error"})
    return
  }
  if has {
    //db.In("Id", id).Delete(result)
    _, err = db.Exec("delete from version where app=" + params["app"] + " and version='" + params["version"] + "'")
    if err == nil {
      var client redis.Client
      fmt.Println(params["app"] + "@" + params["version"])
      client.Del(params["app"] + "@" + params["version"])
      r.JSON(200, result)
      return
    }
  } else {
    r.JSON(404, map[string]interface{}{"error": "Application's ID is not exist!"})
    return
  }
}
