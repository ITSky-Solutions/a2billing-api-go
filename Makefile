OUT = ./a2b-api-go
LOGFILE = ./a2b-api.log	

# redirect stdout & stderr to $(LOGFILE) (append mode)
start: $(OUT)
	$(OUT) >> $(LOGFILE) 2>&1

$(OUT): build

build:
	go build -o $(OUT) .

main: build

test:
	go test -v .

clean:
	rm -f $(OUT)
