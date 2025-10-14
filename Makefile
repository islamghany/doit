# include dev.env

# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

WEB_APP_VERSION = 0.0.1
WEB_APP_NAME = doit


run:
	go run cmd/doit/main.go