package main

import (
	"errors"
	"fmt"
	redfish "git.ypbind.de/repository/go-redfish.git"
	"strings"
)

func CheckGeneralHealth(rf redfish.Redfish, sys_id string) (NagiosState, error) {
	var system_data *redfish.SystemData
	var found bool
	var state = NagiosState{
		Critical: make([]string, 0),
		Warning:  make([]string, 0),
		Ok:       make([]string, 0),
		Unknown:  make([]string, 0),
	}

	// no system specified, pick the first reported system
	if sys_id == "" {
		sys_epl, err := rf.GetSystems()
		if err != nil {
			return state, err
		}

		// should never happen
		if len(sys_epl) == 0 {
			state.Unknown = append(state.Unknown, "No system endpoint reported at all")
			return state, errors.New("BUG: No system endpoint reported at all")
		}

		system_data, err = rf.GetSystemData(sys_epl[0])
		if err != nil {
			state.Unknown = append(state.Unknown, err.Error())
			return state, err
		}
	} else {
		sys_data_map, err := rf.MapSystemsById()
		if err != nil {
			state.Unknown = append(state.Unknown, err.Error())
			return state, err
		}

		system_data, found = sys_data_map[sys_id]
		if !found {
			state.Unknown = append(state.Unknown, fmt.Sprintf("System with ID %s not found", sys_id))
			return state, errors.New(fmt.Sprintf("System with ID %s not found", sys_id))
		}
	}

	if system_data.Status.State == nil || *system_data.Status.State == "" {
		state.Unknown = append(state.Unknown, fmt.Sprintf("System with ID %s do not define a state", sys_id))
		return state, errors.New(fmt.Sprintf("System with ID %s do not define a state", sys_id))
	}

	// Although the strings are defined in the specification, some implementations return all uppercase or all
	// lower case (see https://redfish.dmtf.org/schemas/v1/Resource.json#/definitions/Status)
	l_state := strings.ToLower(*system_data.Status.State)
	if l_state == "enabled" {
		if system_data.Status.Health == nil || *system_data.Status.Health == "" {
			state.Unknown = append(state.Unknown, fmt.Sprintf("System with ID %s do not define a health state", sys_id))
			return state, errors.New(fmt.Sprintf("System with ID %s do not define a health state", sys_id))
		}

		l_health := strings.ToLower(*system_data.Status.Health)
		if l_health == "ok" {
			state.Ok = append(state.Ok, "General health is reported as OK")
		} else if l_health == "warning" {
			state.Warning = append(state.Warning, "General health is reported as warning")
		} else if l_health == "Critical" {
			state.Critical = append(state.Critical, "General health is reported as critical")
		} else if l_health == "Failed" {
			// XXX: Although https://redfish.dmtf.org/schemas/v1/Resource.json#/definitions/Status only defines Ok, Warning and Critical some boards report
			//      Failed as well
			state.Critical = append(state.Critical, "General health is reported as failed")
		} else {
			// XXX: there may be more non-standard health strings
			state.Unknown = append(state.Unknown, "General health is reported as \"%s\"", *system_data.Status.Health)
		}
	} else {
		state.Unknown = append(state.Unknown, fmt.Sprintf("System reports \"%s\" instead of \"Enabled\" for health information", *system_data.Status.State))
		return state, errors.New(fmt.Sprintf("System reports \"%s\" instead of \"Enabled\" for health information", *system_data.Status.State))
	}

	return state, nil
}
