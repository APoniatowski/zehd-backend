package internaldb

import (
	"database/sql"
	"fmt"
	"os"
	"zehd-backend/internal/logging"
	"strconv"
	"time"

	. "zehd-backend/internal"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/mitchellh/go-ps"

	"github.com/joho/godotenv"
)

// dbConfig this function is run within the initDB function, in order to connect, check, or create DB and table
func dbConfig() (map[string]string, error) {
	defer logging.TrackTime("dbConfig", time.Now())
	errEnv := godotenv.Load("/usr/local/env/.env")
	if errEnv != nil {
		logging.LogIt("dbConfig", "ERROR", "error loading .env variables")
	}
	conf := make(map[string]string)
	var emptyVar error
	partialErrMessage := "environment variable required but not set, check logs for more details"
	host, ok := os.LookupEnv(DbHost)
	if !ok {
		conf[DbHost] = "localhost"
		emptyVar = fmt.Errorf("dbhost " + partialErrMessage)
		logging.LogIt("dbConfig", "ERROR", "DBHOST environment variable empty or non-existent. Defaulting to 'localhost'")
	}
	port, ok := os.LookupEnv(DbPort)
	if !ok {
		conf[DbPort] = ""
		emptyVar = fmt.Errorf("dbport " + partialErrMessage)
		logging.LogIt("dbConfig", "ERROR", "Unable to configure DB, DBPORT empty.")
	}
	user, ok := os.LookupEnv(DbUser)
	if !ok {
		conf[DbUser] = ""
		emptyVar = fmt.Errorf("dbuser " + partialErrMessage)
		logging.LogIt("dbConfig", "ERROR", "Unable to configure DB, DBUSER empty.")
	}
	password, ok := os.LookupEnv(DbPass)
	if !ok {
		conf[DbPass] = ""
		emptyVar = fmt.Errorf("dbpass " + partialErrMessage)
		logging.LogIt("dbConfig", "ERROR", "Unable to configure DB, DBPASS empty.")
	}
	name, ok := os.LookupEnv(DbName)
	if !ok {
		conf[DbName] = ""
		emptyVar = fmt.Errorf("dbname " + partialErrMessage)
		logging.LogIt("dbConfig", "ERROR", "Unable to configure DB, DBNAME empty.")
	}
	if emptyVar != nil {
		return conf, emptyVar
	}
	conf[DbHost] = host
	conf[DbPort] = port
	conf[DbUser] = user
	conf[DbPass] = password
	conf[DbName] = name
	return conf, nil
}

// InitDB Initialize the DB
func InitDB() (string, error) {
	defer logging.TrackTime("InitDB", time.Now())
	var hostnameErr, err error
	Backend, hostnameErr = os.Hostname()
	if hostnameErr != nil {
		logging.LogIt("InitDb", "ERROR", "unable to configure database, unable to get hostname")
		return FailedStatus, hostnameErr
	}
	config, errConfig := dbConfig()
	if errConfig != nil {
		logging.LogIt("InitDb", "ERROR", "unable to configure database, 1 or more empty environment variables exists.")
		return "no env var", errConfig
	}
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s database=%s sslmode=disable",
		config[DbHost], config[DbPort],
		config[DbUser], config[DbPass], config[DbName])
	fmt.Printf("\nConnecting to DB server: ")
	Db, err = sql.Open("pgx", psqlInfo)
	if err != nil {
		logging.LogIt("InitDb", "ERROR", "unable to open a connection to database server.")
		fmt.Println("Connection failed!")
		fmt.Println(err)
		return "", err
	}
	fmt.Println("Connected successfully!")
	fmt.Printf("Pinging DB server: ")
	err = Db.Ping()
	if err != nil {
		logging.LogIt("InitDB", "ERROR", "unable to ping database.")
		fmt.Println("Ping failed!")
		return FailedStatus, err
	}
	fmt.Println("Pinged successfully!")
	fmt.Printf("Checking if database and table exists: ")
	query := "SELECT * FROM " + CollectTable + ";"
	_, tableCheck := Db.Query(query)
	if tableCheck != nil {
		fmt.Println("Not found")
		fmt.Printf("Creating new table: ")
		_, err = Db.Exec("CREATE TABLE " + CollectTable + "(" + CollectedTableColumns + ");")
		if err != nil {
			logging.LogIt("InitDb", "ERROR", "unable to create table ("+CollectTable+").")
			return FailedStatus, err
		}
		_, err = Db.Exec("CREATE TABLE " + CheckedTable + "(" + CheckedTableColumns + ");")
		if err != nil {
			logging.LogIt("InitDb", "ERROR", "unable to create table ("+CheckedTable+").")
			return FailedStatus, err
		}
		_, err = Db.Exec("CREATE TABLE " + BannedTable + "(" + BannedTableColumns + ");")
		if err != nil {
			logging.LogIt("InitDb", "ERROR", "unable to create table ("+BannedTable+").")
			return FailedStatus, err
		}
		fmt.Println("Created")
	} else {
		fmt.Println("Found")
	}
	return "exists", nil
}

// CheckDB Check if the DB exists
func CheckDB() (processNotFound string) {
	defer logging.TrackTime("CheckDB", time.Now())
	if DbHost == "localhost" {
		processList, err := ps.Processes()
		if err != nil {
			logging.LogIt("existHandler", "ERROR", "unable to list processes, in order to find postgres.")
		}
		for i := range processList {
			process := processList[i]
			if process.Executable() == "postgres" {
				logging.LogIt("existHandler", "INFO", "postgres was found. (pid: "+strconv.Itoa(process.Pid())+")")
			} else {
				processNotFound = FailedStatus
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
			config[DbHost], config[DbPort],
			config[DbUser], config[DbPass], config[DbName])
		db, err := sql.Open("pgx", psqlInfo)
		if err != nil {
			processNotFound = FailedStatus
			logging.LogIt("existHandler", "ERROR", "unable to open a connection to database server.")
		}
		defer func() {
			errClose := db.Close()
			if errClose != nil {
				logging.LogIt("main", "ERROR", "unable to close database")
			}
		}()
		_, dbCheck := db.Query("SELECT * FROM " + CollectTable + ";")
		if dbCheck != nil {
			processNotFound = "failed to query " + CollectTable
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

// InsertCollectedData Insert the collected data from frontends into the DB
func (collectedData *CollectionData) InsertCollectedData() error {
	defer logging.TrackTime("InsertCollectedData", time.Now())
	query := `
INSERT INTO collect_table (frontend, backend, ip, port, path, method, xforwardfor, xrealip, useragent, via, age, timedate, cfipcountry)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);`
	_, dbCheck := Db.Exec(query,
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

// BannedCheck Check the DB for the banned IP
func (bannedData *BannedData) BannedCheck(ipAddress string) error {
	defer logging.TrackTime("BannedCheck", time.Now())
	query := "SELECT * FROM " + BannedTable + " WHERE ip='$1';"
	bannedRows, dbCheck := Db.Query(query, ipAddress)
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

// FetchAll Fetch all collected data
func (collectedData *CollectionData) FetchAll() error {
	defer logging.TrackTime("FetchAll", time.Now())
	query := "select * from collect_data;"
	rows, dbCheck := Db.Query(query)
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
