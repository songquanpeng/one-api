package common

var UsingSQLite = false
var UsingPostgreSQL = false

var SQLitePath = "one-api.db"
var SQLiteBusyTimeout = GetOrDefaultInt("SQLITE_BUSY_TIMEOUT", 3000)
