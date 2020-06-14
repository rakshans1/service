PROJECTNAME=$(shell basename "$(PWD)")

# Go related variables.
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES="./cmd/sales-api"

# Redirect error output to a file, so we can show it in development mode.
STDERR=/tmp/.$(PROJECTNAME)-stderr.txt

# PID file will store the server process id when it's running on development mode
PID=/tmp/.$(PROJECTNAME)-api-server.pid

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

## start: Start in development mode. Auto-starts when code changes.
start:
	bash -c "trap 'make stop' EXIT; $(MAKE) compile start-server watch run='make compile start-server'"

## stop: Stop development mode.
stop: stop-server

start-server: stop-server
	@echo "  >  $(PROJECTNAME) is available"
	@-$(GOBIN)/$(PROJECTNAME) 2>&1 & echo $$! > $(PID)
	@cat $(PID) | sed "/^/s/^/  \>  PID: /"

stop-server:
	@-touch $(PID)
	@-kill `cat $(PID)` 2> /dev/null || true
	@-rm $(PID)

## watch: Run given command when code changes.
watch:
	find . -type f -name "*.go" | entr -p sh -c "$(run)"

restart-server: stop-server start-server

## compile: Compile the binary.
compile:
	@-touch $(STDERR)
	@-rm $(STDERR)
	@-$(MAKE) -s go-compile 2> $(STDERR)
	@cat $(STDERR) | sed -e '1s/.*/\nError:\n/'  | sed 's/make\[.*/ /' | sed "/^/s/^/     /" 1>&2

go-compile: go-build

go-build:
	@echo "  >  Building binary..."
	@GOBIN=$(GOBIN) go build -o $(GOBIN)/$(PROJECTNAME) $(GOFILES)

.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
