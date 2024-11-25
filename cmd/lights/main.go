package main

import (
	"fmt"
	"math/rand"
	"time"

	randomDataTime "github.com/duktig-solutions/go-random-date-generator"
	"github.com/germanparente/mqttcli/lib"
)

var period time.Duration = 60

const PAYLOADOFFICE = "shellies/shelly1pm-76BC1C/relay/0/command"
const PAYLOADKITCHEN = "shellies/shelly1-kitchen/relay/0/command"
const PAYLOADSPARE = "house/plug/three"

var statusoffice bool = false
var statuskitchen bool = false
var statusspare bool = false

var timeinit bool = false

var maxhour int = 19
var minhour int = 18

/* newtimes
type TimesOn struct {
	startTime time.Time
	endTime   time.Time
}


var timesLights map[string]TimesOn
*/

var starthouroffice int = 0
var startminuteoffice int = 0
var endhouroffice int = 0
var endminuteoffice int = 0

var starthourkitchen int = 0
var startminutekitchen int = 0
var endhourkitchen int = 0
var endminutekitchen int = 0

var starthourspare int = 0
var startminutespare int = 0
var endhourspare int = 0
var endminutespare int = 0

func isOfficeOn() bool {
	return statusoffice
}

func setOffice(onoff bool) {
	statusoffice = onoff
}

func isSpareOn() bool {
	return statusspare
}

func setSpare(onoff bool) {
	statusspare = onoff
}

func setTimeInit(argtimeinit bool) {
	timeinit = argtimeinit
}

func isKitchenOn() bool {
	return statuskitchen
}

func setKitchen(onoff bool) {
	statuskitchen = onoff
}

func isTimeInitd() bool {
	return timeinit
}

func turnOnOffice() {
	lib.MqttPublishValue(PAYLOADOFFICE, "on")
}

func turnOffOffice() {
	lib.MqttPublishValue(PAYLOADOFFICE, "off")
}

func turnOnKitchen() {
	lib.MqttPublishValue(PAYLOADKITCHEN, "on")
}

func turnOffKitchen() {
	lib.MqttPublishValue(PAYLOADKITCHEN, "off")
}

func turnOnSpare() {
	lib.MqttPublishValue(PAYLOADSPARE, "on")
}

func turnOffSpare() {
	lib.MqttPublishValue(PAYLOADSPARE, "off")
}

// new times
func newInitTimes() {
	_Datetime, err := randomDataTime.GenerateDateTime("2022-01-01 00:00:00", "2022-08-21 17:08:26")

	if err != nil {
		fmt.Println(_Datetime)
	}
	/*
		var x = TimesOn{ _Datetime , _Datetime }

		//timesLights{"office"} = x
	*/
}

func initTimes(label string, starth *int, endh *int, startm *int, endm *int) {
	var x string
	*starth = rand.Intn(maxhour-minhour) + minhour
	*endh = (*starth + rand.Intn(3) + 1) % 24
	*startm = rand.Intn(60)
	*endm = rand.Intn(60)
	x = fmt.Sprintf("TIME %s - %s start %02d:%02d - end %02d:%02d", time.Now().Format("2006.01.02 15:04:05"), label, *starth, *startm, *endh, *endm)
	fmt.Println(x)
	lib.SendMyMail(x)
}

func initTimesOffice() {
	initTimes("Office", &starthouroffice, &endhouroffice, &startminuteoffice, &endminuteoffice)
}

func initTimesKitchen() {
	initTimes("Kitchen", &starthourkitchen, &endhourkitchen, &startminutekitchen, &endminutekitchen)
}

func initTimesSpare() {
	initTimes("Spare", &starthourspare, &endhourspare, &startminutespare, &endminutespare)
}

func checkTimesOffice() bool {
	var tomorrow int = 0
	var yesterday int = 0
	var ret bool = false
	now := time.Now()
	if starthouroffice > endhouroffice {
		if now.Hour() > endhouroffice {
			tomorrow = 1
		} else {
			yesterday = 1
		}
	}
	starttimeoffice := time.Date(now.Year(), now.Month(), now.Day()-yesterday, starthouroffice, startminuteoffice, 0, 0, time.Local)
	endtimeoffice := time.Date(now.Year(), now.Month(), now.Day()+tomorrow, endhouroffice, endminuteoffice, 0, 0, time.Local)
	if now.Before(endtimeoffice) && now.After(starttimeoffice) {
		fmt.Printf("TIME %s - we should light on office\n", time.Now().Format("2006.01.02 15:04:05"))
		ret = true
	} else {
		fmt.Printf("TIME %s - we should light off office\n", time.Now().Format("2006.01.02 15:04:05"))
	}
	return ret
}

func checkTimesKitchen() bool {
	var tomorrow int = 0
	var yesterday int = 0
	var ret bool = false
	now := time.Now()
	if starthourkitchen > endhourkitchen {
		if now.Hour() > endhourkitchen {
			tomorrow = 1
		} else {
			yesterday = 1
		}
	}
	starttimekitchen := time.Date(now.Year(), now.Month(), now.Day()-yesterday, starthourkitchen, startminutekitchen, 0, 0, time.Local)
	endtimekitchen := time.Date(now.Year(), now.Month(), now.Day()+tomorrow, endhourkitchen, endminutekitchen, 0, 0, time.Local)
	if now.Before(endtimekitchen) && now.After(starttimekitchen) {
		fmt.Printf("TIME %s - we should light on kitchen\n", time.Now().Format("2006.01.02 15:04:05"))
		ret = true
	} else {
		fmt.Printf("TIME %s - we should light off kitchen\n", time.Now().Format("2006.01.02 15:04:05"))
	}
	return ret
}

func checkTimesSpare() bool {
	var tomorrow int = 0
	var yesterday int = 0
	var ret bool = false
	now := time.Now()
	if starthourspare > endhourspare {
		if now.Hour() > endhourspare {
			tomorrow = 1
		} else {
			yesterday = 1
		}
	}
	starttimespare := time.Date(now.Year(), now.Month(), now.Day()-yesterday, starthourspare, startminutespare, 0, 0, time.Local)
	endtimespare := time.Date(now.Year(), now.Month(), now.Day()+tomorrow, endhourspare, endminutespare, 0, 0, time.Local)
	if now.Before(endtimespare) && now.After(starttimespare) {
		fmt.Printf("TIME %s - we should light on spare\n", time.Now().Format("2006.01.02 15:04:05"))
		ret = true
	} else {
		fmt.Printf("TIME %s - we should light off spare\n", time.Now().Format("2006.01.02 15:04:05"))
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

	lib.LoadIni("lights.ini")
	lib.ConnectToMqtt()

	var st string

	// new times
	newInitTimes()
	initTimesOffice()
	initTimesKitchen()
	initTimesSpare()
	setTimeInit(true)
	setOffice(false)
	turnOffOffice()
	setKitchen(false)
	turnOffKitchen()
	setSpare(false)
	turnOffSpare()

	for {

		if checkTimesOffice() {
			if !isOfficeOn() {
				setOffice(true)
				turnOnOffice()
				st = fmt.Sprintf("TIME %s - OFFICE LIGHTS ON", time.Now().Format("2006.01.02 15:04:05"))
				fmt.Println(st)
				lib.SendMyMail(st)
			}
		} else {
			if isOfficeOn() {
				setOffice(false)
				turnOffOffice()
				st = fmt.Sprintf("TIME %s - OFFICE LIGHTS OFF", time.Now().Format("2006.01.02 15:04:05"))
				fmt.Println(st)
				lib.SendMyMail(st)
				setTimeInit(false)
			}
		}

		if checkTimesKitchen() {
			if !isKitchenOn() {
				setKitchen(true)
				turnOnKitchen()
				st = fmt.Sprintf("TIME %s - KITCHEN LIGHTS ON", time.Now().Format("2006.01.02 15:04:05"))
				fmt.Println(st)
				lib.SendMyMail(st)
			}
		} else {
			if isKitchenOn() {
				setKitchen(false)
				turnOffKitchen()
				st = fmt.Sprintf("TIME %s - KITCHEN LIGHTS OFF", time.Now().Format("2006.01.02 15:04:05"))
				fmt.Println(st)
				lib.SendMyMail(st)
				setTimeInit(false)
			}
		}

		if checkTimesSpare() {
			if !isSpareOn() {
				setSpare(true)
				turnOnSpare()
				st = fmt.Sprintf("TIME %s - SPARE LIGHTS ON", time.Now().Format("2006.01.02 15:04:05"))
				fmt.Println(st)
				lib.SendMyMail(st)
			}
		} else {
			if isSpareOn() {
				setSpare(false)
				turnOffSpare()
				st = fmt.Sprintf("TIME %s - SPARE LIGHTS OFF", time.Now().Format("2006.01.02 15:04:05"))
				fmt.Println(st)
				lib.SendMyMail(st)
				setTimeInit(false)
			}
		}

		if checkTimeInit() && !isTimeInitd() {
			initTimesKitchen()
			initTimesOffice()
			initTimesSpare()
			setTimeInit(true)
		}

		time.Sleep(time.Second * period)
	}
}
