run-onchain:
	go build -o simulate-cli
	./simulate-cli -onchain

run-simulate:
	go build -o simulate-cli
	./simulate-cli 

run:
	go build -o simulate-cli
	./simulate-cli 

env:
	cat .env.example > .env
