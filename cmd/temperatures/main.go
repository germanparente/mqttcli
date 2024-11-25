package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/germanparente/mqttcli/lib"
)

var period time.Duration = 60
var periodext time.Duration = 20

func mapTopic(topic string) string {

	var x string
	prefixIndex := strings.LastIndex(topic, "/")
	fmt.Printf("TOPIC %s\n", topic)
	var s = topic[prefixIndex+1:]
	fmt.Printf("substring %s\n", s)
	switch s {
	case "temp1":
		x = "OFFICE"
	case "temp2":
		x = "OFFICEROOF"
	case "ten":
		x = "LIVING"
	case "four":
		x = "BEDROOM"
	case "ds18b20g":
		x = "GARAGE"
	case "theipex":
		x = "GARAGEFLOOR"
	case "ds18b20":
		x = "LIVING2"
	case "bttemp2":
		x = "SHELTER"
	case "bttemp1":
		x = "EXTERNAL"
	case "pt100":
		x = "PT100"
	case "ttgo":
		x = "CO2"
	}

	return x
}

func temperaturePublish(topic string) {

	lib.MqttPublish(topic)
	location := mapTopic(topic)
	fmt.Printf("%s temperature requested\n", location)
}

var subscribeTemperaturehandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {

	location := mapTopic(msg.Topic())
	temperature := lib.GetFloatTemperature(string(msg.Payload()))
	// hack for ds180 errors
	if location != "PT100" {
		if temperature > 120.0 || temperature < -120.0 {
			temperature = 0.00
		}
	}
	// end hack
	currentTime := time.Now()
	fmt.Printf("TIME: %s - TEMPERATURE %s %v\n", currentTime.Format("2006.01.02 15:04:05"), location, temperature)
	lib.InfluxWriteFloat(location, "temperature", temperature)
}

var subscribeCO2handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	//      fmt.Printf("TOPIC: %s\n", msg.Topic())
	//      fmt.Printf("MSG: %s\n", msg.Payload())

	location := mapTopic(msg.Topic())
	co2, _ := strconv.Atoi(string(msg.Payload()))
	currentTime := time.Now()
	fmt.Printf("TIME: %s - CO2 %s %v\n", currentTime.Format("2006.01.02 15:04:05"), location, co2)
	lib.InfluxWriteInt(location, "co2", co2)
}

func mySubscribeCO2(topic string) {
	lib.MqttSubscribe(topic, subscribeCO2handler)
}

func mySubscribe(topic string) {
	lib.MqttSubscribe(topic, subscribeTemperaturehandler)
}

func getExtTemperature() string {
	var tempreturn string = ""
	req, err := http.NewRequest("GET", "https://www.romma.fr/station_24.php?tempe=1&pluie=&humi=&pressure=&vent=&rayonnement=&id=157", nil)
	if err != nil {
		// handle err
		fmt.Println("new request error")
		return tempreturn
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Android 7.0; Mobile; rv:54.0) Gecko/54.0 Firefox/54.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
		fmt.Printf("http error %s", err)
		return tempreturn
	}
	fmt.Println("Response status:", resp.Status)
	defer resp.Body.Close()

	var loopexit bool = false
	var tempfound bool = false
	var linehtml string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() && !loopexit {
		linehtml = scanner.Text()
		//fmt.Println(linehtml)
		if !tempfound {
			if strings.Contains(linehtml, "<td width=\"405\" colspan=\"5\" bgcolor=\"#A3BDE9\" class=\"grandeur\"><b>&nbsp;Temp") {
				tempfound = true
			}
		} else {
			//if strings.Contains(linehtml, "<td width=\"243\" colspan=\"3\" rowspan=\"2\" align=\"center\">") {
			if strings.Contains(linehtml, "<span class=\"bigTexte\">") {

				//fmt.Println("******************************************************")
				temperatures := strings.SplitAfter(linehtml, "<span class=\"bigTexte\">")
				temperature := strings.Split(temperatures[1], "</span>")
				//fmt.Println(temperature[0])
				//fmt.Println("******************************************************")
				tempreturn = temperature[0]
				loopexit = true
			}

		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	return tempreturn
}

func publishExternalTemperature() {

	var externaltemp = getExtTemperature()
	if externaltemp != "" {
		var result float64 = 0
		var err error
		if result, err = strconv.ParseFloat(externaltemp, 64); err == nil {
			currentTime := time.Now()
			fmt.Printf("TIME: %s - EXTERNAL TEMPERATURE (ROMMA) %v\n", currentTime.Format("2006.01.02 15:04:05"), result)
			lib.InfluxWriteFloat("ROMMA", "temperature", result)
			lib.MqttPublishValue("house/temperature/publish/romma", externaltemp)
		}
	}
}

func publishAllToMqtt() {

	temperaturePublish("house/temphumid/request/ten")
	temperaturePublish("house/temphumid/request/four")
	temperaturePublish("house/temphumid/request/three")
	temperaturePublish("house/temphumid/request/theipex")
	temperaturePublish("house/temphumid/request/ds18b20")
	temperaturePublish("house/temphumid/request/bttemp1")
	temperaturePublish("house/temphumid/request/bttemp2")
	temperaturePublish("house/temphumid/request/temp1")
	temperaturePublish("house/temphumid/request/temp2")
	temperaturePublish("house/temphumid/request/two")
	temperaturePublish("house/pt100/gettemp")
	temperaturePublish("house/co2/request/ttgo")

}

func subscribeToMqtt() bool {

	if lib.IsMqttConnected() {
		mySubscribe("house/temphumid/publish/three")
		mySubscribe("house/temphumid/publish/ten")
		mySubscribe("house/temperature/publish/four")
		mySubscribe("house/temperature/publish/theipex")
		mySubscribe("house/temperature/publish/ds18b20")
		mySubscribe("house/temperature/publish/bttemp1")
		mySubscribe("house/temperature/publish/bttemp2")
		mySubscribe("house/temperature/publish/temp1")
		mySubscribe("house/temperature/publish/temp2")
		mySubscribe("house/temperature/publish/two")
		mySubscribe("house/temperature/publish/ds18b20g")
		mySubscribe("house/temppt100/publish/pt100")
		mySubscribeCO2("house/co2/publish/ttgo")

	}

	return lib.IsMqttConnected()
}

func main() {

	lib.LoadIni("temperature.ini")

	if lib.ConnectToMqtt() {
		subscribeToMqtt()
	}

	starttime := time.Now()
	starttimeext := time.Now()

	for {
		current := time.Now()
		if current.Sub(starttime) > time.Second*period {
			if lib.IsMqttConnected() {
				publishAllToMqtt()
			}
			starttime = time.Now()
		}
		if current.Sub(starttimeext) > time.Second*periodext {
			if lib.IsMqttConnected() {
				publishExternalTemperature()
			}
			starttimeext = time.Now()
		}
		time.Sleep(2 * time.Second)

		if !lib.IsMqttConnected() {
			fmt.Println("Reconnecting")
			if lib.ConnectToMqtt() {
				subscribeToMqtt()
			}
		}

	}

}
