BUILD_TIME = $(shell date -Iminutes)
GIT_COMMIT = $(shell git show --pretty=%h -s HEAD)
LDFLAGS="-X 'main.InfluxEndpoint=${INFLUX_ENDPOINT}' -X 'main.InfluxTags=${INFLUX_TAGS}' -X 'main.GitCommit=${GIT_COMMIT}' -X 'main.BuildTime=${BUILD_TIME}'"

default: 
	go build -ldflags ${LDFLAGS} -mod=vendor -o go-mbpool cmd/go-mbpool/main.go
pi:
	env GOOS=linux GOARCH=arm GOARM=7 go build -ldflags ${LDFLAGS} -mod=vendor -o go-mbpool.arm cmd/go-mbpool/main.go
pi6:
	env GOOS=linux GOARCH=arm GOARM=6 go build -ldflags ${LDFLAGS} -mod=vendor -o go-mbpool6.arm cmd/go-mbpool/main.go
win:
	env GOOS=windows GOARCH=amd64 go build -ldflags ${LDFLAGS} -mod=vendor -o go-mbpool.exe cmd/go-mbpool/main.go

.PHONY: clean
clean:
	rm -f ./go-mbpool ./main ./go-mbpool.arm ./go-mbpool6.arm ./go-mbpool.exe
