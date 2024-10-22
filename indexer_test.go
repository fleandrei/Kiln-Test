package main

import (
	"log"
	"os"
	"testing"
)

/*Tezos mock*/
type TezosMock struct {
	DelegationsToReturn []DelegationReponse
}

func (tm TezosMock) GetDelegationsFromTimestamp(timestamp string, limit uint) ([]DelegationReponse, error) {
	return tm.DelegationsToReturn, nil
}

/*Db mock*/
type DbMock struct {
	AddedDelegations []Delegation
}

func (DB DbMock) Setup() error                        { return nil }
func (DB DbMock) GetLastDelegationTimestamps() string { return "" }
func (DB DbMock) ReadDelegations() ([]byte, error)    { return nil, nil }
func (DB DbMock) Close()                              {}
func (DB *DbMock) WriteNewDelegations(deleg []Delegation) error {
	DB.AddedDelegations = deleg
	return nil
}

func initLogger() {
	InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLog = log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func TestPollDelegations(t *testing.T) {
	cases := []struct {
		msg                         string
		returnedDelegationsResponse []DelegationReponse
		expectedDelegations         []Delegation
	}{
		{
			msg: "No delegations retrieved from Tezos",
		},
		{
			msg: "Only one delegation retrieved from Tezos",
			returnedDelegationsResponse: []DelegationReponse{
				{
					Timestamp: "2024",
					Amount:    1,
					Level:     1,
					Sender: sender{
						Address: "bob",
					},
				},
			},
			expectedDelegations: []Delegation{
				{
					Timestamp: "2024",
					Amount:    1,
					Level:     1,
					Sender:    "bob",
				},
			},
		},
		{
			msg: "Severals delegations retrieved from Tezos",
			returnedDelegationsResponse: []DelegationReponse{
				{
					Timestamp: "2023",
					Amount:    1,
					Level:     1,
					Sender: sender{
						Address: "Alice",
					},
				},
				{
					Timestamp: "2023",
					Amount:    1,
					Level:     2,
					Sender: sender{
						Address: "Bob",
					},
				},
				{
					Timestamp: "2024",
					Amount:    1,
					Level:     3,
					Sender: sender{
						Address: "bob",
					},
				},
			},
			expectedDelegations: []Delegation{
				{
					Timestamp: "2024",
					Amount:    1,
					Level:     3,
					Sender:    "bob",
				},
				{
					Timestamp: "2023",
					Amount:    1,
					Level:     2,
					Sender:    "Bob",
				},
				{
					Timestamp: "2023",
					Amount:    1,
					Level:     1,
					Sender:    "Alice",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			initLogger()

			dbMock := DbMock{}
			tezosMock := TezosMock{
				DelegationsToReturn: c.returnedDelegationsResponse,
			}

			indexer := Indexer{
				db:    &dbMock,
				tezos: tezosMock,
			}

			indexer.pollDelegations()

			if !compareDelegationsArray(dbMock.AddedDelegations, c.expectedDelegations) {
				t.Errorf("delegations slice that have been writen to the DB (%+v) is not the expected one (%+v)", dbMock.AddedDelegations, c.expectedDelegations)
			}
			if len(c.expectedDelegations) > 0 && indexer.lastTimestamp != c.expectedDelegations[0].Timestamp {
				t.Errorf("indexer last timestamp (%s) is not equal to the most recent retrieved delegation timestamp (%s)", indexer.lastTimestamp, c.expectedDelegations[0].Timestamp)
			}
		})

	}
}

func compareDelegationsArray(deleg1 []Delegation, deleg2 []Delegation) bool {
	len1 := len(deleg1)
	if len(deleg2) != len1 {
		return false
	}
	if len1 == 0 {
		return true
	}
	for i := 0; i < len1; i++ {
		if !compareDelegation(deleg1[i], deleg2[i]) {
			return false
		}
	}
	return true
}

func compareDelegation(deleg1 Delegation, deleg2 Delegation) bool {
	return deleg1.Amount == deleg2.Amount &&
		deleg1.Level == deleg2.Level &&
		deleg1.Sender == deleg2.Sender &&
		deleg1.Timestamp == deleg2.Timestamp
}
