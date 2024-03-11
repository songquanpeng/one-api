package common

import "github.com/songquanpeng/one-api/common/helper"

var UsingSQLite = false
var UsingPostgreSQL = false
var UsingMySQL = false

var SQLitePath = "one-api.db"
var SQLiteBusyTimeout = helper.GetOrDefaultEnvInt("SQLITE_BUSY_TIMEOUT", 3000)
