package internal

import "database/sql"

const (
	GET  = "GET"
	POST = "POST"
	// PUT    = "PUT"
	// DELETE = "DELETE"
	// PATCH = "PATCH"
)

type DatabaseExists struct {
	Frontend   string `json:"frontend"`
	Connection string `json:"connection"`
	Tables     string `json:"tables"`
}

const (
	CollectTable = "collect_table"
	CheckedTable = "checked_table"
	BannedTable  = "banned_table"
	DbHost       = "DBHOST"
	DbPort       = "DBPORT"
	DbUser       = "DBUSER"
	DbPass       = "DBPASS"
	DbName       = "DBNAME"
)

// table columns constants
const (
	CollectedTableColumns = `unique_id SERIAL NOT NULL,
frontend TEXT,
backend TEXT,
ip TEXT,
port INT,
path TEXT,
method TEXT,
xforwardfor TEXT,
xrealip TEXT,
useragent TEXT,
via TEXT,
age TEXT,
timedate BIGINT,
checked BOOL,
banned BOOL,
cfipcountry TEXT,
PRIMARY KEY (unique_id)`
	CheckedTableColumns = `unique_id SERIAL NOT NULL,
ip TEXT,
domainname TEXT,
timechecked BIGINT,
PRIMARY KEY (unique_id)`
	BannedTableColumns = `unique_id SERIAL NOT NULL,
ip TEXT,
domainname TEXT,
timechecked BIGINT,
timebanned BIGINT,
PRIMARY KEY (unique_id)`
	FailedStatus = "failed"
)

var (
	Db      *sql.DB
	Backend string
)
