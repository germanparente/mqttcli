package lib

import (
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var client mqtt.Client

var isConnected bool = false

func IsMqttConnected() bool {

	return isConnected
}

func MqttPublish(topic string) {
	token := client.Publish(topic, 0, false, "")
	token.Wait()
	/*go func() {
		_ = token.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
		if token.Error() != nil {
			fmt.Printf(token.Error()) // Use your preferred logging technique (or just fmt.Printf)
		}
	}*/
}

func MqttPublishValue(topic string, value string) {

	token := client.Publish(topic, 0, false, value)
	token.Wait()
	/*
		go func() {
			_ = token.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
			if token.Error() != nil {
				fmt.Printf(token.Error()) // Use your preferred logging technique (or just fmt.Printf)
			}
		}*/
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
	isConnected = true
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
	isConnected = false
	client.Disconnect(250)
}

func MqttSubscribe(topic string, shandler mqtt.MessageHandler) {

	if token := client.Subscribe(topic, 0, shandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
}

func ConnectToMqtt() bool {

	opts := mqtt.NewClientOptions().AddBroker(Myconfig.MqttBroker.Brokerurl).SetClientID(Myconfig.MqttBroker.ClientID)
	opts.SetUsername(Myconfig.MqttBroker.Username)
	opts.SetPassword(Myconfig.MqttBroker.Password)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.SetPingTimeout(10 * time.Second)
	// opts.SetCleanSession(false)
	// opts.SetWill("FIN", "close", 0, true);
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		client = nil
		isConnected = false
	} else {
		isConnected = true
	}

	return isConnected
}
