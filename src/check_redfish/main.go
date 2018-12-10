package main

import (
	"flag"
	"fmt"
	redfish "git.ypbind.de/repository/go-redfish.git"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	var user = flag.String("user", "", "Username for authentication")
	var password = flag.String("password", "", "Password for authentication")
	var password_file = flag.String("password-file", "", "Read password for authentication from file, password is exepected in the first line of the file")
	var host = flag.String("host", "", "Hostname or IP address of management board")
	var port = flag.Int("port", 0, "Port to connect to")
	var insecure_ssl = flag.Bool("insecure-ssl", false, "Don't verifiy SSL certificate")
	var chassis_id = flag.String("chassis-id", "", "Process data of specific chassis")
	var system_id = flag.String("system-id", "", "Process data of specific system")
	var check_installed_memory = flag.String("check-installed-memory", "", "Check installed memory")
	var check_installed_cpus = flag.String("check-installed-cpus", "", "Check installed CPUs")
	var check_thermal = flag.Bool("check-thermal", false, "Check thermal status")
	var check_psu = flag.String("check-psu", "", "Check PSU")
	var check_general = flag.Bool("check-general-health", true, "Check general health")
	var timeout = flag.Uint("timeout", 60, "Connection timeout in seconds")
	var help = flag.Bool("help", false, "Show help")
	var status NagiosState

	flag.Usage = ShowUsage
	flag.Parse()

	if *help {
		ShowUsage()
		os.Exit(0)
	}

	if *host == "" {
		fmt.Fprintf(os.Stderr, "ERROR: No hostname specified\n")
		ShowUsage()
		os.Exit(NAGIOS_UNKNOWN)
	}

	if *user == "" {
		fmt.Fprintf(os.Stderr, "ERROR: No username specified\n")
		ShowUsage()
		os.Exit(NAGIOS_UNKNOWN)
	}

	if *password == "" && *password_file == "" {
		fmt.Fprintf(os.Stderr, "ERROR: No password or password file specified\n")
		ShowUsage()
		os.Exit(NAGIOS_UNKNOWN)
	}

	if *port < 0 || *port > 65535 {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid port number %d\n", *port)
		os.Exit(NAGIOS_UNKNOWN)
	}

	if *password_file != "" {
		passwd, err := ReadSingleLine(*password_file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s", err.Error())
			os.Exit(NAGIOS_UNKNOWN)
		}
		password = &passwd
	}

	rf := redfish.Redfish{
		Hostname:    *host,
		Port:        *port,
		Username:    *user,
		Password:    *password,
		InsecureSSL: *insecure_ssl,
		Timeout:     time.Duration(*timeout) * time.Second,
	}

	// setup session
	err := rf.Initialise()
	if err != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("ERROR: Initialisation failed for %s: %s\n", rf.Hostname, err.Error()))
		os.Exit(NAGIOS_UNKNOWN)
	}

	err = rf.Login()
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Sprintf("ERROR: Login on %s failed for user %s: %s\n", rf.Hostname, rf.Username, err.Error()))
		os.Exit(NAGIOS_UNKNOWN)
	}

	defer rf.Logout()

	if *check_installed_memory != "" {
		m, err := strconv.Atoi(*check_installed_memory)
		if err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("ERROR: Can't convert %s to a number: %s\n", *check_installed_memory, err.Error()))
			os.Exit(NAGIOS_UNKNOWN)
		}

		status, _ = CheckInstalledMemory(rf, *system_id, m)
	} else if *check_installed_cpus != "" {
		c, err := strconv.Atoi(*check_installed_cpus)
		if err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("ERROR: Can't convert %s to a number: %s\n", *check_installed_memory, err.Error()))
			os.Exit(NAGIOS_UNKNOWN)
		}
		status, _ = CheckInstalledCpus(rf, *system_id, c)
	} else if *check_thermal {
		//
	} else if *check_psu != "" {
		splitted := strings.Split(*check_psu, ",")
		if len(splitted) != 2 {
			fmt.Fprintf(os.Stderr, "ERROR: Invalid format for -check-psu\n")
			ShowUsage()
			os.Exit(NAGIOS_UNKNOWN)
		}

		w, err := strconv.Atoi(splitted[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("ERROR: Can't convert %s to a number: %s\n", splitted[0], err.Error()))
			os.Exit(NAGIOS_UNKNOWN)
		}
		c, err := strconv.Atoi(splitted[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("ERROR: Can't convert %s to a number: %s\n", splitted[1], err.Error()))
			os.Exit(NAGIOS_UNKNOWN)
		}
		if w <= 0 || c <= 0 {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("ERROR: Warning and critical threshold must be greater than null"))
			os.Exit(NAGIOS_UNKNOWN)
		}

		if c > w {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("ERROR: Critical threshold must be greater or equal than warning threshold\n"))
			os.Exit(NAGIOS_UNKNOWN)
		}

		status, _ = CheckPsu(rf, *chassis_id, w, c)
	} else if *check_general {
		status, _ = CheckGeneralHealth(rf, *system_id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
			os.Exit(NAGIOS_UNKNOWN)
		}
	}

	rc, msg := ProcessStatus(status)
	fmt.Println(msg)

	os.Exit(rc)
}
