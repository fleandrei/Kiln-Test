# Kiln-Test

This code retrieves delegations from api.tzkt.io, store them in a persistent storage and provide them in a json format via :
`GET localhost:8080/xtz/delegations`

## usage
The app can be launched via `make run`.
We can clean builds and db with `make clean`.


## retrieval strategy
When launching the api, the indexer will try to retrieve all delegations since the last retrieved delegation's timestamp. If the db is empty, it will retrieve all historical delegations.

Once this historical delegations retrieval stage is finished, the indexer start polling continuously the api to for new delegations.

## Configuration
A light configuration can be found in config.yaml:
- PollRate: Set the interval in seconds between 2 delegations retrieval.
- PollMaxSize: The max number of delegations per retrievals

## DB
Since this app only handle one type of data, I didn't use a relational DB. Instead, the persistent storage is a file named "db" which store delegations in a json format:

- lastTimestamp: the timestamps of the last retrieved delegation. This field is used at the app setup to determine from what point in the past we need to retrieve delegations.
- data: An array of delegations. They are sorted from the most recent one to the oldest one. This way, we don't have to sort delegations when we have to provide them via api.