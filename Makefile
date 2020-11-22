
fireberry: main.go
	env GOOS=linux GOARCH=arm GOARM=5 go build

deploy: fireberry
	./deploy.sh
clean:
	@rm -f fireberry

