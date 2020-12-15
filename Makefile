GIT_COMMIT=$(shell git rev-list -1 HEAD | sed 's/^\(........\).*/\1/')

build:
	go build \
		-o s3packer \
		-ldflags "-X github.com/koblas/s3packer/version.GitCommit=$(GIT_COMMIT)" \
		cmd/s3packer/main.go
