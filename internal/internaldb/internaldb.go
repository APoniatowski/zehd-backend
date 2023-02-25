package internaldb

import (
	"database/sql"
	"fmt"
	"os"
	"poniatowski-dev-backend/internal/logging"
	"strconv"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/mitchellh/go-ps"

	"github.com/joho/godotenv"
)

// environment variable and table name constants
const (
	collectTable = "collect_table"
	checkedTable = "checked_table"
	bannedTable  = "banned_table"
	dbHost       = "DBHOST"
	dbPort       = "DBPORT"
	dbUser       = "DBUSER"
	dbPass       = "DBPASS"
	dbName       = "DBNAME"
)

// table columns constants
const (
	collectedTableColumns = `unique_id SERIAL NOT NULL,
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
	checkedTableColumns = `unique_id SERIAL NOT NULL,
ip TEXT,
domainname TEXT,
timechecked BIGINT,
PRIMARY KEY (unique_id)`
	bannedTableColumns = `unique_id SERIAL NOT NULL,
ip TEXT,
domainname TEXT,
timechecked BIGINT,
timebanned BIGINT,
PRIMARY KEY (unique_id)`
	failedStatus = "failed"
)

type CollectionData struct {
	FrontendName string `json:"frontendName"`
	TimeDate     int64  `json:"timeDate"`
	IP           string `json:"ip"`
	Port         int    `json:"port"`
	Path         string `json:"path"`
	Method       string `json:"method"`
	XForwardFor  string `json:"XForwardFor"`
	XRealIP      string `json:"XRealIP"`
	UserAgent    string `json:"useragent"`
	Via          string `json:"via"`
	Age          string `json:"age"`
	CFIPCountry  string `json:"CF-IPCountry"`
}

type BannedData struct {
	FrontendName    string `json:"frontendName"`
	TimeDateBanned  int64  `json:"timeDateBanned"`
	TimeDateChecked int64  `json:"timeDateChecked"`
	IP              string `json:"ip"`
	DomainName      string `json:"domainName"`
	Banned          bool   `json:"banned"`
}

var db *sql.DB
var Backend string

// dbConfig this function is run within the initDB function, in order to connect, check, or create DB and table
func dbConfig() (map[string]string, error) {
	errEnv := godotenv.Load("/usr/local/env/.env")
	if errEnv != nil {
		logging.LogIt("dbConfig", "ERROR", "error loading .env variables")
	}
	conf := make(map[string]string)
	var emptyVar error
	partialErrMessage := "environment variable required but not set, check logs for more details"
	host, ok := os.LookupEnv(dbHost)
	if !ok {
		conf[dbHost] = "localhost"
		emptyVar = fmt.Errorf("dbhost " + partialErrMessage)
		logging.LogIt("dbConfig", "ERROR", "DBHOST environment variable empty or non-existent. Defaulting to 'localhost'")
	}
	port, ok := os.LookupEnv(dbPort)
	if !ok {
		conf[dbPort] = ""
		emptyVar = fmt.Errorf("dbport " + partialErrMessage)
		logging.LogIt("dbConfig", "ERROR", "Unable to configure DB, DBPORT empty.")
	}
	user, ok := os.LookupEnv(dbUser)
	if !ok {
		conf[dbUser] = ""
		emptyVar = fmt.Errorf("dbuser " + partialErrMessage)
		logging.LogIt("dbConfig", "ERROR", "Unable to configure DB, DBUSER empty.")
	}
	password, ok := os.LookupEnv(dbPass)
	if !ok {
		conf[dbPass] = ""
		emptyVar = fmt.Errorf("dbpass " + partialErrMessage)
		logging.LogIt("dbConfig", "ERROR", "Unable to configure DB, DBPASS empty.")
	}
	name, ok := os.LookupEnv(dbName)
	if !ok {
		conf[dbName] = ""
		emptyVar = fmt.Errorf("dbname " + partialErrMessage)
		logging.LogIt("dbConfig", "ERROR", "Unable to configure DB, DBNAME empty.")
	}
	if emptyVar != nil {
		return conf, emptyVar
	}
	conf[dbHost] = host
	conf[dbPort] = port
	conf[dbUser] = user
	conf[dbPass] = password
	conf[dbName] = name
	return conf, nil
}

func InitDB() (string, error) {
	var hostnameErr, err error
	Backend, hostnameErr = os.Hostname()
	if hostnameErr != nil {
		logging.LogIt("InitDb", "ERROR", "unable to configure database, unable to get hostname")
		return failedStatus, hostnameErr
	}
	config, errConfig := dbConfig()
	if errConfig != nil {
		logging.LogIt("InitDb", "ERROR", "unable to configure database, 1 or more empty environment variables exists.")
		return "no env var", errConfig
	}
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s database=%s sslmode=disable",
		config[dbHost], config[dbPort],
		config[dbUser], config[dbPass], config[dbName])
	fmt.Printf("\nConnecting to DB server: ")
	db, err = sql.Open("pgx", psqlInfo)
	if err != nil {
		logging.LogIt("InitDb", "ERROR", "unable to open a connection to database server.")
		fmt.Println("Connection failed!")
		fmt.Println(err)
		return "", err
	}
	fmt.Println("Connected successfully!")
	fmt.Printf("Pinging DB server: ")
	err = db.Ping()
	if err != nil {
		logging.LogIt("InitDB", "ERROR", "unable to ping database.")
		fmt.Println("Ping failed!")
		return failedStatus, err
	}
	fmt.Println("Pinged successfully!")
	fmt.Printf("Checking if database and table exists: ")
	var query = "SELECT * FROM " + collectTable + ";"
	_, tableCheck := db.Query(query)
	if tableCheck != nil {
		fmt.Println("Not found")
		fmt.Printf("Creating new table: ")
		_, err = db.Exec("CREATE TABLE " + collectTable + "(" + collectedTableColumns + ");")
		if err != nil {
			logging.LogIt("InitDb", "ERROR", "unable to create table ("+collectTable+").")
			return failedStatus, err
		}
		_, err = db.Exec("CREATE TABLE " + checkedTable + "(" + checkedTableColumns + ");")
		if err != nil {
			logging.LogIt("InitDb", "ERROR", "unable to create table ("+checkedTable+").")
			return failedStatus, err
		}
		_, err = db.Exec("CREATE TABLE " + bannedTable + "(" + bannedTableColumns + ");")
		if err != nil {
			logging.LogIt("InitDb", "ERROR", "unable to create table ("+bannedTable+").")
			return failedStatus, err
		}
		fmt.Println("Created")
	} else {
		fmt.Println("Found")
	}
	return "exists", nil
}

func CheckDB() (processNotFound string) {
	if dbHost == "localhost" {
		processList, err := ps.Processes()
		if err != nil {
			logging.LogIt("existHandler", "ERROR", "unable to list processes, in order to find postgres.")
		}
		for i := range processList {
			process := processList[i]
			if process.Executable() == "postgres" {
				logging.LogIt("existHandler", "INFO", "postgres was found. (pid: "+strconv.Itoa(process.Pid())+")")
			} else {
				processNotFound = failedStatus
				logging.LogIt("existHandler", "INFO", "postgres not found. please make sure it is installed")
			}
		}
	} else {
		config, errConfig := dbConfig()
		if errConfig != nil {
			logging.LogIt("existHandler", "ERROR", "unable to configure database, 1 or more empty environment variables exists.")
		}
		psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
			"password=%s database=%s sslmode=disable",
			config[dbHost], config[dbPort],
			config[dbUser], config[dbPass], config[dbName])
		db, err := sql.Open("pgx", psqlInfo)
		if err != nil {
			processNotFound = failedStatus
			logging.LogIt("existHandler", "ERROR", "unable to open a connection to database server.")
		}
		defer func() {
			errClose := db.Close()
			if errClose != nil {
				logging.LogIt("main", "ERROR", "unable to close database")
			}
		}()
		_, dbCheck := db.Query("SELECT * FROM " + collectTable + ";")
		if dbCheck != nil {
			processNotFound = "failed to query " + collectTable
			logging.LogIt("existHandler", "ERROR", "database not found. please check your database server")
			pidQuery, pidCheck := db.Query("SELECT pg_backend_pid();")
			if pidCheck != nil {
				processNotFound = "failed to query pid"
				logging.LogIt("existHandler", "ERROR", "unable to obtain postgres pid. please check your database server")
			}
			var pid int
			errPidQuery := pidQuery.Scan(pid)
			if errPidQuery != nil {
				processNotFound = "no pid"
				logging.LogIt("existHandler", "ERROR", "unable to obtain postgres pid. pid query returned nil")
			}
			logging.LogIt("existHandler", "INFO", "database found. (pid:"+strconv.Itoa(pid)+")")
			logging.LogIt("existHandler", "ERROR", "postgres not found.")
		} else {
			result, initErr := InitDB() // TODO wrong result from here
			if initErr != nil {
				logging.LogIt("existHandler", "ERROR", "database initialization error after get request")
			}
			processNotFound = result
		}
	}
	return processNotFound
}

func (collectedData *CollectionData) InsertCollectedData() error {
	query := `
INSERT INTO collect_table (frontend, backend, ip, port, path, method, xforwardfor, xrealip, useragent, via, age, timedate, cfipcountry)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);`
	_, dbCheck := db.Exec(query,
		collectedData.FrontendName,
		Backend,
		collectedData.IP,
		collectedData.Port,
		collectedData.Path,
		collectedData.Method,
		collectedData.XForwardFor,
		collectedData.XRealIP,
		collectedData.UserAgent,
		collectedData.Via,
		collectedData.Age,
		collectedData.TimeDate,
		collectedData.CFIPCountry,
	)
	if dbCheck != nil {
		logging.LogIt("insertCollectedData", "ERROR", "unable to insert data into database")
	}
	return nil
}

func (bannedData *BannedData) BannedCheck(ipAddress string) error {
	query := "SELECT * FROM " + bannedTable + " WHERE ip='$1';"
	bannedRows, dbCheck := db.Query(query, ipAddress)
	if dbCheck != nil {
		logging.LogIt("bannedCheck", "ERROR", "unable to query db")
	}
	defer func() {
		errClose := bannedRows.Close()
		if errClose != nil {
			logging.LogIt("bannedCheck", "ERROR", "error closing query")
		}
	}()

	for bannedRows.Next() {
		errRows := bannedRows.Scan(
			&bannedData.IP,
			&bannedData.TimeDateBanned,
			&bannedData.TimeDateChecked,
			&bannedData.DomainName,
			&bannedData.Banned,
		)
		if errRows != nil {
			logging.LogIt("bannedCheck", "ERROR", "unable to scan rows")
		}
	}
	return nil
}

func (collectedData *CollectionData) FetchAll() error {
	query := "select * from collect_data;"
	rows, dbCheck := db.Query(query)
	if dbCheck != nil {
		logging.LogIt("bannedCheck", "ERROR", "unable to query db")
	}
	defer func() {
		errClose := rows.Close()
		if errClose != nil {
			logging.LogIt("bannedCheck", "ERROR", "error closing query")
		}
	}()
	// TODO  plonk everything into a slice, rather than write it to a singular struct
	for rows.Next() {
		errRows := rows.Scan(
			&collectedData.FrontendName,
			&collectedData.IP,
			&collectedData.Port,
			&collectedData.Path,
			&collectedData.Method,
			&collectedData.XForwardFor,
			&collectedData.XRealIP,
			&collectedData.UserAgent,
			&collectedData.Via,
			&collectedData.Age,
			&collectedData.TimeDate,
			&collectedData.CFIPCountry,
		)
		if errRows != nil {
			logging.LogIt("bannedCheck", "ERROR", "unable to scan rows")
		}
	}
	return nil
}
