package common

import "one-api/common/helper"

var UsingSQLite = false
var UsingPostgreSQL = false

var SQLitePath = "one-api.db"
var SQLiteBusyTimeout = helper.GetOrDefaultEnvInt("SQLITE_BUSY_TIMEOUT", 3000)
