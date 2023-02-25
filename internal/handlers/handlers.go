package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"poniatowski-dev-backend/internal/helper"
	"poniatowski-dev-backend/internal/internaldb"
	"poniatowski-dev-backend/internal/logging"
)

const (
	GET  = "GET"
	POST = "POST"
	// PUT    = "PUT"
	// DELETE = "DELETE"
	// PATCH = "PATCH"
)

type databaseExists struct {
	Frontend   string `json:"frontend"`
	Connection string `json:"connection"`
	Tables     string `json:"tables"`
}

func ExistHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/database/exist" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	switch r.Method {
	case GET:
		logging.LogIt("existHandler", "INFO", "get request received, for checking if the database exists")
		processNotFound := internaldb.CheckDB()
		if processNotFound != "exists" {
			w.WriteHeader(500)
			w.Header().Set("Content-Type", "application/text")
			_, writeErr := w.Write([]byte(processNotFound))
			if writeErr != nil {
				logging.LogIt("existHandler", "ERROR", "error writing response")
			}
		} else {
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "application/text")
			_, writeErr := w.Write([]byte("exists"))
			if writeErr != nil {
				logging.LogIt("existHandler", "ERROR", "error writing response")
			}
		}
	case POST:
		if r.Header.Get("Content-Type") == "application/json" {
			var dbExists databaseExists
			errJson := json.NewDecoder(r.Body).Decode(&dbExists)
			if errJson != nil {
				helper.ErrorResponse(w, "Bad Request: Wrong Content-Type provided", http.StatusBadRequest)
				logging.LogIt("existHandler", "ERROR", "error decoding json request")
			}
			logging.LogIt("existHandler", "INFO", dbExists.Frontend+" has "+dbExists.Connection+" as its connection/database status")
			if dbExists.Tables == "create" {
				processStatus, errInit := internaldb.InitDB()
				if errInit != nil {
					logging.LogIt("existHandler", "ERROR", "unable to initialize db. please review the logs for more details")
					w.WriteHeader(500)
					w.Header().Set("Content-Type", "application/text")
					_, writeErr := w.Write([]byte(dbExists.Tables))
					if writeErr != nil {
						logging.LogIt("existHandler", "INFO", "init failed: "+fmt.Sprintln(errInit))
					}
				}
				dbExists.Tables = processStatus
				w.WriteHeader(200)
				w.Header().Set("Content-Type", "application/text")
				_, writeErr := w.Write([]byte(dbExists.Tables))
				if writeErr != nil {
					logging.LogIt("existHandler", "INFO", "init completed")
				}
			}
		} else {
			helper.ErrorResponse(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		}
	default:
		http.Error(w, "405 Status Method Not Allowed.", http.StatusMethodNotAllowed)
	}
}

func CollectHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/collect" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	var collectionData internaldb.CollectionData
	switch r.Method {
	case GET:
		// provide function here, or send error back?
	case POST:
		headerContentType := r.Header.Get("Content-Type")
		if headerContentType != "application/json" {
			helper.ErrorResponse(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
			logging.LogIt("collectHandler", "WARNING", "invalid 'Content-Type' received")
			return
		}
		var unmarshalErr *json.UnmarshalTypeError
		decoder := json.NewDecoder(r.Body)
		//decoder.DisallowUnknownFields()
		err := decoder.Decode(&collectionData)
		if err != nil {
			if errors.As(err, &unmarshalErr) {
				helper.ErrorResponse(w, "Bad Request: Wrong Type provided for field: "+unmarshalErr.Field, http.StatusBadRequest)
				logging.LogIt("collectHandler", "WARNING", "Bad Request: Wrong Type provided for field: "+unmarshalErr.Field)
			} else {
				helper.ErrorResponse(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
				logging.LogIt("collectHandler", "WARNING", "Bad Request: "+fmt.Sprintln(err))
			}
			return
		}
		helper.ErrorResponse(w, "Exists", http.StatusOK)
		err = collectionData.InsertCollectedData()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logging.LogIt("collectHandler", "ERROR", "error inserting data into database: "+fmt.Sprintln(err))
		}
	default:
		http.Error(w, "405 Status Method Not Allowed.", http.StatusMethodNotAllowed)
		logging.LogIt("collectHandler", "WARNING", "received invalid method")
	}
}

func BannedHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/banned" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		helper.ErrorResponse(w, "Bad Request: ", http.StatusBadRequest)
		logging.LogIt("collectHandler", "WARNING", "invalid 'Content-Type' received")
	}
	var bannedData internaldb.BannedData
	switch r.Method {
	case GET:
		// get data from banned table
		errCheck := bannedData.BannedCheck(r.URL.Query().Get("banned"))
		if errCheck != nil {
			http.Error(w, errCheck.Error(), http.StatusInternalServerError)
			logging.LogIt("collectHandler", "ERROR", "error querying the database: "+fmt.Sprintln(errCheck))
			return
		}
		jsonFromDB, err := json.Marshal(bannedData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logging.LogIt("collectHandler", "ERROR", "error marshalling json data: "+fmt.Sprintln(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonFromDB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logging.LogIt("collectHandler", "ERROR", "error sending/writing data to user: "+fmt.Sprintln(err))
			return
		}
		return
	case POST:
		// provide function here, or send error back?
	default:
		http.Error(w, "405 Status Method Not Allowed.", http.StatusMethodNotAllowed)
		logging.LogIt("collectHandler", "WARNING", "received invalid method")
	}
}

func FetchAllCollectedHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/fetchall" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		helper.ErrorResponse(w, "Bad Request: ", http.StatusBadRequest)
		logging.LogIt("collectHandler", "WARNING", "invalid 'Content-Type' received")
	}
	var collectedData internaldb.CollectionData
	switch r.Method {
	case GET:
		// get data from banned table
		errCheck := collectedData.FetchAll()
		if errCheck != nil {
			http.Error(w, errCheck.Error(), http.StatusInternalServerError)
			logging.LogIt("collectHandler", "ERROR", "error querying the database: "+fmt.Sprintln(errCheck))
			return
		}
		jsonFromDB, err := json.Marshal(collectedData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logging.LogIt("collectHandler", "ERROR", "error marshalling json data: "+fmt.Sprintln(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonFromDB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logging.LogIt("collectHandler", "ERROR", "error sending/writing data to user: "+fmt.Sprintln(err))
			return
		}
		return
	case POST:
		// provide function here, or send error back?
	default:
		http.Error(w, "405 Status Method Not Allowed.", http.StatusMethodNotAllowed)
		logging.LogIt("collectHandler", "WARNING", "received invalid method")
	}
}
