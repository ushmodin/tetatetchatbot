all: build

build:
	CGO_ENABLED=0 GO111MODULE=on vgo build
