build:
	rm bootstrap.zip && GOOS=linux go build -o bootstrap main.go && zip bootstrap.zip bootstrap