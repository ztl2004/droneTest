package logInfo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type LEVEL int32

const (
	OFF LEVEL = iota
	ERROR
	WARN
	INFO
	DEBUG
	ALL
)

var LevelMapping = map[string]LEVEL{
	"OFF":   OFF,
	"ERROR": ERROR,
	"WARN":  WARN,
	"INFO":  INFO,
	"DEBUG": DEBUG,
	"ALL":   ALL,
}

type Log struct {
	App       int64
	Level     string
	Action    string
	FromModel string
	Log       string
	ParentLog string
}

func (loginfo *Log) WriteLog(logChan chan string, SysLogLevel LEVEL) {
	if SysLogLevel >= LevelMapping[loginfo.Level] {
		logJson, _ := json.Marshal(*loginfo)
		logChan <- string(logJson)
	}
}

func Sendlog(logChan chan string) {
	for {
		log_data := <-(logChan)
		var logModel Log
		json.Unmarshal([]byte(log_data), &logModel)
		request, _ := http.NewRequest("POST", "http://log.arkors.com/v1/log", bytes.NewReader([]byte(log_data)))
		request.Header.Set("X-Arkors-Application-Log", logModel.Log)
		request.Header.Set("X-Arkors-Application-Client", "127.0.0.1,UPDATE")
		request.Header.Set("Accept", "application/json")
		client := &http.Client{}
		resp, err := client.Do(request)
		if err != nil {
			fmt.Println(err)
			break
		}
		if resp.StatusCode == http.StatusOK {
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("read reponse body error", err)
			}
			fmt.Println(string(data))
		}
	}
}
