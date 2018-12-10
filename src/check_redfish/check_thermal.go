package main

import (
	"errors"
	"fmt"
	redfish "git.ypbind.de/repository/go-redfish.git"
	"strconv"
	"strings"
)

func CheckThermal(rf redfish.Redfish, cha_id string) (NagiosState, error) {
	var chassis_data *redfish.ChassisData
	var found bool
	var state = NagiosState{
		Critical: make([]string, 0),
		Warning:  make([]string, 0),
		Ok:       make([]string, 0),
		Unknown:  make([]string, 0),
		PerfData: make([]string, 0),
	}

	// no system specified, pick the first reported system
	if cha_id == "" {
		cha_epl, err := rf.GetChassis()
		if err != nil {
			return state, err
		}

		// should never happen
		if len(cha_epl) == 0 {
			state.Unknown = append(state.Unknown, "No chassis  endpoint reported at all")
			return state, errors.New("BUG: No chassis endpoint reported at all")
		}

		chassis_data, err = rf.GetChassisData(cha_epl[0])
		if err != nil {
			state.Unknown = append(state.Unknown, err.Error())
			return state, err
		}
	} else {
		cha_data_map, err := rf.MapChassisById()
		if err != nil {
			state.Unknown = append(state.Unknown, err.Error())
			return state, err
		}

		chassis_data, found = cha_data_map[cha_id]
		if !found {
			state.Unknown = append(state.Unknown, fmt.Sprintf("Chassis with ID %s not found", cha_id))
			return state, errors.New(fmt.Sprintf("Chassis with ID %s not found", cha_id))
		}
	}

	// check for "Thermal" endpoint
	if chassis_data.Thermal == nil {
		state.Unknown = append(state.Unknown, fmt.Sprintf("No Thermal endpoint defined for chassis with ID %s", cha_id))
		return state, errors.New(fmt.Sprintf("No Thermal endpoint defined for chassis with ID %s", cha_id))
	}

	if chassis_data.Thermal.Id == nil || *chassis_data.Thermal.Id == "" {
		state.Unknown = append(state.Unknown, fmt.Sprintf("Thermal endpoint of chassis with ID %s has no Id attribute", cha_id))
		return state, errors.New(fmt.Sprintf("Thermal endpoint of chassis with ID %s has no Id attribute", cha_id))
	}

	t, err := rf.GetThermalData(*chassis_data.Thermal.Id)
	if err != nil {
		state.Unknown = append(state.Unknown, err.Error())
		return state, err
	}

	for _, tmp := range t.Temperatures {
		tmp_name := "<unnamed sensor>"
		if tmp.Name != nil {
			tmp_name = *tmp.Name
		}

		if tmp.Status.State == nil || *tmp.Status.State == "" {
			continue
		}

		if tmp.Status.Health == nil || *tmp.Status.Health == "" {
			continue
		}

		tmp_health := strings.ToLower(*tmp.Status.Health)

		tmp_state := strings.ToLower(*tmp.Status.State)
		/*
						According to the specification at https://redfish.dmtf.org/schemas/v1/Resource.json#/definitions/State possible
						values are:

							"Enabled": "This function or resource has been enabled.",
							"Disabled": "This function or resource has been disabled.",
							"StandbyOffline": "This function or resource is enabled, but a waiting an external action to activate it.",
							"StandbySpare": "This function or resource is part of a redundancy set and is awaiting a failover or
											 other external action to activate it.",
							"InTest": "This function or resource is undergoing testing.",
							"Starting": "This function or resource is starting.",
							"Absent": "This function or resource is not present or not detected.",
							"UnavailableOffline": "This function or resource is present but cannot be used.",
							"Deferring": "The element will not process any commands but will queue new requests.",
							"Quiesced": "The element is enabled but only processes a restricted set of commands.",
							"Updating": "The element is updating and may be unavailable or degraded."

			           Temperatures.Status seem to only to use Enabled/Disabled as state
		*/
		if tmp_state == "enabled" {
			if tmp_health == "critical" {
				state.Critical = append(state.Critical, fmt.Sprintf("Sensor \"%s\" is reported as critical", tmp_name))
			} else if tmp_health == "warning" {
				state.Warning = append(state.Warning, fmt.Sprintf("Sensor \"%s\" is reported as warning", tmp_name))
			} else if tmp_health == "ok" {
				state.Ok = append(state.Ok, fmt.Sprintf("Sensor \"%s\" is reported as ok", tmp_name))
			}
			if tmp.ReadingCelsius != nil && *tmp.ReadingCelsius > 0 {
				_cur := strconv.FormatInt(int64(*tmp.ReadingCelsius), 10)
				_min := ""
				_max := ""
				_wrn := ""
				_crt := ""

				if tmp.MinReadingRangeTemp != nil && *tmp.MinReadingRangeTemp > 0 {
					_min = strconv.FormatInt(int64(*tmp.MinReadingRangeTemp), 10)
				}

				if tmp.MaxReadingRangeTemp != nil && *tmp.MaxReadingRangeTemp > 0 {
					_max = strconv.FormatInt(int64(*tmp.MaxReadingRangeTemp), 10)
				}

				// we report _upper_ ranges but _lower_ ranges also exist
				if tmp.UpperThresholdNonCritical != nil && *tmp.UpperThresholdNonCritical > 0 {
					_min = strconv.FormatInt(int64(*tmp.UpperThresholdNonCritical), 10)
				}

				if tmp.UpperThresholdCritical != nil && *tmp.UpperThresholdCritical > 0 {
					_max = strconv.FormatInt(int64(*tmp.UpperThresholdCritical), 10)
				}

				label := fmt.Sprintf("temperature_%s", tmp_name)
				perfdata, err := MakePerfDataString(label, _cur, nil, &_wrn, &_crt, &_min, &_max)
				if err == nil {
					state.PerfData = append(state.PerfData, perfdata)
				}
			}

		}
	}

	return state, nil
}
