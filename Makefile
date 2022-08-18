MAIN_DIR = ./cmd/dsync
BINARY_NAME = dirsynchronizer

help:
	@echo "Possible targets:"
	@echo "- help    for printing this help message"
	@echo "- run     for building and running the project binary in the background"
	@echo "- build   for building the project"
	@echo "- debug   for running the project binary in the foreground with some debug options turned on"
	@echo "- clean   for cleaning the app main directory"
	@echo "For the 'run' and 'debug' targets the variables 'srcdir' and 'copydir' are required, i.e. they set the directories for synchronization."
	@echo "For example:"
	@echo "make run srcdir=/path/to/source/dir copydir=/path/to/mirror/dir"

build:
	@cd ${MAIN_DIR} && go build -o ${BINARY_NAME} main.go

run: build
	@${MAIN_DIR}/${BINARY_NAME} ${srcdir} ${copydir} &

debug:
	@go run -race ${MAIN_DIR}/main.go -log2std -loglvl=DEBUG ${srcdir} ${copydir}

clean:
	@cd ${MAIN_DIR} && go clean
	@rm -f ${MAIN_DIR}/${BINARY_NAME}
