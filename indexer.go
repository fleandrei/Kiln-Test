package main

import (
	"fmt"
	"time"
)

type Indexer struct {
	db            DB
	tezos         Tezos
	endPolling    chan struct{}
	pollRate      uint
	pollMaxSize   uint
	lastTimestamp string
}

func NewIndexer(db DB, tezos Tezos, pollRate uint, pollMaxSize uint) Indexer {
	return Indexer{
		db:            db,
		tezos:         tezos,
		pollRate:      pollRate,
		pollMaxSize:   pollMaxSize,
		endPolling:    make(chan struct{}),
		lastTimestamp: db.GetLastDelegationTimestamps(),
	}
}

func (I Indexer) Launch() {
	go I.run()
}

func (I Indexer) Close() {
	I.endPolling <- struct{}{}
}

func (I Indexer) run() {
	if err := I.PollMissingPastDelegations(); err != nil {
		WarningLog.Printf("Got error during delegation history polling: %+v", err)
	}

	for {
		time.Sleep(time.Duration(int64(time.Second) * int64(I.pollRate)))
		select {
		case <-I.endPolling:
			close(I.endPolling)
			return
		default:
			I.pollDelegations()
		}
	}
}

/*poll from remote delegations that occured since last polled delegation's timestamps*/
func (I *Indexer) pollDelegations() {
	InfoLog.Printf("poll new delegations after timestamps %s\n", I.lastTimestamp)
	retrievedDelegations, err := I.tezos.GetDelegationsFromTimestamp(I.lastTimestamp, I.pollMaxSize)
	if err != nil {
		WarningLog.Printf("Could not poll delegations that have occured after %s, got error: %+v\n", I.lastTimestamp, err)
		return
	}

	/*No new delegations since last poll */
	if len(retrievedDelegations) == 0 {
		InfoLog.Printf("No new delegations since %s", I.lastTimestamp)
		return
	}

	/*convert and reverse delegations slice in order that it is sorted by their timestamps */
	delegations := formateDelegations(retrievedDelegations)

	/*Add ne delegations to DB*/
	if err := I.db.WriteNewDelegations(delegations); err != nil {
		WarningLog.Printf("Could not add delegations that have occured after %s to the DB, got error: %+v\n", I.lastTimestamp, err)
		return
	}

	/* update the last polled delegation timestamps*/
	I.lastTimestamp = delegations[0].Timestamp
}

func (I *Indexer) PollMissingPastDelegations() error {
	lastTimestamp := I.lastTimestamp
	var retrievedDelegations []DelegationReponse
	stepSize := 3000
	InfoLog.Printf("Retrieve delegtions since %s", lastTimestamp)
	for {
		delegations, err := I.tezos.GetDelegationsFromTimestamp(lastTimestamp, uint(stepSize))
		if err != nil {
			WarningLog.Printf("History polling: Could not poll delegations that have occured after %s, got error: %+v\n", lastTimestamp, err)
			break
		}

		/* merge slices*/
		retrievedDelegations = append(retrievedDelegations, delegations...)

		len := len(delegations)
		if len > 0 {
			InfoLog.Printf(">> %d past delegations have been retrieved, current timestamp: %s", len, delegations[len-1].Timestamp)
		}

		/*All delegations have been retrieved*/
		if len < stepSize {
			break
		}

		/*the last provided delegations is the most recent one*/
		lastTimestamp = delegations[len-1].Timestamp
	}

	/*No new delegations since last poll */
	if len(retrievedDelegations) == 0 {
		return fmt.Errorf("could no find any delegations since %s", I.lastTimestamp)
	}

	/* convert and reverse delegations slice in order that it is sorted by their timestamps */
	res := formateDelegations(retrievedDelegations)

	/*Write delegations to DB*/
	if err := I.db.WriteNewDelegations(res); err != nil {
		return fmt.Errorf("could not save delegations to the DB, got error: %w", err)
	}

	/*update the last polled delegation timestamps */
	I.lastTimestamp = res[0].Timestamp
	return nil
}

/*
formatRetrievedDelegations convert []DelegationReponse type to []delegation
It also revers the array so that most recent delegations are the first ones
*/
func formateDelegations(delegations []DelegationReponse) []Delegation {
	lenDelegations := len(delegations)
	formatedDelegations := make([]Delegation, lenDelegations)
	for i := 0; i < lenDelegations; i++ {
		formatedDelegations[lenDelegations-1-i] = Delegation{
			Timestamp: delegations[i].Timestamp,
			Amount:    delegations[i].Amount,
			Sender:    delegations[i].Sender.Address,
			Level:     delegations[i].Level,
		}
	}
	return formatedDelegations
}
