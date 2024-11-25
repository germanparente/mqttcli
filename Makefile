all:
	go build -o mqttlights cmd/lights/main.go
	go build -o mqtttemperatures cmd/temperatures/main.go
	go build -o mqttplants cmd/plants/main.go

clean:
	rm -rf mqttlights mqttplants mqtttemperatures
	
