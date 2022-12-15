
---
1. Service implementation:
 - fix all <stub> occurrences
 - implement new features

---
2. Module init:
```
go mod init github.com/spendmail/stub
go mod tidy
```

---
3. Linting & Testing:
```
golangci-lint run .
golangci-lint run ./...
go test -v -count=1 -race -timeout=1m .
go test -v -race -count=100 .
go test -v -count=1 -timeout=30s -tags bench .
```

---
4. Checking module Availability
```
cd /tmp/
go get -v -d github.com/spendmail/stub@main
```

---
5. Local run:
```
make build
make launch
curl -X GET 'http://localhost:8888/path/100/hello/stub'
```


---
6. Run via docker-compose:
```
cd /tmp
git clone --branch main git@github.com:spendmail/go_stub.git stub
cd stub
make run
curl -X GET 'http://localhost:8888/path/100/hello/stub'
```
