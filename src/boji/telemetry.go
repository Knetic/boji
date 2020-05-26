package boji

import (
	"context"
	"time"
	"fmt"
	"github.com/influxdata/influxdb-client-go"
)

type telemetry struct {
	client influxdb2.Client
	stats telemetryStats
	
	Bucket string
}

type telemetryStats struct {
	filesCreated int
	filesOpened int
	filesDeleted int
	filesStatted int

	directoriesCreated int
	bytesWritten int64
	bytesRead int64
	failedAuths int
}

func newTelemetry(url string, bucket string) *telemetry {

	if url == "" {
		return nil
	}

	client := influxdb2.NewClient(url, "")
	return &telemetry{
		client: client,
		Bucket: bucket,
	}
}

func (this *telemetry) publish() error {

	snapshot := this.stats
	this.stats = telemetryStats{}

	if this.client == nil {
		return nil
	}
	
	point := influxdb2.NewPoint(
		"boji",
		map[string]string{},
		map[string]interface{}{
			"filesCreated": snapshot.filesCreated,
			"fileOpened": snapshot.filesOpened,
			"filesDeleted": snapshot.filesDeleted,
			"fileStatted": snapshot.filesStatted,
			"directoriesCreated": snapshot.directoriesCreated,
			"bytesWritten": snapshot.bytesWritten,
			"bytesRead": snapshot.bytesRead,
			"failedAuths": snapshot.failedAuths,
		},
		time.Now(),
	)

	writeApi := this.client.WriteApiBlocking("boji", this.Bucket)
	fmt.Printf("Publishing point\n") // TODO: remove after debug
	err := writeApi.WritePoint(context.Background(), point)
	return err
}