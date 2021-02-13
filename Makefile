NAMESPACE	= akeil.net/akeil
NAME    	= elsewhere
MAIN		= ./cmd/$(NAME)
BINDIR  	= ./bin

build:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/$(NAME) $(MAIN)


pkgs = $(wildcard internal/*)

test:
	go test
	for package in $(pkgs) ; do\
		go test $(NAMESPACE)/$(NAME)/$$package ; \
	done

src = $(wildcard *.go) $(wildcard ./*/*/*.go)

fmt: ${src}
	for file in $^ ; do\
		gofmt -w $${file} ; \
	done
