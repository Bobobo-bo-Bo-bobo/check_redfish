package main

import (
	"errors"
	"fmt"
	redfish "git.ypbind.de/repository/go-redfish.git"
	"strconv"
	"strings"
)

func CheckPsu(rf redfish.Redfish, cha_id string, warn int, crit int) (NagiosState, error) {
	var chassis_data *redfish.ChassisData
	var found bool
	var state = NagiosState{
		Critical: make([]string, 0),
		Warning:  make([]string, 0),
		Ok:       make([]string, 0),
		Unknown:  make([]string, 0),
		PerfData: make([]string, 0),
	}
	var psu_count int
	var working_psu_count int

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

	for _, psu := range p.PowerSupplies {
		psu_name := "<unnamed PSU>"
		if psu.Name != nil {
			psu_name = *psu.Name
		}

		psu_sn := "<no serial number reported>"
		if psu.SerialNumber != nil {
			psu_sn = *psu.SerialNumber
		}

		if psu.Status.State == nil || *psu.Status.State == "" {
			continue
		}

		if psu.Status.Health == nil || *psu.Status.Health == "" {
			continue
		}

		psu_health := strings.ToLower(*psu.Status.Health)

		psu_state := strings.ToLower(*psu.Status.State)
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
		*/
		if psu_state == "enabled" || psu_state == "standbyspare" || psu_state == "quiesced" {
			psu_count += 1
			if psu_health == "critical" {
				state.Critical = append(state.Critical, fmt.Sprintf("PSU %s (SN: %s) is reported as critical", psu_name, psu_sn))
			} else if psu_health == "warning" {
				state.Warning = append(state.Warning, fmt.Sprintf("PSU %s (SN: %s) is reported as warning", psu_name, psu_sn))
			} else if psu_health == "ok" {
				state.Ok = append(state.Ok, fmt.Sprintf("PSU %s (SN: %s) is reported as ok", psu_name, psu_sn))
				working_psu_count += 1
			}
		} else if psu_state == "standbyoffline" || psu_state == "intest" || psu_state == "disabled" || psu_state == "starting" || psu_state == "deferring" || psu_state == "updating" {
			psu_count += 1
			if psu_health == "critical" {
				state.Critical = append(state.Critical, fmt.Sprintf("PSU %s (SN: %s) is reported as critical", psu_name, psu_sn))
			} else if psu_health == "warning" {
				state.Warning = append(state.Warning, fmt.Sprintf("PSU %s (SN: %s) is reported as warning", psu_name, psu_sn))
			} else if psu_health == "ok" {
				state.Ok = append(state.Ok, fmt.Sprintf("PSU %s (SN: %s) is reported as ok", psu_name, psu_sn))
				working_psu_count += 1
			}
		} else if psu_state == "unavailableoffline" {
			psu_count += 1
			if psu_health == "critical" {
				state.Critical = append(state.Critical, fmt.Sprintf("PSU %s (SN: %s) is reported as critical", psu_name, psu_sn))
			} else if psu_health == "warning" {
				state.Warning = append(state.Warning, fmt.Sprintf("PSU %s (SN: %s) is reported as warning", psu_name, psu_sn))
			} else if psu_health == "ok" {
				state.Ok = append(state.Ok, fmt.Sprintf("PSU %s (SN: %s) is reported as ok", psu_name, psu_sn))
				working_psu_count += 1
			}
		} else if psu_state == "absent" {
			// nothing to do here
			continue
		} else {
			state.Unknown = append(state.Unknown, fmt.Sprintf("PSU %s (SN: %s) reports unknown state %s", psu_name, psu_sn, *psu.Status.State))
			continue
		}

		// add power performance data
		if psu.LastPowerOutputWatts != nil && *psu.LastPowerOutputWatts > 0 {
			_cur := strconv.FormatInt(int64(*psu.LastPowerOutputWatts), 10)

			// get limits
			_max := ""
			if psu.PowerCapacityWatts != nil && *psu.PowerCapacityWatts > 0 {
				_max = strconv.FormatInt(int64(*psu.PowerCapacityWatts), 10)
			}
			label := fmt.Sprintf("power_output_%s", psu_name)
			perfdata, err := MakePerfDataString(label, _cur, nil, nil, nil, nil, &_max)
			if err == nil {
				state.PerfData = append(state.PerfData, perfdata)
			}
		}

		if psu.LineInputVoltage != nil && *psu.LineInputVoltage > 0 {
			_cur := strconv.FormatInt(int64(*psu.LineInputVoltage), 10)
			label := fmt.Sprintf("input_voltage_%s", psu_name)
			perfdata, err := MakePerfDataString(label, _cur, nil, nil, nil, nil, nil)
			if err == nil {
				state.PerfData = append(state.PerfData, perfdata)
			}
		}
	}

	// some b0rken implementations of the power controller firmware may report no PSUs at all
	if working_psu_count == 0 {
		state.Unknown = append(state.Unknown, fmt.Sprintf("No working power supplies reported. This may be a firmware bug!"))
		return state, errors.New(fmt.Sprintf("No working power supplies reported. This may be a firmware bug!"))
	}

	if psu_count == 0 {
		state.Unknown = append(state.Unknown, fmt.Sprintf("No power supplies reported at all. This may be a firmware bug!"))
		return state, errors.New(fmt.Sprintf("No power supplies reported at all. This may be a firmware bug!"))
	}

	if working_psu_count < crit {
		// prepend number of working PSUs to message slice
		_state := state.Critical
		state.Critical = make([]string, 1)
		state.Critical[0] = fmt.Sprintf("Only %d out of %d power supplies are working", working_psu_count, psu_count)
		state.Critical = append(state.Critical, _state...)
	} else if working_psu_count < warn {
		// prepend number of working PSUs to message slice
		_state := state.Warning
		state.Warning = make([]string, 1)
		state.Warning[0] = fmt.Sprintf("Only %d out of %d power supplies are working", working_psu_count, psu_count)
		state.Warning = append(state.Warning, _state...)
	} else {
		// prepend number of working PSUs to message slice
		_state := state.Ok
		state.Ok = make([]string, 1)
		state.Ok[0] = fmt.Sprintf("%d out of %d power supplies are working", working_psu_count, psu_count)
		state.Ok = append(state.Ok, _state...)
	}

	return state, nil
}
