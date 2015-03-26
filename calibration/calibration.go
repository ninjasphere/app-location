package calibration

import (
	"time"

	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/go-ninja/logger"
)

var log = logger.GetLogger("calibration")

type Service struct {
	conn               *ninja.Connection
	roomModel          *ninja.ServiceClient
	calibrationService *ninja.ServiceClient
	ConnectedWaypoints int
}

type Status struct {
	Progress int
}

func NewService(conn *ninja.Connection) *Service {

	service := &Service{
		conn:               conn,
		roomModel:          conn.GetServiceClient("$home/services/RoomModel"),
		calibrationService: conn.GetServiceClient("$i/dont/know/what/this/is"),
	}

	conn.SubscribeRaw("$location/waypoints", func(waypoints *[]int, values map[string]string) bool {
		if waypoints != nil && len(*waypoints) != 0 {
			service.ConnectedWaypoints = (*waypoints)[0]
		}
		return true
	})

	return service

}

func (s *Service) getCalibrationScores() (scores map[string]float64, err error) {
	err = s.calibrationService.Call("fetch", config.MustString("user"), &scores, time.Second*15)

	return
}

type RSSI struct {
	Device   string `json:"device"`
	Waypoint string `json:"waypoint"`
	RSSI     int    `json:"rssi"`
	IsSphere bool   `json:"isSphere"`
}

func (s *Service) GetCalibrationDevice(minimumRssi int, timeout time.Duration) string {

	done := false
	device := make(chan string, 1)
	s.conn.Subscribe("$device/:deviceId/:channel/rssi", func(rssi *RSSI, values map[string]string) bool {

		if done {
			return false
		}

		if rssi.RSSI >= minimumRssi && rssi.IsSphere {
			device <- rssi.Device
			return false
		}

		return true
	})

	select {
	case deviceID := <-device:
		log.Infof("Found calibration device:", deviceID)
		return deviceID
	case <-time.After(timeout):
		done = true
		return ""
	}
}

type Location struct {
	ID      string
	Name    string
	Quality int
}
