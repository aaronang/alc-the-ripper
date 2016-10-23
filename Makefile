execs = master slave

.PHONY: clean $(execs)

all: $(execs)

$(execs):
	mkdir -p bin
	go build -v -o bin/$@ github.com/aaronang/cong-the-ripper/cmd/$@

clean:
	rm -rf bin
