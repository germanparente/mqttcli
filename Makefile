all:
	go build -o mqttlights cmd/lights/main.go
	go build -o mqtttemperatures cmd/temperatures/main.go
	go build -o mqttplants cmd/plants/main.go
	env GOARCH=arm GOARM=6 go build -o teleinfo cmd/teleinfo/main.go

teleinfo:
	env GOARCH=arm GOARM=6 go build -o teleinfo cmd/teleinfo/main.go

clean:
	rm -rf mqttlights mqttplants mqtttemperatures teleinfo

#install:
#        rm -rf install
#        mkdir install
#	mkdir install/plants install/temperatures install/lights
#	cp mqttlights install/lights/
#	cp mqtttemperatures temperatures/
