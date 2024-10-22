package main

import (
	"log"
	"os"
)

type Delegation struct {
	Timestamp string
	Amount    uint
	Sender    string
	Level     uint
}

var (
	InfoLog    *log.Logger
	WarningLog *log.Logger
	ErrorLog   *log.Logger
)

func main() {
	/*Setup loggers*/
	InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLog = log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	/*Config*/
	InfoLog.Println("Get config")
	config, err := getConfig()
	if err != nil {
		ErrorLog.Printf("Could not setup config: %s\n", err.Error())
		panic(err)
	}

	/*DB*/
	InfoLog.Println("Setup DB")
	db := FileDb{}
	db.Setup()

	/*Tezos*/
	tezos := TezosDriver{}

	/*Indexer*/
	InfoLog.Println("Setup Indexer")
	indexer := NewIndexer(&db, tezos, config.PollRate, config.PollMaxSize)
	indexer.Launch()

	/*Launch the API*/
	controler := NewControler(&db, indexer)
	controler.LaunchApi()
}
