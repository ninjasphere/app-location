package ui

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/ninjasphere/app-location/calibration"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/suit"
)

var log = logger.GetLogger("ui")

var locations = []*calibration.Location{
	&calibration.Location{
		ID:      "1",
		Name:    "Kitchen",
		Quality: 0,
	},
	&calibration.Location{
		ID:      "2",
		Name:    "Bathroom",
		Quality: 30,
	},
	&calibration.Location{
		ID:      "3",
		Name:    "Lounge Room",
		Quality: 100,
	},
	&calibration.Location{
		ID:      "4",
		Name:    "Master Bedroom",
		Quality: 60,
	},
}

type CalibrationUI struct {
	calibrationService *calibration.Service
}

func NewUI(service *calibration.Service) *CalibrationUI {
	return &CalibrationUI{service}
}

func (c *CalibrationUI) GetActions(request *model.ConfigurationRequest) (*[]suit.ReplyAction, error) {
	return &[]suit.ReplyAction{
		suit.ReplyAction{
			Name:        "listLocations",
			Label:       "Calibrate a Room",
			DisplayIcon: "location-arrow",
		},
	}, nil
}

func (c *CalibrationUI) Configure(request *model.ConfigurationRequest) (*suit.ConfigurationScreen, error) {
	log.Infof("Incoming configuration request. Action:%s Data:%s", request.Action, string(request.Data))

	var vals map[string]string
	json.Unmarshal(request.Data, &vals)

	switch request.Action {
	case "listLocations":
		return c.listLocations()
	case "clear":
		location, ok := vals["location"]
		if !ok {
			return c.error("No location to clear", "listLocations")
		}
		return c.clear(location)

	case "calibrate":
		_, ok := vals["location"]
		if !ok {
			return c.error("No location to calibrate", "listLocations")
		}
		return c.calibrate(vals)
	case "status":
		calibrationID, ok := vals["calibration"]
		if !ok {
			return c.error("No calibration to get status of", "listLocations")
		}
		return c.status(calibrationID)
	case "clearAll":
		return c.clearAll()
	default:
		return c.error(fmt.Sprintf("Unknown action: '%s'", request.Action), "listLocations")
	}
}

func (c *CalibrationUI) clear(locationID string) (*suit.ConfigurationScreen, error) {
	log.Infof("Clearing calibration for location: %s", locationID)

	for _, location := range locations {
		if location.ID == locationID {
			location.Quality = 0
		}
	}
	return c.listLocations()
}

func (c *CalibrationUI) clearAll() (*suit.ConfigurationScreen, error) {
	log.Infof("Clearing all calibration")

	for _, location := range locations {

		location.Quality = 0
	}

	return c.listLocations()
}

func (c *CalibrationUI) listLocations() (*suit.ConfigurationScreen, error) {

	var items []suit.ActionListOption

	for _, location := range locations {

		quality := ""

		if location.Quality > 85 {
			quality = "excellent"
		} else if location.Quality > 85 {
			quality = "good"
		} else if location.Quality > 85 {
			quality = "ok"
		} else if location.Quality > 85 {
			quality = "poor"
		} else if location.Quality > 1 {
			quality = "terrible"
		}

		items = append(items, suit.ActionListOption{
			Title:    location.Name,
			Subtitle: quality,
			Value:    location.ID,
		})
	}

	screen := suit.ConfigurationScreen{
		Title:       "Calibrate a Room",
		DisplayIcon: "location-arrow",
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.Alert{
						Title: "Which room would you like to calibrate?",
					},

					suit.ActionList{
						Name:    "location",
						Options: items,
						PrimaryAction: &suit.ReplyAction{
							Name:        "calibrate",
							DisplayIcon: "map-marker",
						},
						SecondaryAction: &suit.ReplyAction{
							Name:         "clear",
							Label:        "Reset",
							DisplayIcon:  "remove",
							DisplayClass: "danger",
						},
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.AutomaticAction{
				Name:  "listLocations",
				Delay: 3000,
			},
			suit.ReplyAction{
				Name:         "clearAll",
				Label:        "Reset All",
				DisplayIcon:  "trash",
				DisplayClass: "danger",
			},
			suit.CloseAction{
				Label: "Close",
			},
		},
	}

	return &screen, nil
}

func (c *CalibrationUI) calibrate(data map[string]string) (*suit.ConfigurationScreen, error) {

	var location *calibration.Location

	for _, l := range locations {
		if l.ID == data["location"] {
			location = l
		}
	}

	startTimeString, ok := data["startTime"]
	if !ok {
		startTimeString = time.Now().Format(time.RFC3339)
	}

	startTime, _ := time.Parse(time.RFC3339, startTimeString)

	if time.Since(startTime) > time.Second*10 {
		// Just pretend they placed it for now.
		return c.startCalibration(location, "my-fake-tag")
	}

	log.Infof("Calibrating location: %s", location.Name)

	screen := suit.ConfigurationScreen{
		Title:       "Calibrating " + location.Name,
		DisplayIcon: "bar-chart",
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.InputHidden{
						Name:  "location",
						Value: location.ID,
					},
					suit.InputHidden{
						Name:  "startTime",
						Value: startTimeString,
					},
					suit.Alert{
						Title: "Please place a Ninja tag on top of the Sphere.",
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.AutomaticAction{
				Name:  "calibrate",
				Delay: 1000,
			},
			suit.ReplyAction{
				Name:  "listLocations",
				Label: "Cancel",
			},
		},
	}

	if time.Since(startTime) > time.Second*5 {
		screen.Sections[0].Contents = append(screen.Sections[0].Contents, suit.Alert{
			Title:        "Having trouble?",
			Subtitle:     "Make sure you've placed a tag on top of the master sphere.",
			DisplayClass: "warning",
		})
	}

	return &screen, nil
}

func (c *CalibrationUI) startCalibration(location *calibration.Location, tagID string) (*suit.ConfigurationScreen, error) {

	log.Infof("Starting calibration of location %s using tag %s", location.Name, tagID)

	calibrationID := "my-calibration"

	screen := suit.ConfigurationScreen{
		Title:       "Starting calibration...",
		DisplayIcon: "bar-chart",
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.Alert{
						Title: fmt.Sprintf("Starting calibration of location %s using tag %s", location.Name, tagID),
					},
					suit.InputHidden{
						Name:  "calibration",
						Value: calibrationID,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.AutomaticAction{
				Name:  "status",
				Delay: 5000,
			},
		},
	}

	return &screen, nil
}

var progress = map[string]int{}

func (c *CalibrationUI) status(calibrationID string) (*suit.ConfigurationScreen, error) {

	log.Infof("Getting status of calibration: %s", calibrationID)

	if _, ok := progress[calibrationID]; !ok {
		progress[calibrationID] = 0
	}

	progress[calibrationID] += rand.Intn(15)

	if progress[calibrationID] >= 100 {
		return &suit.ConfigurationScreen{
			Title:       "Calibration Complete",
			DisplayIcon: "gear",
			Sections: []suit.Section{
				suit.Section{
					Contents: []suit.Typed{
						suit.Alert{
							Title: fmt.Sprintf("Saving calibration..."),
						},
					},
				},
			},
			Actions: []suit.Typed{
				suit.AutomaticAction{
					Name:  "listLocations",
					Delay: 2000,
				},
			},
		}, nil
	}

	screen := suit.ConfigurationScreen{
		Title:       "Calibrating",
		DisplayIcon: "gear fa-spin",
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.Alert{
						Title: fmt.Sprintf("%d%% complete...", progress[calibrationID]),
					},
					suit.InputHidden{
						Name:  "calibration",
						Value: calibrationID,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.AutomaticAction{
				Name:  "status",
				Delay: 500,
			},
		},
	}

	return &screen, nil
}

func (c *CalibrationUI) error(message, nextAction string) (*suit.ConfigurationScreen, error) {
	return &suit.ConfigurationScreen{
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.Alert{
						Title:        "Error",
						Subtitle:     message,
						DisplayClass: "danger",
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label: "Continue",
				Name:  nextAction,
			},
		},
	}, nil
}
