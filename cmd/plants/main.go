package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/germanparente/mqttcli/lib"
)

// #define MSOCKET

const lightningEnabled bool = false
const heatingEnabled bool = true
const coolingEnabled bool = true
const wateringEnabled bool = false

// #ifdef MSOCKET
const plugHeating = "house/msocket/request/set/one"
const plugCooler = "house/msocket/request/set/two"

// #else
// const plugHeating = "house/plug/three"
// const plugCooler = "house/plug/two"
// #endif

const plugLightning = "house/plug/ee"
const plugWatering = "house/msocket/request/set/four"

var starthours [5]int = [5]int{8, 21}
var startminutes [5]int = [5]int{0, 0}
var endhours [5]int = [5]int{8, 21}
var endminutes [5]int = [5]int{1, 1}
var lightstarted bool = false
var waterstarted bool = false

var maxtemp, mintemp float64 = 22.0, 17.0
var hightemp, lowtemp float64 = 33.0, 5.0
var period time.Duration = 8
var minmaxPeriod time.Duration = 5
var lastTimeMinMaxPeriod time.Time

func setMinMaxPeriod() {
	minmaxPeriod = minmaxPeriod * time.Minute
	lastTimeMinMaxPeriod = time.Now()
}

func checkMinMaxPeriod() bool {

	var finished bool = false
	if time.Since(lastTimeMinMaxPeriod) > minmaxPeriod {
		finished = true
		lastTimeMinMaxPeriod = time.Now()
	}
	return finished
}

func startHeating() {
	if heatingEnabled {
		lib.MqttPublishValue(plugHeating, "on")
		lib.InfluxWriteString("HEATING", "unit", "ON")
		fmt.Println("Heating started")
	}
}

func stopHeating() {
	if heatingEnabled {
		lib.MqttPublishValue(plugHeating, "off")
		fmt.Println("Heating started")
		lib.InfluxWriteString("HEATING", "unit", "OFF")
		fmt.Println("Heating stopped")
	}
}

func startLightning() {
	if lightningEnabled {
		lib.MqttPublishValue(plugLightning, "on")
		fmt.Println("Light started")
		lib.InfluxWriteString("LIGHTNING", "unit", "ON")
	}
}

func stopLightning() {
	if lightningEnabled {
		lib.MqttPublishValue(plugLightning, "off")
		fmt.Println("Light stopped")
		lib.InfluxWriteString("LIGHTNING", "unit", "OFF")
	}
}

func startCooling() {
	if coolingEnabled {
		lib.MqttPublishValue(plugCooler, "on")
		fmt.Println("Cooling started")
		lib.InfluxWriteString("COOLER", "unit", "ON")
	}
}

func stopCooling() {
	if coolingEnabled {
		lib.MqttPublishValue(plugCooler, "off")
		fmt.Println("Cooling stopped")
		lib.InfluxWriteString("COOLER", "unit", "OFF")
	}
}

func startWatering() {
	if wateringEnabled {
		lib.MqttPublishValue(plugWatering, "on")
		fmt.Println("Watering started")
		lib.InfluxWriteString("WATER", "unit", "ON")
	}
}

func stopWatering() {
	if wateringEnabled {
		lib.MqttPublishValue(plugWatering, "off")
		fmt.Println("Watering stopped")
		lib.InfluxWriteString("WATER", "unit", "OFF")
	}
}

func inRanges() bool {

	var inranges bool = false
	now := time.Now()
	index := 0
	for !inranges && index < len(starthours) {
		starttime := time.Date(now.Year(), now.Month(), now.Day(), starthours[index], startminutes[index], 0, 0, time.Local)
		endtime := time.Date(now.Year(), now.Month(), now.Day(), endhours[index], endminutes[index], 0, 0, time.Local)
		//     fmt.Printf("Times to check %s %s and now is %s\n",starttime.Format("2006.01.02 15:04:05"),endtime.Format("2006.01.02 15:04:05"),now.Format("2006.01.02 15:04:05"))
		if now.Before(endtime) && now.After(starttime) {
			fmt.Printf("In ranges is true\n")
			inranges = true
		}
		index++
	}

	return inranges
}

func checkLights() {

	var isinranges bool = inRanges()

	if isinranges {
		if !lightstarted {
			lightstarted = true
			startLightning()
		}
	} else {
		if lightstarted {
			lightstarted = false
			stopLightning()
		}
	}
}

func checkWatering() {

	var isinranges bool = inRanges()

	if isinranges {
		if !waterstarted {
			waterstarted = true
			mailWatering("ON")
		}
		startWatering()
		fmt.Println("watering on")
	} else {
		if waterstarted {
			waterstarted = false
			mailWatering("OFF")
		}
		stopWatering()
		fmt.Println("watering off")
	}

}

func mailAlertMax(temperature float64) {
	lib.SendMyMail(fmt.Sprintf("MAXIMA  %v", temperature))
}

func mailAlertMin(temperature float64) {
	lib.SendMyMail(fmt.Sprintf("MINIMA  %v", temperature))
}

func mailWatering(onoff string) {
	lib.SendMyMail(fmt.Sprintf("WATERING %s", onoff))
}

var subscribehandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	//	fmt.Printf("TOPIC: %s\n", msg.Topic())
	//	fmt.Printf("MSG: %s\n", msg.Payload())

	currentTime := time.Now()
	temperature := lib.GetFloatTemperature(string(msg.Payload()))
	fmt.Printf("TIME: %s - TEMPERATURE %v\n", currentTime.Format("2006.01.02 15:04:05"), temperature)
	if temperature > maxtemp {
		stopHeating()
		if temperature > hightemp {
			startCooling()
			if checkMinMaxPeriod() {
				mailAlertMax(temperature)
			}
		} else {
			// temperature between highest supported and high. Still need to cool.
			// hack: Not really. If temperature is high -3 , stop cooling
			if temperature < hightemp-3 {
				stopCooling()
			}
		}
	} else {
		stopCooling()
		if temperature < mintemp {
			startHeating()
			if temperature < lowtemp {
				if checkMinMaxPeriod() {
					mailAlertMin(temperature)
				}
			}
		} else {
			fmt.Println("Temperature in the range")
		}
	}
}

func main() {

	lib.LoadGenericIni("config.ini")
	lib.LoadPlantsIni("plants.ini")
	setMinMaxPeriod()
	if lib.ConnectToMqtt() {
		lib.MqttSubscribe("house/temperature/publish/bttemp2", subscribehandler)
	}
	// stop lights / watering at the beginning
	if lightningEnabled {
		stopLightning()
	}
	if wateringEnabled {
		stopWatering()
	}
	for {
		if lightningEnabled {
			checkLights()
		}
		if wateringEnabled {
			checkWatering()
		}
		time.Sleep(time.Second * period)

		if lib.IsMqttConnected() {
			lib.MqttPublish("house/temphumid/request/bttemp2")
			fmt.Println("published")
		} else {
			fmt.Println(" Reconnecting ")
			if lib.ConnectToMqtt() {
				lib.MqttSubscribe("house/temperature/publish/bttemp2", subscribehandler)
			}
		}

	}

}
