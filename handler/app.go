package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/arkors/update/logInfo"
	"github.com/arkors/update/model"
	"github.com/go-martini/martini"
	"github.com/go-xorm/xorm"
	"github.com/hoisie/redis"
	"github.com/martini-contrib/render"
)

func CreateVersion(db *xorm.Engine, params martini.Params, version model.Version, r render.Render, client redis.Client, res *http.Request, logChan chan string, sys_log_level logInfo.LEVEL) {
	appId, err := strconv.ParseInt(params["app"], 0, 64)
	if err != nil {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
		return
	}

	logModel := new(logInfo.Log)
	logModel.ParentLog, logModel.FromModel = getHeader(res)
	logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
	logModel.Level = "DEBUG"
	logModel.Action = "开始检验http中的json数据是否完整 UPDATE:handler.CreateVersion LOG_START"
	logModel.WriteLog(logChan, sys_log_level)
	logModel.ParentLog = logModel.Log

	if version.VersionId == 0 || version.Name == "" || version.Changed == "" || version.Url == "" || version.Client == "" || version.Compatible == "" || version.Updated == "" {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "ERROR"
		logModel.Action = "Invalid json body UPDATE:handler.CreateVersion LOG_END"
		logModel.WriteLog(logChan, sys_log_level)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid json body "})
		return
	}
	logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
	logModel.Level = "DEBUG"
	logModel.Action = "json数据完整验证完成,查询此App的版本号是否已经存在 UPDATE:handler.CreateVersion"
	logModel.WriteLog(logChan, sys_log_level)
	logModel.ParentLog = logModel.Log
	versionBeforeInsert := new(model.Version)

	has, errDb := db.Where("app=?", appId).And("version_id=?", version.VersionId).Get(versionBeforeInsert)
	if has && errDb == nil {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "ERROR"
		logModel.Action = "The application's id already exist UPDATE:handler.CreateVersion LOG_END"
		logModel.WriteLog(logChan, sys_log_level)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id already exist"})
		return
	} else {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "DEBUG"
		logModel.Action = "App未注册版本，开始进行注册 UPDATE:handler.CreateVersion"
		logModel.WriteLog(logChan, sys_log_level)
		logModel.ParentLog = logModel.Log

		version.App = appId
		_, err2 := db.Insert(version)
		if err2 != nil {
			logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
			logModel.Level = "ERROR"
			logModel.Action = "Database Insert error UPDATE:handler.CreateVersion LINE:65 LOG_END"
			logModel.WriteLog(logChan, sys_log_level)
			r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Database error"})
			return
		} else {
			logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
			logModel.Level = "DEBUG"
			logModel.Action = "注册成功，插入redis  UPDATE:handler.CreateVersion"
			logModel.WriteLog(logChan, sys_log_level)
			logModel.ParentLog = logModel.Log
			//写入redis内存库
			versionJson, err := json.Marshal(version)
			if err != nil {
				r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "version struct to json error"})
				return
			}
			client.Set(params["app"]+"@"+strconv.Itoa(version.VersionId), versionJson)
			logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
			logModel.Level = "DEBUG"
			logModel.Action = "插入成功，处理返回值  UPDATE:handler.CreateVersion LOG_END"
			logModel.WriteLog(logChan, sys_log_level)
			r.JSON(http.StatusCreated, version)
			return
		}
	}
}

func GetVersion(db *xorm.Engine, params martini.Params, r render.Render, client redis.Client, res *http.Request, logChan chan string, sys_log_level logInfo.LEVEL) {
	appId, errAppId := strconv.ParseInt(params["app"], 0, 64)
	versionNumber, versionErr := strconv.Atoi(params["version"])
	if errAppId != nil && versionErr != nil {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
		return
	}
	logModel := new(logInfo.Log)
	logModel.ParentLog, logModel.FromModel = getHeader(res)
	logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
	logModel.Level = "DEBUG"
	logModel.Action = "开始检验http中的Header数据是否完整 UPDATE:handler.CreateVersion LOG_START"
	logModel.WriteLog(logChan, sys_log_level)
	logModel.ParentLog = logModel.Log

	_, id := res.Header["X-Arkors-Application-Id"]
	_, Token := res.Header["X-Arkors-Application-Token"]
	if id != true || Token != true {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "ERROR"
		logModel.Action = "Invalid request header,it should be include 'id' and 'Token' UPDATE:handler.GetVersion LINE:112 LOG_END"
		logModel.WriteLog(logChan, sys_log_level)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid request header,it should be include 'id' and 'Token'"})
		return
	}
	//获取内存库中所有的key值
	logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
	logModel.Level = "DEBUG"
	logModel.Action = "从redis中读取数据 UPDATE:handler.GetVersion"
	logModel.WriteLog(logChan, sys_log_level)
	logModel.ParentLog = logModel.Log
	keyAll, redisErr := client.Keys(params["app"] + "@*")
	if redisErr != nil {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "ERROR"
		logModel.Action = "redis error UPDATE:handler.GetVersion LOG_END"
		logModel.WriteLog(logChan, sys_log_level)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "redis error"})
		return
	}
	//内存库key值不存在，查询mysql确认是否存在，如果存在，把数据重新插入到redis中
	if keyAll == nil {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "DEBUG"
		logModel.Action = "内存库key值不存在，查询mysql确认是否存在，如果存在，把数据重新插入到redis中 UPDATE:handler.GetVersion"
		logModel.WriteLog(logChan, sys_log_level)
		logModel.ParentLog = logModel.Log
		versionInMysql := new(model.Version)
		has, errDb := db.Where("app=?", appId).And("version_id=?", versionNumber).Get(versionInMysql)
		if errDb == nil && has {
			versionJson, err := json.Marshal(versionInMysql)
			if err != nil {
				logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
				logModel.Level = "ERROR"
				logModel.Action = "redis error UPDATE:handler.GetVersion LOG_END"
				logModel.WriteLog(logChan, sys_log_level)
				r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "version struct to json error"})
				return
			}
			logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
			logModel.Level = "DEBUG"
			logModel.Action = "把数据重新插入到redis UPDATE:handler.GetVersion LOG_END"
			logModel.WriteLog(logChan, sys_log_level)
			client.Set(params["app"]+"@"+strconv.Itoa(versionNumber), versionJson)
			return
		}
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "the version is newst!"})
		return
	}
	logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
	logModel.Level = "DEBUG"
	logModel.Action = "获取最新版本信息 UPDATE:handler.GetVersion"
	logModel.WriteLog(logChan, sys_log_level)
	logModel.ParentLog = logModel.Log
	var versionArray = make([]int, len(keyAll))
	for i, _ := range versionArray {
		versionArray[i], _ = strconv.Atoi(strings.Split(keyAll[i], "@")[1])
	}
	sort.Ints(versionArray)
	if versionArray[len(keyAll)-1] == versionNumber {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "ERROR"
		logModel.Action = "redis error UPDATE:handler.GetVersion LOG_END"
		logModel.WriteLog(logChan, sys_log_level)
		r.JSON(http.StatusNotFound, map[string]interface{}{"error": "The application version is newest!"})
		return
	}
	var upgradeVersion model.Version
	result, err := client.Get(params["app"] + "@" + strconv.Itoa(versionArray[len(keyAll)-1]))
	if err == nil && result != nil {
		json2versionErr := json.Unmarshal(result, &upgradeVersion)
		if json2versionErr != nil {
			logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
			logModel.Level = "ERROR"
			logModel.Action = "json trans struct error UPDATE:handler.GetVersion LOG_END"
			logModel.WriteLog(logChan, sys_log_level)
			r.JSON(http.StatusNotFound, map[string]interface{}{"error": "json trans struct error"})
		}
	}
	var currentVersion model.Version
	result, err = client.Get(params["app"] + "@" + params["version"])
	if err == nil && result != nil {
		json2versionErr := json.Unmarshal(result, &currentVersion)
		if json2versionErr != nil {
			logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
			logModel.Level = "ERROR"
			logModel.Action = "json trans struct error UPDATE:handler.GetVersion LOG_END"
			logModel.WriteLog(logChan, sys_log_level)
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
			logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
			logModel.Level = "ERROR"
			logModel.Action = "Prase Time Error UPDATE:handler.GetVersion LOG_END"
			logModel.WriteLog(logChan, sys_log_level)
			r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Prase Time Error"})
		}
		currentTime := time.Now()
		timeResult := currentTime.After(upgradeTime)
		if timeResult {
			r.JSON(http.StatusOK, upgradeVersion)
			return
		} else {
			logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
			logModel.Level = "ERROR"
			logModel.Action = "The App version can't in right time to upgrade UPDATE:handler.GetVersion LOG_END"
			logModel.WriteLog(logChan, sys_log_level)
			r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The App version can't in right time to upgrade"})
			return
		}
	} else {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "ERROR"
		logModel.Action = "The App version can't upgrade UPDATE:handler.GetVersion LOG_END"
		logModel.WriteLog(logChan, sys_log_level)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The App version can't upgrade"})
		return
	}
}

func UpdateVersion(db *xorm.Engine, params martini.Params, version model.Version, r render.Render, client redis.Client, res *http.Request, logChan chan string, sys_log_level logInfo.LEVEL) {
	appId, errAppId := strconv.ParseInt(params["app"], 0, 64)
	if errAppId != nil {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
		return
	}
	logModel := new(logInfo.Log)
	logModel.ParentLog, logModel.FromModel = getHeader(res)
	logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
	logModel.Level = "DEBUG"
	logModel.Action = "开始检验http中的Header数据是否完整 UPDATE:handler.UpdateVersion LOG_START"
	logModel.WriteLog(logChan, sys_log_level)
	logModel.ParentLog = logModel.Log
	if version.VersionId == 0 || version.Name == "" || version.Changed == "" || version.Url == "" || version.Client == "" || version.Compatible == "" {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "ERROR"
		logModel.Action = "Invalid json body UPDATE:handler.UpdateVersion LINE:261 LOG_END"
		logModel.WriteLog(logChan, sys_log_level)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid json body "})
		return
	}
	version.App = appId
	logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
	logModel.Level = "DEBUG"
	logModel.Action = "开始检验App是否存在 UPDATE:handler.UpdateVersion"
	logModel.WriteLog(logChan, sys_log_level)
	logModel.ParentLog = logModel.Log
	has, err := db.Where("app=?", appId).And("version_id=?", version.VersionId).Update(version)
	if err != nil {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "ERROR"
		logModel.Action = "Database Error  UPDATE:handler.UpdateVersion LOG_END"
		logModel.WriteLog(logChan, sys_log_level)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Database Error"})
		return
	}
	if has == 0 {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "ERROR"
		logModel.Action = "Not found any version records UPDATE:handler.UpdateVersion LINE:275"
		logModel.WriteLog(logChan, sys_log_level)
		r.JSON(http.StatusNotFound, map[string]interface{}{"error": "Not found any version records"})
		return
	} else {
		versionJson, err := json.Marshal(version)
		if err != nil {
			logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
			logModel.Level = "ERROR"
			logModel.Action = "Not found any version records UPDATE:handler.UpdateVersion LINE:275"
			logModel.WriteLog(logChan, sys_log_level)
			r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "version struct to json error"})
			return
		}
		client.Set(params["app"]+"@"+strconv.Itoa(version.VersionId), versionJson)
		r.JSON(http.StatusOK, version)
		return
	}
}

func DelVersion(db *xorm.Engine, params martini.Params, r render.Render, client redis.Client, res *http.Request, logChan chan string, sys_log_level logInfo.LEVEL) {
	appId, err := strconv.ParseInt(params["app"], 0, 64)
	versionNumber, versionErr := strconv.Atoi(params["version"])
	if err != nil || versionErr != nil {
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "The application's id must be numrical"})
		return
	}
	logModel := new(logInfo.Log)
	logModel.ParentLog, logModel.FromModel = getHeader(res)
	logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
	logModel.Level = "DEBUG"
	logModel.Action = "开始检验删除的APP是否存在 UPDATE:handler.DelVersion LOG_START"
	logModel.WriteLog(logChan, sys_log_level)
	logModel.ParentLog = logModel.Log
	result := new(model.Version)
	result.App = appId
	result.VersionId = versionNumber
	has, err := db.Where("app=?", appId).And("version_id=?", versionNumber).Get(result)
	if err != nil {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "ERROR"
		logModel.Action = "Database Error UPDATE:handler.DelVersion"
		logModel.WriteLog(logChan, sys_log_level)
		r.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Datebase Error"})
		return
	}
	if has {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "DEBUG"
		logModel.Action = "查到记录，开始删除 UPDATE:handler.DelVersion"
		logModel.WriteLog(logChan, sys_log_level)
		logModel.ParentLog = logModel.Log
		deleteVersion := new(model.Version)
		deleteVersion.App = appId
		deleteVersion.VersionId = versionNumber
		affect, err := db.Delete(deleteVersion)
		if affect == 1 && err == nil {
			logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
			logModel.Level = "DEBUG"
			logModel.Action = "删除成功，处理返回值 UPDATE:handler.DelVersion LOG_END"
			logModel.WriteLog(logChan, sys_log_level)
			logModel.ParentLog = logModel.Log
			client.Del(params["app"] + "@" + params["version"])
			r.JSON(http.StatusOK, result)
			return
		}
	} else {
		logModel.Log = getLogId(logModel.ParentLog, logModel.FromModel)
		logModel.Level = "ERROR"
		logModel.Action = "Application's ID is not exist! UPDATE:handler.DelVersion LOG_END"
		logModel.WriteLog(logChan, sys_log_level)

		r.JSON(http.StatusNotFound, map[string]interface{}{"error": "Application's ID is not exist!"})
		return
	}
}

func getHeader(res *http.Request) (string, string) {
	headerLog, _ := res.Header["X-Arkors-Application-Log"]
	//_, _ := res.Header["X-Arkors-Application-Client"]
	var parentLogId string
	//var fromModel string
	parentLogId = headerLog[0]
	//fromModel = strings.Split(headerClient[0], ",")[1]

	return parentLogId, "UPDATE"
}

func getLogId(parentLogId string, fromModel string) string {
	//加密的公式md5(sign(userkey)+app+时间戳)= Oauth 分配的 Secret Key
	md5String := fmt.Sprintf("%v%v%v", parentLogId, fromModel, string(time.Now().Unix()))
	h := md5.New()
	h.Write([]byte(md5String))
	logId := hex.EncodeToString(h.Sum(nil))
	return logId
}
