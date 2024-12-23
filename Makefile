.PHONY=run
local-start:
	go run infrastructure/local/main.go start

.PHONY=stop
local-stop:
	go run infrastructure/local/main.go stop
