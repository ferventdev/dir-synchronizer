MAIN_DIR = ./cmd/dsync
BINARY_NAME = dirsynchronizer

help:
	@echo "Possible targets:"
	@echo "- help    	for printing this help message"
	@echo "- run     	for building and running the project binary in the background"
	@echo "- stop		for terminating all processes (via pkill), that were started by this project binary"
	@echo "- stoplast	for terminating the last process (via pkill), that was started by this project binary"
	@echo "- build   	for building the project"
	@echo "- debug   	for running the project binary in the foreground with some debug options turned on"
	@echo "- clean   	for cleaning the app main directory"
	@echo "- getpid 	for printing the PID(s) of all processes (via pgrep), that were started by this project binary"
	@echo "For the 'run' and 'debug' targets the variables 'srcdir' and 'copydir' are required, i.e. they set the directories for synchronization."
	@echo "For example:"
	@echo "make run srcdir=/path/to/source/dir copydir=/path/to/mirror/dir"

build:
	@cd ${MAIN_DIR} && go build -mod vendor -o ${BINARY_NAME} main.go

run: build
	@${MAIN_DIR}/${BINARY_NAME} -pid ${srcdir} ${copydir} &

debug:
	@go run -mod vendor -race ${MAIN_DIR}/main.go -scanperiod=1s -copydirs -pid -log2std -loglvl=DEBUG ${srcdir} ${copydir}

clean:
	@cd ${MAIN_DIR} && go clean
	@rm -f ${MAIN_DIR}/${BINARY_NAME}

getpid:
	@pgrep ${BINARY_NAME} || echo "No ${BINARY_NAME} processes found"

stop:
	@pkill -SIGTERM ${BINARY_NAME} || echo "No ${BINARY_NAME} processes found"

stoplast:
	@pkill -SIGTERM -n ${BINARY_NAME} || echo "No ${BINARY_NAME} processes found"