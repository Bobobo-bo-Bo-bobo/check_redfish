package main

import (
	"errors"
	"fmt"
	redfish "git.ypbind.de/repository/go-redfish.git"
	"strconv"
	"strings"
)

func CheckVoltages(rf redfish.Redfish, cha_id string) (NagiosState, error) {
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

	// check for "Power" endpoint
	if chassis_data.Power == nil {
		state.Unknown = append(state.Unknown, fmt.Sprintf("No Power endpoint defined for chassis with ID %s", cha_id))
		return state, errors.New(fmt.Sprintf("No Power endpoint defined for chassis with ID %s", cha_id))
	}

	if chassis_data.Power.Id == nil || *chassis_data.Power.Id == "" {
		state.Unknown = append(state.Unknown, fmt.Sprintf("Power endpoint of chassis with ID %s has no Id attribute", cha_id))
		return state, errors.New(fmt.Sprintf("Power endpoint of chassis with ID %s has no Id attribute", cha_id))
	}

	p, err := rf.GetPowerData(*chassis_data.Power.Id)
	if err != nil {
		state.Unknown = append(state.Unknown, err.Error())
		return state, err
	}

	for _, vlt := range p.Voltages {
		vlt_name := "<unnamed voltage>"
		if vlt.Name != nil {
			vlt_name = *vlt.Name
		}

		if vlt.Status.State == nil || *vlt.Status.State == "" {
			continue
		}

		if vlt.Status.Health == nil || *vlt.Status.Health == "" {
			continue
		}

		vlt_health := strings.ToLower(*vlt.Status.Health)

		vlt_state := strings.ToLower(*vlt.Status.State)
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
			            But Voltages.State is only Enabled/Disabled/Absent
		*/
		if vlt_state == "enabled" {
			if vlt_health == "critical" {
				state.Critical = append(state.Critical, fmt.Sprintf("Voltage %s is reported as critical", vlt_name))
			} else if vlt_health == "warning" {
				state.Warning = append(state.Warning, fmt.Sprintf("Voltage %s is reported as warning", vlt_name))
			} else if vlt_health == "ok" {
				state.Ok = append(state.Ok, fmt.Sprintf("Voltage %s is reported as ok", vlt_name))
			}

			if vlt.ReadingVolts != nil && *vlt.ReadingVolts > 0.0 {
				// XXX: Is there a way to handle upper/lower warning/critical thresholds (LowerThresholdNonCritical, LowerThresholdCritical
				//      UpperThresholdNonCritical and UpperThresholdCritical) ?

				_cur := strconv.FormatFloat(*vlt.ReadingVolts, 'f', -1, 32)
				_min := ""
				_max := ""

				if vlt.MinReadingRange != nil && *vlt.MinReadingRange > 0.0 {
					_min = strconv.FormatFloat(*vlt.MinReadingRange, 'f', -1, 32)
				}

				if vlt.MaxReadingRange != nil && *vlt.MaxReadingRange > 0.0 {
					_max = strconv.FormatFloat(*vlt.MaxReadingRange, 'f', -1, 32)
				}
				label := fmt.Sprintf("voltage_%s", vlt_name)
				perfdata, err := MakePerfDataString(label, _cur, nil, nil, nil, &_min, &_max)
				if err == nil {
					state.PerfData = append(state.PerfData, perfdata)
				}
			}
		}
	}
	return state, nil
}
