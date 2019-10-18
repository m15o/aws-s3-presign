EXE_NAME=aws-s3-presign 
SOURCES=$(wildcard *.go)

.PHONY: clean 

$(EXE_NAME): $(SOURCES)
	go get -d
	go build -o $(EXE_NAME)

clean:
	rm -f $(EXE_NAME)

