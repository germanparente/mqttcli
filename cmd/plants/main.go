package main

import (
	"fmt"
	"strconv"
	"strings"
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

func mailChickenDoor(status string) {
	lib.SendMyMail(fmt.Sprintf("CHICKEN DOOR %s", status))
}

func getChickenDoorTopic() string {
	if cfg().Topics.ChickenDoorStatus != "" {
		return cfg().Topics.ChickenDoorStatus
	}
	return cfg().Topics.ChickenDoor
}

func getChickenDoorRequestTopic() string {
	if cfg().Topics.ChickenDoorRequest != "" {
		return cfg().Topics.ChickenDoorRequest
	}
	return "chicken/door/request"
}

type ScheduledTime struct {
	Hour   int
	Minute int
}

func parseScheduledTime(s string) (ScheduledTime, bool) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return ScheduledTime{}, false
	}
	h, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	m, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return ScheduledTime{}, false
	}
	return ScheduledTime{Hour: h, Minute: m}, true
}

func getChickenDoorRequestTimes() []ScheduledTime {
	var times []ScheduledTime

	for _, entry := range cfg().Schedule.ChickenDoorTimes {
		for _, part := range strings.Split(entry, ",") {
			if st, ok := parseScheduledTime(part); ok {
				times = append(times, st)
			}
		}
	}
	return times
}

func requestChickenDoorStatus() {
	topic := getChickenDoorRequestTopic()
	lib.MqttPublishValue(topic, "status")
	fmt.Printf("TIME: %s - PUBLISHED REQUEST '%s' to '%s'\n", time.Now().Format("2006.01.02 15:04:05"), "status", topic)
}

var lastChickenDoorExec = make(map[string]string)

func checkChickenDoorRequest() {
	if !lib.IsMqttConnected() {
		return
	}
	now := time.Now()
	today := now.Format("2006-01-02")
	window := 15 * time.Minute
	if loopPeriod := period * time.Second; loopPeriod*2 > window {
		window = loopPeriod * 2
	}

	for _, st := range getChickenDoorRequestTimes() {
		key := fmt.Sprintf("%02d:%02d", st.Hour, st.Minute)
		target := time.Date(now.Year(), now.Month(), now.Day(), st.Hour, st.Minute, 0, 0, time.Local)

		// If current time has reached or passed the target time today, and is within the grace window
		if (now.Equal(target) || now.After(target)) && now.Sub(target) < window {
			if lastChickenDoorExec[key] != today {
				lastChickenDoorExec[key] = today
				requestChickenDoorStatus()
			}
		}
	}
}

var subscribeChickenDoorHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {

	currentTime := time.Now()
	status := strings.TrimSpace(string(msg.Payload()))
	fmt.Printf("TIME: %s - CHICKEN DOOR STATUS: %s\n", currentTime.Format("2006.01.02 15:04:05"), status)
	if status != "" {
		lib.InfluxWriteString("CHICKENDOOR", "status", status)
		mailChickenDoor(status)
	}
}

func subscribeToMqtt() {
	lib.MqttSubscribe(cfg().Topics.Subscribe, subscribehandler)
	if cfg().Features.ChickenDoor && getChickenDoorTopic() != "" {
		lib.MqttSubscribe(getChickenDoorTopic(), subscribeChickenDoorHandler)
	}
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
		subscribeToMqtt()
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
		if cfg().Features.ChickenDoor {
			checkChickenDoorRequest()
		}
		time.Sleep(time.Second * period)

		if lib.IsMqttConnected() {
			lib.MqttPublish(cfg().Topics.Publish)
			fmt.Println("published")
		} else {
			fmt.Println(" Reconnecting ")
			if lib.ConnectToMqtt() {
				subscribeToMqtt()
			}
		}

	}

}
