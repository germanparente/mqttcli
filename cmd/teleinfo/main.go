package main

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"

	// TEST
	/*
	   "math/rand"
	   "os"
	*/
	// FINTEST
	"log"

	"github.com/germanparente/mqttcli/lib"
)

//============================= DB stuff =========================

func writeToDB(diffconso float64, colorday string, totalconso float64, sinsts int) {
	lib.InfluxWriteFloat("WH", "wh", diffconso)
	lib.InfluxWriteString("COLOR", "color", colorday)
	lib.InfluxWriteFloat("CONSO", "conso", totalconso)
	lib.InfluxWriteInt("SINSTS", "sinsts", sinsts)

}

//===============================================================

func listPorts() {

	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	for _, port := range ports {
		fmt.Printf("Found port: %v\n", port)
	}

}

type FTeleInfo struct {
	Description string
	Value       string
}

const period = 30

// TEST
// const path = "ptinfoutput"
// FIN TEST
const path = "/dev/ttyAMA0"

var serialport serial.Port

const startrecord byte = 2
const endrecord byte = 3

func initSerialPort() bool {
	var err error
	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.EvenParity,
		DataBits: 7,
		StopBits: serial.OneStopBit,
	}
	serialport, err = serial.Open(path, mode)
	if err != nil {
		fmt.Println("opening serial")
		fmt.Println(err)
		return false
	}
	return true
}

func closeSerialPort() {
	serialport.Close()
}
func getRecordTeleInfo() string {

	var record string
	buf := make([]byte, 1)
	var endofloop bool = false
	var started bool = false
	var first bool = false

	if !initSerialPort() {
		fmt.Println("Cannot open Serial port")
		return ""
	}
	defer closeSerialPort()

	for !endofloop {
		_, err := serialport.Read(buf)
		if err == io.EOF {
			fmt.Println("End of file")
			endofloop = true
			break
		}
		if err != nil {
			fmt.Println(err)
			endofloop = true
			continue
		}
		// to set in Debug
		// fmt.Printf("the buf [%c] [%i]\n", buf[0], buf[0])
		if startrecord == buf[0] {
			fmt.Println("STARTED !")
			started = true
			first = true
		} else if started {
			if first {
				first = false
			} else if endrecord == buf[0] {
				endofloop = true
			} else {
				record = record + string(buf[0])
			}
		}
	}

	return record
}

/*
func printSlice(s []string) {
	fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
	fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
}
*/

func setTeleInfo(record string) ([50]FTeleInfo, int) {

	var linesrecord []string
	var localteleinfo [50]FTeleInfo
	var i int
	var error int = 0
	var line string
	var x []string
	linesrecord = strings.Split(record, "\n")
	if len(linesrecord) == 1 {
		fmt.Println("setTeleInfo: bad record")
		error = -1
	} else {
		for i, line = range linesrecord {
			fmt.Printf("the line is = [%s] and the index is = [%d]\n", line, i)
			x = strings.Split(line, "\t")
			if len(x) == 1 {
				error = -1
				fmt.Println("setTeleInfo: bad line")
				break
			} else {
				localteleinfo[i].Description = x[0]
				localteleinfo[i].Value = x[1]
			}
		}
		if error == 0 {
			localteleinfo[i+1].Description = "END"
			localteleinfo[i+1].Value = "END"
		}
	}

	return localteleinfo, error
}

func getTotalConso(ti [50]FTeleInfo) float64 {
	var conso float64
	for _, telei := range ti {
		if telei.Description == "EAST" {
			conso, _ = strconv.ParseFloat(telei.Value, 64)
		}
	}
	return conso / 1000
}

func getSinsts(ti [50]FTeleInfo) int {
	var sinsts int
	for _, telei := range ti {
		if telei.Description == "SINSTS" {
			sinsts, _ = strconv.Atoi(telei.Value)
			break
		}
	}
	return sinsts
}

func getCurrentColor(ti [50]FTeleInfo) string {
	var color string
	for _, telei := range ti {
		if telei.Description == "LTARF" {
			color = strings.Replace(telei.Value, " ", "", -1)
			break
		}
	}
	return color
}

func dumpTeleinfo(ti [50]FTeleInfo) {
	for i := 0; ti[i].Description != "END"; i++ {
		fmt.Printf(" description=[%s] value=[%s]\n", ti[i].Description, ti[i].Value)
	}
}

func main() {

	var teleinfo [50]FTeleInfo
	var frame string
	var totalconso float64 = 0.0
	var currentcolor string
	var err int
	var diffconso float64 = 0.0
	var formerconso float64 = 0.0
	var formercolor string
	var sinsts int = 0

	lib.LoadGenericIni("config.ini")
	lib.LoadTeleinfoIni("teleinfo.ini")

	if !lib.ConnectToMqtt() {
		fmt.Println("Cannot connnect to mqtt")
	}
	listPorts()

	formercolor = "DUMMY"
	for {
		frame = getRecordTeleInfo()
		if frame != "" {
			teleinfo, err = setTeleInfo(frame)
			if err == 0 {
				dumpTeleinfo(teleinfo)
				totalconso = getTotalConso(teleinfo)
				currentcolor = getCurrentColor(teleinfo)
				sinsts = getSinsts(teleinfo)
				fmt.Printf("Total conso = [%f] current color = [%s] diff conso [%f]\n", totalconso, currentcolor, diffconso)
				fmt.Printf("The former color is [%s] the current color is [%s]\n", formercolor, currentcolor)
				fmt.Printf("The colors to open are %v", lib.MyTeleinfoConfig.Teleinfo.ColorsToOpen)
				if formercolor != currentcolor {
					formercolor = currentcolor
					if !lib.IsMqttConnected() {
						if !lib.ConnectToMqtt() {
							fmt.Println("Cannot connnect to mqtt")
						}
					}
					// color has changed. We need to test if switch has to be open.
					index := slices.Index(lib.MyTeleinfoConfig.Teleinfo.ColorsToOpen, currentcolor)
					if index != -1 {
						fmt.Println("setting chauffe on")
						lib.MqttPublishValue(lib.MyTeleinfoConfig.Teleinfo.Payload, "on")
					} else {
						fmt.Println("setting chauffe off")
						lib.MqttPublishValue(lib.MyTeleinfoConfig.Teleinfo.Payload, "off")
					}
				}

				if formerconso != 0.0 {
					diffconso = totalconso - formerconso
				}
				formerconso = totalconso
				if totalconso > 10000 && diffconso >= 0.0 {

					writeToDB(diffconso, currentcolor, totalconso, sinsts)
				}
			}
		}
		time.Sleep(period * time.Second)
	}
}
