package main

import (
	"fmt"
	"log"
	"net/http"
	"zehd-backend/internal/handlers"
	"zehd-backend/internal/internaldb"
	"zehd-backend/internal/logging"
)

func main() {
	fmt.Printf("Initializing DB... ")
	_, err := internaldb.InitDB()
	if err != nil {
		fmt.Println("Failed.")
		logging.LogIt("main", "ERROR", "unable to initialize database on startup. please review the logs for more details")
	}
	fmt.Printf("Done.\n")

	http.HandleFunc("/database/exist", handlers.ExistHandler)
	http.HandleFunc("/api/collect", handlers.CollectHandler)
	http.HandleFunc("/api/banned", handlers.BannedHandler)

	fmt.Printf("Listening on port 8080.\n")
	log.Println(http.ListenAndServe(":8080", nil))
	fmt.Println("===============================================================================================")
}
