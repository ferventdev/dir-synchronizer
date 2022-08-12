MAIN_DIR = ./cmd/dsync
BINARY_NAME = dsync

help:
	@echo "Possible targets:"
	@echo "- help    for printing this help message"
	@echo "- run     for building and running the project binary"
	@echo "- build   for building the project"
	@echo "- clean   for cleaning the app main directory"

build:
	@cd ${MAIN_DIR} && go build -o ${BINARY_NAME} main.go

run: build
	@${MAIN_DIR}/${BINARY_NAME}

clean:
	@cd ${MAIN_DIR} && go clean