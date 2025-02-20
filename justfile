build:
  CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/dev-kit ./main.go

docs:
  go run scripts/docs/update-doc.go

scan:
  trufflehog git file://. --only-verified

install:
  go install ./...
