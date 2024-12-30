package main

import (
	"fmt"
	"math/rand"
	"slices"
	"time"

	"github.com/germanparente/mqttcli/lib"
)

var period time.Duration = 60

const PAYLOADOFFICE = "shellies/shelly1pm-76BC1C/relay/0/command"
const PAYLOADKITCHEN = "shellies/shelly1-kitchen/relay/0/command"
const PAYLOADSPARE = "house/plug/three"

var timeinit bool = false

const OFF = "off"
const ON = "on"

/*
var maxhour int = 19
var minhour int = 18
var duration int = 3
*/

type Period struct {
	starthour   int
	endhour     int
	startminute int
	endminute   int
}

var timesLights = make(map[string]Period)

var statusLights = make(map[string]bool)

//var payloads = map[string]string{ {"office", PAYLOADOFFICE} , { "kitchen", PAYLOADKITCHEN }, { "spare", PAYLOADSPARE }}

func isOn(label string) bool {
	return statusLights[label]
}

func setONOFF(label string, onoff bool) {
	statusLights[label] = onoff
}

func setTimeInit(argtimeinit bool) {
	timeinit = argtimeinit
}

func isTimeInitd() bool {
	return timeinit
}

func turnOnOff(label string, onoff string) {
	index := slices.Index(lib.MyLightsConfig.Lights.Labels, label)
	if index != -1 {
		lib.MqttPublishValue(lib.MyLightsConfig.Lights.Payloads[index], onoff)
	}
}

func initTimes(label string) {
	var x Period
	var msg string

	x.starthour = rand.Intn(lib.MyLightsConfig.Hours.EndHour-lib.MyLightsConfig.Hours.StartHour) + lib.MyLightsConfig.Hours.StartHour
	x.endhour = (x.starthour + rand.Intn(lib.MyLightsConfig.Hours.Duration) + 1) % 24
	x.startminute = rand.Intn(60)
	x.endminute = rand.Intn(60)
	msg = fmt.Sprintf("TIME %s - %s start %02d:%02d - end %02d:%02d", time.Now().Format("2006.01.02 15:04:05"), label, x.starthour, x.startminute, x.endhour, x.endminute)
	timesLights[label] = x
	fmt.Println(msg)
	lib.SendMyMail(msg)
}

func checkTimes(label string) bool {
	var tomorrow int = 0
	var yesterday int = 0
	var ret bool = false
	now := time.Now()
	if timesLights[label].starthour > timesLights[label].endhour {
		if now.Hour() > timesLights[label].endhour {
			tomorrow = 1
		} else {
			yesterday = 1
		}
	}
	starttime := time.Date(now.Year(), now.Month(), now.Day()-yesterday, timesLights[label].starthour, timesLights[label].startminute, 0, 0, time.Local)
	endtime := time.Date(now.Year(), now.Month(), now.Day()+tomorrow, timesLights[label].endhour, timesLights[label].endminute, 0, 0, time.Local)
	if now.Before(endtime) && now.After(starttime) {
		fmt.Printf("TIME %s - we should light on %s\n", time.Now().Format("2006.01.02 15:04:05"), label)
		ret = true
	} else {
		fmt.Printf("TIME %s - we should light off %s\n", time.Now().Format("2006.01.02 15:04:05"), label)
	}
	return ret
}

func checkTimeInit() bool {
	var ret bool = false
	now := time.Now()
	starttime := time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.Local)
	endtime := time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0, 0, time.Local)
	if now.Before(endtime) && now.After(starttime) {
		ret = true
		fmt.Printf("TIME %s - is time to init\n", time.Now().Format("2006.01.02 15:04:05"))
	}
	return ret
}

func main() {

	lib.LoadGenericIni("config.ini")
	lib.LoadLightsIni("lights.ini")

	fmt.Println(lib.Myconfig)
	fmt.Println(lib.MyLightsConfig)
	fmt.Println(lib.MyLightsConfig.Lights.Labels[1])
	fmt.Println(lib.MyLightsConfig.Lights.Payloads[1])

	lib.ConnectToMqtt()

	var st string

	for _, thelabel := range lib.MyLightsConfig.Lights.Labels {
		initTimes(thelabel)
		setONOFF(thelabel, false)
	}
	setTimeInit(true)

	for _, thelabel := range lib.MyLightsConfig.Lights.Labels {
		turnOnOff(thelabel, OFF)
	}

	for {

		for _, thelabel := range lib.MyLightsConfig.Lights.Labels {
			if !checkTimes(thelabel) {
				if isOn(thelabel) {
					setONOFF(thelabel, false)
					turnOnOff(thelabel, OFF)
					st = fmt.Sprintf("TIME %s - %s LIGHTS OFF", time.Now().Format("2006.01.02 15:04:05"), thelabel)
					fmt.Println(st)
					lib.SendMyMail(st)
					setTimeInit(false)
				}
			} else {
				if !isOn(thelabel) {
					setONOFF(thelabel, true)
					turnOnOff(thelabel, ON)
					st = fmt.Sprintf("TIME %s - %s LIGHTS ON", time.Now().Format("2006.01.02 15:04:05"), thelabel)
					fmt.Println(st)
					lib.SendMyMail(st)
				}
			}
		}

		if checkTimeInit() && !isTimeInitd() {
			for _, thelabel := range lib.MyLightsConfig.Lights.Labels {
				initTimes(thelabel)
				setONOFF(thelabel, false)
			}
			setTimeInit(true)
		}

		time.Sleep(time.Second * period)
	}
}
