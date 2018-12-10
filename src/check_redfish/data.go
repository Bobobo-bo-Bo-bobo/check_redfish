package main

// Nagios/Icinga exit codes
const (
	NAGIOS_OK int = iota
	NAGIOS_WARNING
	NAGIOS_CRITICAL
	NAGIOS_UNKNOWN
)

type NagiosState struct {
	Critical []string
	Warning  []string
	Ok       []string
	Unknown  []string
	PerfData []string
}
