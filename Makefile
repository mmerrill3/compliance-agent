FLAGS =
APPNAME = compliance-agent
COMMONENVVAR = GOOS=linux GOARCH=amd64
BUILDENVVAR = CGO_ENABLED=0
HASHTAG = $(shell git rev-parse --short HEAD)
export GOPATH

deps:
	dep ensure -update

all : clean deps build

build:
	$(COMMONENVVAR) $(BUILDENVVAR) go build -o $(APPNAME) .


clean:
	rm -f $(APPNAME)
.PHONY: all container push clean
