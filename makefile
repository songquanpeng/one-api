NAME=one-api
DISTDIR=dist
WEBDIR=web
VERSION=$(shell git describe --tags || echo "dev")
GOBUILD=go build -ldflags "-s -w -X 'one-api/common.Version=$(VERSION)'"

all: one-api

web: $(WEBDIR)/build

$(WEBDIR)/build: 
	cd $(WEBDIR) && npm install && REACT_APP_VERSION=$(VERSION) npm run build

one-api: web
	$(GOBUILD) -o $(DISTDIR)/$(NAME)

clean:
	rm -rf $(DISTDIR) && rm -rf $(WEBDIR)/build