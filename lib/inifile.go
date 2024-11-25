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

var Myconfig Config

func LoadIni(filename string) {
	inidata, err := ini.Load(filename)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	err = inidata.MapTo(&Myconfig)
	if err != nil {
		fmt.Printf("Fail to map file: %v", err)
		os.Exit(1)
	}
}
