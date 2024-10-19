OUT = pagepal

all: build

run: all
	./$(OUT)

build:
	go mod tidy
	go build -o $(OUT)

clean:
	rm $(OUT)