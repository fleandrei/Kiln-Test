package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type DB interface {
	Setup() error
	GetLastDelegationTimestamps() string
	ReadDelegations() ([]byte, error)
	WriteNewDelegations([]Delegation) error
	Close()
}

type FileDb struct {
	dbFile                   *os.File
	lastDelegationTimestamps string
}

const (
	timestampsLength     = 20
	dbPath               = "db"
	dbFileInit           = "{\"lastTimestamp\":\"1970-01-01T00:00:00Z\",\"data\":[]}"
	newDelegationsOffset = len(dbFileInit) - 2 // after the first '['
	timestampsOffset     = 18
	zeroTimeStamp        = "1970-01-01T00:00:00Z"
)

func (db *FileDb) Setup() error {

	dbFile, err := os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0664)
	if err != nil {
		return fmt.Errorf("could not open db file %s, got error : %w", dbPath, err)
	}

	/* Get the size of the db file*/
	fileInfo, err := os.Stat(dbPath)
	if err != nil {
		return fmt.Errorf("could not get db file size, got %w", err)
	}

	data := []byte{}
	if fileInfo.Size() >= timestampsOffset+timestampsLength {
		/*Retrieve just needed data (i.e. the timestamps of the last delegation) from db*/
		data = make([]byte, timestampsOffset+timestampsLength)
		_, err = dbFile.Read(data)
		if err != nil {
			return fmt.Errorf("could not read data from db file %s, got error : %w", dbPath, err)
		}
	}

	/*If the file is empty*/
	if len(data) == 0 {
		/*Init the delegation file with "{\"lastTimestamp\":\"1970-01-01T00:00:00Z\",\"data\":[]}"*/
		_, err := dbFile.Write([]byte(dbFileInit))
		if err != nil {
			return fmt.Errorf("could not initiate DB file, got %w", err)
		}

		/*Setup lastDelegationTimestamps with "1970-01-01T00:00:00Z"*/
		db.lastDelegationTimestamps = zeroTimeStamp
	} else {
		/*Retrieve the timestamps of the last retrieved delegation*/
		db.lastDelegationTimestamps = string(data[timestampsOffset : timestampsOffset+timestampsLength])
	}

	/*Setup db file descriptor*/
	db.dbFile = dbFile
	return nil
}

func (db FileDb) GetLastDelegationTimestamps() string {
	return db.lastDelegationTimestamps
}

/*add new delegations at the top of the "data" json array.*/
func (db *FileDb) WriteNewDelegations(delegations []Delegation) error {
	/*marshall to byte json*/
	byteNewDelegations, err := json.Marshal(delegations)
	if err != nil {
		return fmt.Errorf("could not marshall json array encoded delegations, got error : %w", err)
	}

	/*remove '[' and ']' from delegations json array*/
	byteNewDelegations = byteNewDelegations[1 : len(byteNewDelegations)-1]

	/*If there already are delegations in the db, add ',' at the end to respect json format*/
	if db.lastDelegationTimestamps != zeroTimeStamp {
		byteNewDelegations = append(byteNewDelegations, []byte(",")[0])
	}

	/*Retrieve all data from file db*/
	currentData, err := os.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf("could not add delegation to the db : could not read the file")
	}

	/*Update the last updated timestamps*/
	updatedData := append(
		currentData[:timestampsOffset],
		append(
			[]byte(delegations[0].Timestamp),
			currentData[timestampsOffset+timestampsLength:]...,
		)...,
	)

	/*update delegations*/
	updatedData = append(
		updatedData[:newDelegationsOffset],
		append(
			byteNewDelegations,
			updatedData[newDelegationsOffset:]...,
		)...,
	)

	_, err = db.dbFile.WriteAt(updatedData, 0)
	if err != nil {
		return fmt.Errorf("could not add delegations to the db")
	}
	InfoLog.Printf("%d new delegations have been added to the DB\n", len(delegations))

	/*Update lastDelegationTimestamps*/
	db.lastDelegationTimestamps = delegations[0].Timestamp

	return nil
}

/*Return a byte containing all the json in the db*/
func (db FileDb) ReadDelegations() ([]byte, error) {
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not read db file %s, got error: %w", dbPath, err)
	}

	/*Return only the json array of delegations*/
	return data, nil
}

func (db FileDb) Close() {
	db.dbFile.Close()
}
