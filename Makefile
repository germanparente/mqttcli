all:
	go build -o mqttlights cmd/lights/main.go
	go build -o mqtttemperatures cmd/temperatures/main.go
	go build -o mqttplants cmd/plants/main.go

clean:
	rm -rf mqttlights mqttplants mqtttemperatures

#install:
#        rm -rf install
#        mkdir install
#	mkdir install/plants install/temperatures install/lights
#	cp mqttlights install/lights/
#	cp mqtttemperatures temperatures/
