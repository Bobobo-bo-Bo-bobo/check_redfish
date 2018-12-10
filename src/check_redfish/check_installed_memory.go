package main

import (
	"errors"
	"fmt"
	redfish "git.ypbind.de/repository/go-redfish.git"
)

func CheckInstalledMemory(rf redfish.Redfish, sys_id string, m int) (NagiosState, error) {
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

	if system_data.MemorySummary == nil {
		state.Unknown = append(state.Unknown, fmt.Sprintf("No MemorySummary data found for system with ID %s", sys_id))
		return state, errors.New(fmt.Sprintf("No MemorySummary data found for system with ID %s", sys_id))
	}

	if system_data.MemorySummary.TotalSystemMemoryGiB == 0 {
		state.Unknown = append(state.Unknown, fmt.Sprintf("No TotalSystemMemoryGiB reported in MemorySummary data for system with ID %s", sys_id))
		return state, errors.New(fmt.Sprintf("No TotalSystemMemoryGiB reported in MemorySummary data for system with ID %s", sys_id))
	}

	if system_data.MemorySummary.TotalSystemMemoryGiB < m {
		state.Critical = append(state.Critical, fmt.Sprintf("Only %d GiB of memory (instead of %d) installed", system_data.MemorySummary.TotalSystemMemoryGiB, m))
		return state, nil
	}

	if system_data.MemorySummary.TotalSystemMemoryGiB > m {
		state.Warning = append(state.Warning, fmt.Sprintf("%d GiB of memory (instead of %d) installed", system_data.MemorySummary.TotalSystemMemoryGiB, m))
		return state, nil
	}

	state.Ok = append(state.Ok, fmt.Sprintf("%d GiB of memory installed", system_data.MemorySummary.TotalSystemMemoryGiB))
	return state, nil
}
