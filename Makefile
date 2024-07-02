OUT = a2b-api-go
LOGFILE = a2b-api.log

# redirect stdout & stderr to $(LOGFILE) (append mode)
start: $(OUT)
	GIN_MODE=release ./$(OUT) >> $(LOGFILE) 2>&1 &
	pgrep $(OUT)

$(OUT): build

build:
	go build -o $(OUT) .

main: build

test:
	go test -v .

run:
	go run main.go

clean:
	rm -f $(OUT)

kill:
	kill -15 $(shell pgrep $(OUT))
