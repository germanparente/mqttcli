package lib

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

func InfluxWriteFloat(stat string, label string, value float64) {

	p := influxdb2.NewPoint(stat,
		map[string]string{"unit": label},
		map[string]interface{}{"value": value},
		time.Now())
	TheInfluxWrite(p)
}

func InfluxWriteInt(stat string, label string, value int) {

	p := influxdb2.NewPoint(stat,
		map[string]string{"unit": label},
		map[string]interface{}{"value": value},
		time.Now())
	TheInfluxWrite(p)
}

func InfluxWriteString(stat string, label string, value string) {

	p := influxdb2.NewPoint(stat,
		map[string]string{"unit": label},
		map[string]interface{}{"value": value},
		time.Now())
	TheInfluxWrite(p)
}

func TheInfluxWrite(point *write.Point) {
	client := influxdb2.NewClient(Myconfig.InfluxDB.Url, Myconfig.InfluxDB.Token)
	writeAPI := client.WriteAPIBlocking(Myconfig.InfluxDB.Org, Myconfig.InfluxDB.Bucket)
	// write point asynchronously
	err := writeAPI.WritePoint(context.Background(), point)
	fmt.Println("influx error ", err)
	// always close client at the end
	defer client.Close()
}
