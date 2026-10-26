.PHONY: run build clean

BIN_DIR		:=bin
APP_NAME	:=app
BIN			:=$(BIN_DIR)/$(APP_NAME)
LIBRARIES 	:=ed-dsa
LIB_DIR		:=pkg

run: build
	@./$(BIN)

build: lib
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN) ./cmd

lib:
	cd $(LIB_DIR)/$(LIBRARIES) && go build -buildmode=archive -o $(BIN_DIR)/$(LIBRARIES).a

clean:
	@rm -rf $(BIN_DIR)
