package model

import (
  _ "github.com/go-sql-driver/mysql"
)

type Version struct {
  Id         int64
  App        int64
  Version    int    `xorm:"int"`
  Name       string `xorm:"text"`
  Updated    string `xorm:"date"`
  Changed    string `xorm:"text"`
  Url        string `xorm:"text"`
  Client     string `xorm:"text"`
  Compatible string `xorm:"text"`
}
