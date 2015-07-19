proj = kappa
binary = $(proj)
datadir = data
CWD=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))


default: build
build: **/*.go
	@echo "------------------"
	@echo " Building binary"
	@echo "------------------"
	@godep go build .

test: **/*.go
	@echo "------------------"
	@echo " test"
	@echo "------------------"
	@godep go test -coverprofile=$(CWD)/coverage.out

deps:
	@echo "------------------"
	@echo " Downloading deps"
	@echo "------------------"
	@godep get .

clean:
	@echo "------------------"
	@echo " Cleaning"
	@echo "------------------"
	@rm $(binary)

env: 
	@export LOGXI='*=DBG'
	@export LOGXI_COLORS='*=white,key=green+h,message=blue+h,TRC,DBG,WRN=red+h,INF=green,ERR=red+h,maxcol=1000'
	@export LOGXI_FORMAT='happy,t=2006-01-02 15:04:05.000000'
	@export GIN_MODE=release

ca: export LOGXI =*=DBG
ca: export LOGXI_COLORS = *=black,key=red+h,message=blue+h,TRC,DBG,WRN=red+h,INF=green,ERR=red+h,maxcol=1000
ca: export LOGXI_FORMAT = happy,t=2006-01-02 15:04:05.000000
ca: build
	./$(binary) init-ca

cert: export LOGXI =*=DBG
cert: export LOGXI_COLORS = *=black,key=red+h,message=blue+h,TRC,DBG,WRN=red+h,INF=green,ERR=red+h,maxcol=1000
cert: export LOGXI_FORMAT = happy,t=2006-01-02 15:04:05.000000
cert: build
	./$(binary) new-cert

run: export LOGXI =*=DBG
run: export LOGXI_COLORS = *=white,key=green+h,message=blue+h,TRC,DBG,WRN=red+h,INF=green,ERR=red+h,maxcol=1000
run: export LOGXI_FORMAT = happy,t=2006-01-02 15:04:05.000000
run: export GIN_MODE=release
run: env build
	@mkdir -p $(datadir)
	./$(binary) server --http-listen=:19022 --ssh-listen=:9022 -D=data --ssh-key=pki/private/localhost.key --ca-cert=pki/ca.crt

docker: export GOOS=linux
docker: export CGO_ENABLED=0
docker: export GOARCH=amd64
docker:
	@godep go build -a -installsuffix cgo -o news .
	@docker build -t kappa -f Dockerfile.scratch .

html:
	@echo "------------------"
	@echo " html report"
	@echo "------------------"
	@go tool cover -html=$(CWD)/coverage.out -o $(CWD)/coverage.html
	@open coverage.html

detail:
	@echo "------------------"
	@echo " detailed report"
	@echo "------------------"
	@gocov test | gocov report

report: test detail html


.PHONY: build deps clean run
