package lib

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

type Config struct {
	MqttBroker struct {
		Brokerurl string `ini:"brokerurl"`
		Username  string `ini:"username"`
		Password  string `ini:"password"`
		ClientID  string `ini:"clientid"`
	} `ini:"Broker"`
	Email struct {
		Address  string `ini:"address"`
		Server   string `ini:"server"`
		Password string `ini:"password"`
	} `ini:"Email"`
	InfluxDB struct {
		Url    string `ini:"url"`
		Token  string `ini:"token"`
		Bucket string `ini:"bucket"`
		Org    string `ini:"org"`
	} `ini:"InfluxDB"`
}

type LightsConfig struct {
	Lights struct {
		MqttClientID string   `ini:"mqttclientid"`
		Labels       []string `ini:"labels"`
		Payloads     []string `ini:"payloads"`
	} `ini:"Lights"`
	Hours struct {
		StartHour int `ini:"starthour"`
		EndHour   int `ini:"endhour"`
		Duration  int `ini:"duration"`
	} `ini:"Hours"`
}

type TempsConfig struct {
	Temperatures struct {
		MqttClientID string `ini:"mqttclientid"`
	} `ini:"Temperatures"`
}

type PlantsConfig struct {
	Plants struct {
		MqttClientID string `ini:"mqttclientid"`
	} `ini:"Plants"`
}

type TeleinfoConfig struct {
	Teleinfo struct {
		MqttClientID string   `ini:"mqttclientid"`
		ColorsToOpen []string `ini:"colorstoopen"`
		Payload      string   `ini:"payload"`
	} `ini:"Teleinfo"`
}

var MyPlantsConfig PlantsConfig
var MyLightsConfig LightsConfig
var MyTempsConfig TempsConfig

var MyTeleinfoConfig TeleinfoConfig

var Myconfig Config

func LoadLightsIni(filename string) {
	inidata := loadIni(filename)
	err := inidata.MapTo(&MyLightsConfig)
	if err != nil {
		fmt.Printf("Fail to map file: %v", err)
		os.Exit(1)
	}
	// copy clientid from specific conf to global conf
	Myconfig.MqttBroker.ClientID = MyLightsConfig.Lights.MqttClientID
}

func LoadTempsIni(filename string) {
	inidata := loadIni(filename)
	err := inidata.MapTo(&MyTempsConfig)
	if err != nil {
		fmt.Printf("Fail to map file: %v", err)
		os.Exit(1)
	}
	// copy clientid from specific conf to global conf
	Myconfig.MqttBroker.ClientID = MyLightsConfig.Lights.MqttClientID
}

func LoadPlantsIni(filename string) {
	inidata := loadIni(filename)
	err := inidata.MapTo(&MyPlantsConfig)
	if err != nil {
		fmt.Printf("Fail to map file: %v", err)
		os.Exit(1)
	}
	// copy clientid from specific conf to global conf
	Myconfig.MqttBroker.ClientID = MyPlantsConfig.Plants.MqttClientID
}

func LoadTeleinfoIni(filename string) {
	inidata := loadIni(filename)
	err := inidata.MapTo(&MyTeleinfoConfig)
	if err != nil {
		fmt.Printf("Fail to map file: %v", err)
		os.Exit(1)
	}
	// copy clientid from specific conf to global conf
	//Myconfig.MqttBroker.ClientID = MyTeleinfoConfig.Teleinfo.MqttClientID
}

func LoadGenericIni(filename string) {
	inidata := loadIni(filename)
	err := inidata.MapTo(&Myconfig)
	if err != nil {
		fmt.Printf("Fail to map file: %v", err)
		os.Exit(1)
	}
}

func loadIni(filename string) *ini.File {
	inidata, err := ini.Load(filename)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	return inidata
}
