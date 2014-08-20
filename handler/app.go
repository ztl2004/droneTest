package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	logmodel "github.com/arkors/log/model"
	"github.com/arkors/update/model"
	"github.com/go-martini/martini"
	"github.com/go-xorm/xorm"
	"github.com/hoisie/redis"
	"github.com/martini-contrib/render"
)

func CreateVersion(db *xorm.Engine, params martini.Params, version model.Version, r render.Render, client redis.Client, res *http.Request, logChan *chan string, sys_log_level string) {
	appId, err := strconv.ParseInt(params["app"], 0, 64)
	if err != nil {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
		return
	}

	logId, fromModel, rootLogId := getHeader(res)
	logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]开始检验http中的json数据是否完整 UPDATE:handler.CreateVersion LOG_START", fromModel, logId, rootLogId)

	if version.VersionId == 0 || version.Name == "" || version.Changed == "" || version.Url == "" || version.Client == "" || version.Compatible == "" || version.Updated == "" {
		logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]Invalid json body UPDATE:handler.CreateVersion LOG_END", fromModel, logId, rootLogId)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid json body "})
		return
	}
	logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]json数据完整验证完成,查询此App的版本号是否已经存在 UPDATE:handler.CreateVersion", fromModel, logId, rootLogId)

	versionBeforeInsert := new(model.Version)
	has, errDb := db.Where("app=?", appId).And("version_id=?", version.VersionId).Get(versionBeforeInsert)
	if has && errDb == nil {
		logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]The application's id already exist UPDATE:handler.CreateVersion LOG_END", fromModel, logId, rootLogId)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id already exist"})
		return
	} else {
		logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]App未注册版本，开始进行注册 UPDATE:handler.CreateVersion", fromModel, logId, rootLogId)
		version.App = appId
		_, err2 := db.Insert(version)
		if err2 != nil {
			logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]Database Insert error UPDATE:handler.CreateVersion LINE:65 LOG_END", fromModel, logId, rootLogId)
			r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Database error"})
			return
		} else {
			logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]注册成功，插入redis  UPDATE:handler.CreateVersion", fromModel, logId, rootLogId)
			//写入redis内存库
			versionJson, err := json.Marshal(version)
			if err != nil {
				r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "version struct to json error"})
				return
			}
			client.Set(params["app"]+"@"+strconv.Itoa(version.VersionId), versionJson)
			logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]插入成功，处理返回值  UPDATE:handler.CreateVersion LOG_END", fromModel, logId, rootLogId)
			r.JSON(http.StatusCreated, version)
			return
		}
	}
}

func GetVersion(db *xorm.Engine, params martini.Params, r render.Render, client redis.Client, res *http.Request, logChan *chan string, sys_log_level string) {
	appId, errAppId := strconv.ParseInt(params["app"], 0, 64)
	versionNumber, versionErr := strconv.Atoi(params["version"])
	if errAppId != nil && versionErr != nil {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
		return
	}
	logId, fromModel, rootLogId := getHeader(res)
	logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]开始检验http中的Header数据是否完整 UPDATE:handler.CreateVersion LOG_START", fromModel, logId, rootLogId)

	_, id := res.Header["X-Arkors-Application-Id"]
	_, Token := res.Header["X-Arkors-Application-Token"]
	if id != true || Token != true {
		logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]Invalid request header,it should be include 'id' and 'Token' UPDATE:handler.GetVersion LINE:112 LOG_END", fromModel, logId, rootLogId)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid request header,it should be include 'id' and 'Token'"})
		return
	}
	//获取内存库中所有的key值
	logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]从redis中读取数据 UPDATE:handler.GetVersion", fromModel, logId, rootLogId)
	keyAll, redisErr := client.Keys(params["app"] + "@*")
	if redisErr != nil {
		logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]redis error UPDATE:handler.GetVersion LOG_END", fromModel, logId, rootLogId)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "redis error"})
		return
	}
	//内存库key值不存在，查询mysql确认是否存在，如果存在，把数据重新插入到redis中
	if keyAll == nil {
		logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]内存库key值不存在，查询mysql确认是否存在，如果存在，把数据重新插入到redis中 UPDATE:handler.GetVersion", fromModel, logId, rootLogId)
		versionInMysql := new(model.Version)
		has, errDb := db.Where("app=?", appId).And("version_id=?", versionNumber).Get(versionInMysql)
		if errDb == nil && has {
			versionJson, err := json.Marshal(versionInMysql)
			if err != nil {
				logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]redis error UPDATE:handler.GetVersion LOG_END", fromModel, logId, rootLogId)
				r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "version struct to json error"})
				return
			}
			logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]把数据重新插入到redis UPDATE:handler.GetVersion LOG_END", fromModel, logId, rootLogId)
			client.Set(params["app"]+"@"+strconv.Itoa(versionNumber), versionJson)
			return
		}
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "the version is newst!"})
		return
	}
	logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]获取最新版本信息 UPDATE:handler.GetVersion", fromModel, logId, rootLogId)
	var versionArray = make([]int, len(keyAll))
	for i, _ := range versionArray {
		versionArray[i], _ = strconv.Atoi(strings.Split(keyAll[i], "@")[1])
	}
	sort.Ints(versionArray)
	if versionArray[len(keyAll)-1] == versionNumber {
		logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]redis error UPDATE:handler.GetVersion LOG_END", fromModel, logId, rootLogId)
		r.JSON(http.StatusNotFound, map[string]interface{}{"error": "The application version is newest!"})
		return
	}
	var upgradeVersion model.Version
	result, err := client.Get(params["app"] + "@" + strconv.Itoa(versionArray[len(keyAll)-1]))
	if err == nil && result != nil {
		json2versionErr := json.Unmarshal(result, &upgradeVersion)
		if json2versionErr != nil {
			logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]json trans struct error UPDATE:handler.GetVersion LOG_END", fromModel, logId, rootLogId)
			r.JSON(http.StatusNotFound, map[string]interface{}{"error": "json trans struct error"})
		}
	}
	var currentVersion model.Version
	result, err = client.Get(params["app"] + "@" + params["version"])
	if err == nil && result != nil {
		json2versionErr := json.Unmarshal(result, &currentVersion)
		if json2versionErr != nil {
			logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]json trans struct error UPDATE:handler.GetVersion LOG_END", fromModel, logId, rootLogId)
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
			logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]Prase Time Error UPDATE:handler.GetVersion LOG_END", fromModel, logId, rootLogId)
			r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Prase Time Error"})
		}
		currentTime := time.Now()
		timeResult := currentTime.After(upgradeTime)
		if timeResult {
			r.JSON(http.StatusOK, upgradeVersion)
			return
		} else {
			logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]The App version can't in right time to upgrade UPDATE:handler.GetVersion LOG_END", fromModel, logId, rootLogId)
			r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The App version can't in right time to upgrade"})
			return
		}
	} else {
		logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]The App version can't upgrade UPDATE:handler.GetVersion LOG_END", fromModel, logId, rootLogId)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The App version can't upgrade"})
		return
	}
}

func UpdateVersion(db *xorm.Engine, params martini.Params, version model.Version, r render.Render, client redis.Client, res *http.Request, logChan *chan string, sys_log_level string) {
	appId, errAppId := strconv.ParseInt(params["app"], 0, 64)
	if errAppId != nil {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
		return
	}
	logId, fromModel, rootLogId := getHeader(res)
	logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]开始检验http中的Header数据是否完整 UPDATE:handler.UpdateVersion LOG_START", fromModel, logId, rootLogId)
	if version.VersionId == 0 || version.Name == "" || version.Changed == "" || version.Url == "" || version.Client == "" || version.Compatible == "" {
		logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]Invalid json body UPDATE:handler.UpdateVersion LINE:261 LOG_END", fromModel, logId, rootLogId)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid json body "})
		return
	}
	version.App = appId
	logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]开始检验App是否存在 UPDATE:handler.UpdateVersion", fromModel, logId, rootLogId)
	has, err := db.Where("app=?", appId).And("version_id=?", version.VersionId).Update(version)
	if err != nil {
		logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]Database Error  UPDATE:handler.UpdateVersion LOG_END", fromModel, logId, rootLogId)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Database Error"})
		return
	}
	if has == 0 {
		logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]Not found any version records UPDATE:handler.UpdateVersion LINE:275", fromModel, logId, rootLogId)
		r.JSON(http.StatusNotFound, map[string]interface{}{"error": "Not found any version records"})
		return
	} else {
		versionJson, err := json.Marshal(version)
		if err != nil {
			logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]Not found any version records UPDATE:handler.UpdateVersion LINE:275", fromModel, logId, rootLogId)
			r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "version struct to json error"})
			return
		}
		client.Set(params["app"]+"@"+strconv.Itoa(version.VersionId), versionJson)
		r.JSON(http.StatusOK, version)
		return
	}
}

func DelVersion(db *xorm.Engine, params martini.Params, r render.Render, client redis.Client, res *http.Request, logChan *chan string, sys_log_level string) {
	appId, err := strconv.ParseInt(params["app"], 0, 64)
	versionNumber, versionErr := strconv.Atoi(params["version"])
	if err != nil || versionErr != nil {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
		return
	}
	logId, fromModel, rootLogId := getHeader(res)
	logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]开始检验删除的APP是否存在 UPDATE:handler.DelVersion LOG_START LINE:275", fromModel, logId, rootLogId)
	result := new(model.Version)
	result.App = appId
	result.VersionId = versionNumber
	has, err := db.Where("app=?", appId).And("version_id=?", versionNumber).Get(result)
	if err != nil {
		logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]Database Error UPDATE:handler.DelVersion", fromModel, logId, rootLogId)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Datebase Error"})
		return
	}
	if has {
		logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]查到记录，开始删除 UPDATE:handler.DelVersion", fromModel, logId, rootLogId)
		deleteVersion := new(model.Version)
		deleteVersion.App = appId
		deleteVersion.VersionId = versionNumber
		affect, err := db.Delete(deleteVersion)
		if affect == 1 && err == nil {
			logId = logmodel.WriteLog(logChan, sys_log_level, "DEBUG", appId, "[normal]删除成功，处理返回值 UPDATE:handler.DelVersion LOG_END", fromModel, logId, rootLogId)
			client.Del(params["app"] + "@" + params["version"])
			r.JSON(http.StatusOK, result)
			return
		}
	} else {
		logId = logmodel.WriteLog(logChan, sys_log_level, "ERROR", appId, "[error]Application's ID is not exist! UPDATE:handler.DelVersion LOG_END", fromModel, logId, rootLogId)
		r.JSON(http.StatusNotFound, map[string]interface{}{"error": "Application's ID is not exist!"})
		return
	}
}

func getHeader(res *http.Request) (string, string, string) {
	headerLog, _ := res.Header["X-Arkors-Application-Log"]
	//headerClient, _ := res.Header["X-Arkors-Application-Client"]
	var parentLogId string
	//var fromModel string
	logIdArr := headerLog[0]
	parentLogId = strings.Split(logIdArr, "|")[1]
	rootLogId := strings.Split(logIdArr, "|")[0]
	//fromModel = strings.Split(headerClient[0], ",")[1]

	return parentLogId, "OAUTH", rootLogId
}
