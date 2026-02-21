.PHONY: build test vet race clean

build:
	go build -o viewscreen .

test:
	go test ./...

vet:
	go vet ./...

race:
	go test -race ./...

clean:
	rm -f viewscreen coverage.out
