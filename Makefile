# File name
DIR := bin
PREFIX := race-the-web
TAG := $(shell git describe --always --dirty --tags)

# Build command
CMD_BUILD := go build -o $(DIR)/$(PREFIX)_$(TAG)_

# Environment variables for build
ENV_OSX64 := GOOS=darwin GOARCH=amd64
ENV_OSX32 := GOOS=darwin GOARCH=386
ENV_LIN64 := GOOS=linux GOARCH=amd64
ENV_LIN32 := GOOS=linux GOARCH=386
ENV_WIN64 := GOOS=windows GOARCH=amd64
ENV_WIN32 := GOOS=windows GOARCH=386

all: windows linux osx

build:
	@go build -o $(PREFIX)$(TAG) .

windows:
	@$(ENV_WIN64) $(CMD_BUILD)win64.exe
	@$(ENV_WIN32) $(CMD_BUILD)win32.exe
	@echo "Windows complete."

linux:
	@$(ENV_LIN64) $(CMD_BUILD)lin32.bin
	@$(ENV_LIN32) $(CMD_BUILD)lin64.bin
	@echo "Linux complete."

osx:
	@$(ENV_OSX64) $(CMD_BUILD)osx64.app
	@$(ENV_OSX32) $(CMD_BUILD)osx32.app
	@echo "OSX complete."