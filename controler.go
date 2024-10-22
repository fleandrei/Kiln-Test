package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Controler struct {
	db      DB
	indexer Indexer
}

type getDelegationsResponse struct {
	Data []Delegation
}

func NewControler(db DB, indexer Indexer) Controler {
	return Controler{
		db:      db,
		indexer: indexer,
	}
}

func (c Controler) LaunchApi() {
	/*The endpoint to retrieve delegations*/
	http.HandleFunc("/xtz/delegations", c.getDelegations)

	server := &http.Server{
		Addr: ":8080",
	}

	go func() {
		InfoLog.Println("Starting the server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	serverClosed := make(chan os.Signal, 1)
	signal.Notify(serverClosed, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	/*Waiting the server to close*/
	<-serverClosed
	InfoLog.Println("Received intteruption signal: Closing Server...")

	c.closeServer(server)
}

func (c Controler) closeServer(server *http.Server) {
	c.indexer.Close()
	InfoLog.Println("Indexer closed")

	c.db.Close()
	InfoLog.Println("DB closed")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		panic(fmt.Sprintf("Server Shutdown Failed:%+v", err))
	}
}

func (c Controler) getDelegations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	/*Retrieve all data from store*/
	data, err := c.db.ReadDelegations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := getDelegationsResponse{}

	if err := json.Unmarshal(data, &response); err != nil {
		http.Error(w, "Failed to decode retrieved data", http.StatusInternalServerError)
		return
	}

	/*Encode the response data as JSON and write it to the w http.ResponseWriter*/
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
