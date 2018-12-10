package main

import (
	"fmt"
	"strings"
)

func ProcessStatus(n NagiosState) (int, string) {
	u_str := strings.Join(n.Unknown, ", ")
	c_str := strings.Join(n.Critical, ", ")
	w_str := strings.Join(n.Warning, ", ")
	o_str := strings.Join(n.Ok, ", ")

	if len(n.Unknown) > 0 {
		return NAGIOS_UNKNOWN, fmt.Sprintf("%s; %s; %s; %s", u_str, c_str, w_str, o_str)
	} else if len(n.Critical) > 0 {
		return NAGIOS_CRITICAL, fmt.Sprintf("%s; %s; %s", c_str, w_str, o_str)
	} else if len(n.Warning) > 0 {
		return NAGIOS_WARNING, fmt.Sprintf("%s; %s", w_str, o_str)
	} else if len(n.Ok) > 0 {
		return NAGIOS_OK, fmt.Sprintf("%s", o_str)
	} else {
		// shouldn't happen
		return NAGIOS_UNKNOWN, "No results at all found"
	}
}
