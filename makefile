APPNAME=delegateAPI
CONFIG_FILE=config.yaml

build:
	go build -o ${APPNAME} -buildvcs=false

run: build
	./${APPNAME}

cleanApp: 
	rm -f ${APPNAME}

cleanDB: # retrieve the DB path from config.yaml and remove the folder.
	rm -f db

clean: cleanApp cleanDB

test:
	go test
