execs = master slave job_creator metrics_collector

.PHONY: clean test $(execs)

all: $(execs)

$(execs):
	mkdir -p bin
	go build -v -o bin/$@ github.com/aaronang/cong-the-ripper/cmd/$@

test:
	go test -tags aws ./...

clean:
	rm -rf bin
