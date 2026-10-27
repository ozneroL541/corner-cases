.PHONY: run build lib tidy test clean

BIN_DIR		:= bin
APP_NAME	:= app
LIBRARIES	:= ed-dsa
LIB_DIR		:= pkg

ifeq ($(OS),Windows_NT)
EXE_SUFFIX	:= .exe
MKDIR		= if not exist "$(1)" mkdir "$(1)"
RMDIR		= if exist "$(1)" rmdir /s /q "$(1)"
RMFILE		= if exist "$(1)" del /q "$(1)"
RUN		= $(BIN)
else
EXE_SUFFIX	:=
MKDIR		= mkdir -p "$(1)"
RMDIR		= rm -rf "$(1)"
RMFILE		= rm -f "$(1)"
RUN		= ./$(BIN)
endif

BIN		:= $(BIN_DIR)/$(APP_NAME)$(EXE_SUFFIX)

run: build
	@$(RUN)

build: lib
	@$(call MKDIR,$(BIN_DIR))
	@go build -o $(BIN) ./cmd

lib:
	@$(call MKDIR,$(LIB_DIR)/$(LIBRARIES)/bin)
	cd $(LIB_DIR)/$(LIBRARIES) && go build -buildmode=archive -o $(BIN_DIR)/$(LIBRARIES).a

tidy:
	@go mod tidy

test:
	@go test -v ./...

clean:
	@$(call RMDIR,$(BIN_DIR))
	@$(call RMFILE,test.json)
	@$(call RMFILE,cmd/test.json)
	@$(call RMDIR,$(LIB_DIR)/$(LIBRARIES)/bin)
