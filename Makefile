NAMESPACE	= akeil.net/akeil
NAME    	= elsewhere
MAIN		= ./cmd/$(NAME)
BINDIR  	= ./bin

build:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/$(NAME) $(MAIN)

test:
	go test

src = $(wildcard *.go) $(wildcard ./*/*/*.go)

fmt: ${src}
	for file in $^ ; do\
		gofmt -w $${file} ; \
	done
