package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/germanparente/mqttcli/lib"
)

var lightstarted bool = false
var waterstarted bool = false

var period time.Duration
var minmaxPeriod time.Duration
var lastTimeMinMaxPeriod time.Time

// shortcuts to config
func cfg() *lib.PlantsConfig {
	return &lib.MyPlantsConfig
}

func setMinMaxPeriod() {
	minmaxPeriod = time.Duration(cfg().Timing.MinMaxPeriod) * time.Minute
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
	if cfg().Features.Heating {
		lib.MqttPublishValue(cfg().Plugs.Heating, "on")
		lib.InfluxWriteString("HEATING", "unit", "ON")
		fmt.Println("Heating started")
	}
}

func stopHeating() {
	if cfg().Features.Heating {
		lib.MqttPublishValue(cfg().Plugs.Heating, "off")
		lib.InfluxWriteString("HEATING", "unit", "OFF")
		fmt.Println("Heating stopped")
	}
}

func startLightning() {
	if cfg().Features.Lightning {
		lib.MqttPublishValue(cfg().Plugs.Lightning, "on")
		fmt.Println("Light started")
		lib.InfluxWriteString("LIGHTNING", "unit", "ON")
	}
}

func stopLightning() {
	if cfg().Features.Lightning {
		lib.MqttPublishValue(cfg().Plugs.Lightning, "off")
		fmt.Println("Light stopped")
		lib.InfluxWriteString("LIGHTNING", "unit", "OFF")
	}
}

func startCooling() {
	if cfg().Features.Cooling {
		lib.MqttPublishValue(cfg().Plugs.Cooler, "on")
		lib.MqttPublishValue(cfg().Plugs.Cooler2, "on")
		fmt.Println("Cooling started")
		lib.InfluxWriteString("COOLER", "unit", "ON")
	}
}

func stopCooling() {
	if cfg().Features.Cooling {
		lib.MqttPublishValue(cfg().Plugs.Cooler, "off")
		lib.MqttPublishValue(cfg().Plugs.Cooler2, "off")
		fmt.Println("Cooling stopped")
		lib.InfluxWriteString("COOLER", "unit", "OFF")
	}
}

func startWatering() {
	if cfg().Features.Watering {
		lib.MqttPublishValue(cfg().Plugs.Watering, "on")
		fmt.Println("Watering started")
		lib.InfluxWriteString("WATER", "unit", "ON")
	}
}

func stopWatering() {
	if cfg().Features.Watering {
		lib.MqttPublishValue(cfg().Plugs.Watering, "off")
		fmt.Println("Watering stopped")
		lib.InfluxWriteString("WATER", "unit", "OFF")
	}
}

func inRanges() bool {

	var inranges bool = false
	now := time.Now()
	sched := &cfg().Schedule
	for index := 0; !inranges && index < len(sched.StartHours); index++ {
		starttime := time.Date(now.Year(), now.Month(), now.Day(), sched.StartHours[index], sched.StartMinutes[index], 0, 0, time.Local)
		endtime := time.Date(now.Year(), now.Month(), now.Day(), sched.EndHours[index], sched.EndMinutes[index], 0, 0, time.Local)
		if now.Before(endtime) && now.After(starttime) {
			fmt.Printf("In ranges is true\n")
			inranges = true
		}
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

func mailAlertError(temperature float64) {
	lib.SendMyMail(fmt.Sprintf("ERROR TEMPERATURA  %v", temperature))
}

func mailWatering(onoff string) {
	lib.SendMyMail(fmt.Sprintf("WATERING %s", onoff))
}

var subscribehandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {

	currentTime := time.Now()
	temperature := lib.GetFloatTemperature(string(msg.Payload()))
	fmt.Printf("TIME: %s - TEMPERATURE %v\n", currentTime.Format("2006.01.02 15:04:05"), temperature)

	temps := &cfg().Temperature
	// DS180 hack
	if temperature > -120.0 && temperature < 120.0 {
		if temperature > temps.MaxTemp {
			stopHeating()
			if temperature > temps.HighTemp {
				startCooling()
				if checkMinMaxPeriod() {
					mailAlertMax(temperature)
				}
			} else {
				// temperature between highest supported and high. Still need to cool.
				// hack: Not really. If temperature is high -3 , stop cooling
				if temperature < temps.HighTemp-3 {
					stopCooling()
				}
			}
		} else {
			stopCooling()
			if temperature < temps.MinTemp {
				startHeating()
				if temperature < temps.LowTemp {
					if checkMinMaxPeriod() {
						mailAlertMin(temperature)
					}
				}
			} else {
				fmt.Println("Temperature in the range")
			}
		}
	} else {
		mailAlertError(temperature)
	}

}

func main() {

	lib.LoadGenericIni("config.ini")
	lib.LoadPlantsIni("plants.ini")
	period = time.Duration(cfg().Timing.Period)
	setMinMaxPeriod()
	if lib.ConnectToMqtt() {
		lib.MqttSubscribe(cfg().Topics.Subscribe, subscribehandler)
	}
	// stop lights / watering at the beginning
	if cfg().Features.Lightning {
		stopLightning()
	}
	if cfg().Features.Watering {
		stopWatering()
	}
	for {
		if cfg().Features.Lightning {
			checkLights()
		}
		if cfg().Features.Watering {
			checkWatering()
		}
		time.Sleep(time.Second * period)

		if lib.IsMqttConnected() {
			lib.MqttPublish(cfg().Topics.Publish)
			fmt.Println("published")
		} else {
			fmt.Println(" Reconnecting ")
			if lib.ConnectToMqtt() {
				lib.MqttSubscribe(cfg().Topics.Subscribe, subscribehandler)
			}
		}

	}

}
