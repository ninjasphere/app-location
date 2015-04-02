package calibration

import (
	"time"

	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/go-ninja/logger"
)

var log = logger.GetLogger("calibration")

type Service struct {
	conn                        *ninja.Connection
	roomModel                   *ninja.ServiceClient
	calibrationService          *ninja.ServiceClient
	calibrationFlushUserService *ninja.ServiceClient
	calibrationFlushZoneService *ninja.ServiceClient
	ConnectedWaypoints          int
}

type Status struct {
	Progress int
}

func NewService(conn *ninja.Connection) *Service {

	service := &Service{
		conn:                        conn,
		roomModel:                   conn.GetServiceClient("$home/services/RoomModel"),
		calibrationService:          conn.GetServiceClient("$ninja/services/rpc/Location/calibrationScore"),
		calibrationFlushUserService: conn.GetServiceClient("$ninja/services/rpc/Location/flushuser"),
		calibrationFlushZoneService: conn.GetServiceClient("$ninja/services/rpc/Location/flushzone"),
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
	err = s.calibrationService.Call("Location.calibrationScore", []interface{}{config.MustString("userId")}, &scores, time.Second*15)

	return
}

func (s *Service) doFlushUser() (result map[string]interface{}, err error) {
	log.Debugf("location.flushuser")
	err = s.calibrationFlushUserService.Call("Location.flushuser", []interface{}{}, &result, time.Second*15)

	return
}

func (s *Service) doFlushZone(zone string) (result map[string]interface{}, err error) {
	log.Debugf("location.flushzone zone=%s", zone)
	err = s.calibrationFlushZoneService.Call("Location.flushzone", []interface{}{zone}, &result, time.Second*15)

	return
}

type RSSI struct {
	Device   string `json:"device"`
	Waypoint string `json:"waypoint"`
	RSSI     int    `json:"rssi"`
	IsSphere bool   `json:"isSphere"`
}

// ClearAll clear all calibration for this user.
func (s *Service) ClearAll() {
	res, err := s.doFlushUser()
	if err != nil {
		log.Errorf("failed to clear all err: %s", err)
	}
	log.Infof("doFlushUser: %v", res)
}

// ClearAll clear all calibration for this user.
func (s *Service) ClearLocation(locationID string) {
	res, err := s.doFlushZone(locationID)
	if err != nil {
		log.Errorf("failed to clear all err: %s", err)
	}
	log.Infof("doFlushZone: %v", res)
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
