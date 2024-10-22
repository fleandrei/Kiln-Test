package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type sender struct {
	Address string
}

type DelegationReponse struct {
	Timestamp string
	Amount    uint
	Sender    sender
	Level     uint
}

type Tezos interface {
	GetDelegationsFromTimestamp(timestamp string, limit uint) ([]DelegationReponse, error)
}

type TezosDriver struct {
}

/*retrieve delegations that have been commited after the timestamp*/
func (Td TezosDriver) GetDelegationsFromTimestamp(timestamp string, limit uint) ([]DelegationReponse, error) {
	var delegations []DelegationReponse
	requestURL := fmt.Sprintf("https://api.tzkt.io/v1/operations/delegations?timestamp.gt=%s&limit=%d", timestamp, limit)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request, got error: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error when sending request: %w", err)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body, got error: %w", err)
	}

	if err := json.Unmarshal(resBody, &delegations); err != nil {
		return nil, fmt.Errorf("could not unmarshall json reponse, got error: %w", err)
	}
	return delegations, nil
}
