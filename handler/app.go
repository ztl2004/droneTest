package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/arkors/update/model"
	"github.com/go-martini/martini"
	"github.com/go-xorm/xorm"
	"github.com/hoisie/redis"
	"github.com/martini-contrib/render"
)

func CreateVersion(db *xorm.Engine, params martini.Params, version model.Version, r render.Render, client redis.Client, res *http.Request, logChan chan string) {
	appId, err := strconv.ParseInt(params["app"], 0, 64)
	if err != nil {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
		return
	}
	if version.VersionId == 0 || version.Name == "" || version.Changed == "" || version.Url == "" || version.Client == "" || version.Compatible == "" || version.Updated == "" {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid json body "})
		return
	}
	versionBeforeInsert := new(model.Version)
	has, errDb := db.Where("app=?", appId).And("version_id=?", version.VersionId).Get(versionBeforeInsert)
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
			versionJson, err := json.Marshal(version)
			if err != nil {
				r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "version struct to json error"})
				return
			}
			client.Set(params["app"]+"@"+strconv.Itoa(version.VersionId), versionJson)
			r.JSON(http.StatusCreated, version)
			return
		}
	}
}
func GetVersion(db *xorm.Engine, params martini.Params, r render.Render, client redis.Client, res *http.Request) {
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
	//获取内存库中所有的key值
	keyAll, redisErr := client.Keys(params["app"] + "@*")
	if redisErr != nil {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "redis error"})
		return
	}
	//内存库key值不存在，查询mysql确认是否存在，如果存在，把数据重新插入到redis中
	if keyAll == nil {
		versionInMysql := new(model.Version)
		has, errDb := db.Where("app=?", appId).And("version_id=?", versionNumber).Get(versionInMysql)
		if errDb == nil && has {
			versionJson, err := json.Marshal(versionInMysql)
			if err != nil {
				r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "version struct to json error"})
				return
			}
			client.Set(params["app"]+"@"+strconv.Itoa(versionNumber), versionJson)
		}
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "the version is newst!"})
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
	var upgradeVersion model.Version
	result, err := client.Get(params["app"] + "@" + strconv.Itoa(versionArray[len(keyAll)-1]))
	if err == nil && result != nil {
		json2versionErr := json.Unmarshal(result, &upgradeVersion)
		if json2versionErr != nil {
			r.JSON(http.StatusNotFound, map[string]interface{}{"error": "json trans struct error"})
		}
	}
	var currentVersion model.Version
	result, err = client.Get(params["app"] + "@" + params["version"])
	if err == nil && result != nil {
		json2versionErr := json.Unmarshal(result, &currentVersion)
		if json2versionErr != nil {
			r.JSON(http.StatusNotFound, map[string]interface{}{"error": "json trans struct error"})
		}
	}
	upgrade := false
	appNameArr := strings.Split(upgradeVersion.Compatible, " ")
	for _, nameValue := range appNameArr {
		if nameValue == currentVersion.Name {
			upgrade = true
		}
	}
	if upgrade {
		upgradeTime, err := time.Parse("2006-01-02 15:04:05", upgradeVersion.Updated)
		if err != nil {
			r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Prase Time Error"})
		}
		currentTime := time.Now()
		timeResult := currentTime.After(upgradeTime)
		if timeResult {
			r.JSON(http.StatusOK, upgradeVersion)
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

func UpdateVersion(db *xorm.Engine, params martini.Params, version model.Version, r render.Render, client redis.Client, res *http.Request) {
	appId, errAppId := strconv.ParseInt(params["app"], 0, 64)
	if errAppId != nil {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
		return
	}
	if version.VersionId == 0 || version.Name == "" || version.Changed == "" || version.Url == "" || version.Client == "" || version.Compatible == "" {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid json body "})
		return
	}
	version.App = appId
	has, err := db.Where("app=?", appId).And("version_id=?", version.VersionId).Update(version)
	if err != nil {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Database Error"})
		return
	}
	if has == 0 {
		r.JSON(http.StatusNotFound, map[string]interface{}{"error": "Not found any version records"})
		return
	} else {
		versionJson, err := json.Marshal(version)
		if err != nil {
			r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "version struct to json error"})
			return
		}
		client.Set(params["app"]+"@"+strconv.Itoa(version.VersionId), versionJson)
		r.JSON(http.StatusOK, version)
		return
	}
}

func DelVersion(db *xorm.Engine, params martini.Params, r render.Render, client redis.Client, res *http.Request) {
	appId, err := strconv.ParseInt(params["app"], 0, 64)
	versionNumber, versionErr := strconv.Atoi(params["version"])
	if err != nil || versionErr != nil {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
		return
	}
	result := new(model.Version)
	result.App = appId
	result.VersionId = versionNumber
	has, err := db.Where("app=?", appId).And("version_id=?", versionNumber).Get(result)
	if err != nil {
		fmt.Println(err)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Datebase Error"})
		return
	}
	if has {
		deleteVersion := new(model.Version)
		deleteVersion.App = appId
		deleteVersion.VersionId = versionNumber
		affect, err := db.Delete(deleteVersion)
		if affect == 1 && err == nil {
			client.Del(params["app"] + "@" + params["version"])
			r.JSON(http.StatusOK, result)
			return
		}
	} else {
		r.JSON(http.StatusNotFound, map[string]interface{}{"error": "Application's ID is not exist!"})
		return
	}
}
