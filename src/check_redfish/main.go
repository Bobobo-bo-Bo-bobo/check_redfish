package main

import (
	"flag"
	"fmt"
	redfish "git.ypbind.de/repository/go-redfish.git"
	"os"
	"time"
)

func main() {
	var user = flag.String("user", "", "Username for authentication")
	var password = flag.String("password", "", "Password for authentication")
	var password_file = flag.String("password-file", "", "Read password for authentication from file, password is exepected in the first line of the file")
	var host = flag.String("host", "", "Hostname or IP address of management board")
	var port = flag.Int("port", 0, "Port to connect to")
	var insecure_ssl = flag.Bool("insecure-ssl", false, "Don't verifiy SSL certificate")
	//	var chassis_id = flag.String("chassis-id", "", "Process data of specific chassis")
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
		os.Exit(NAGIOS_UNKNOWN)
	}

	if *user == "" {
		fmt.Fprintf(os.Stderr, "ERROR: No username specified\n")
		os.Exit(NAGIOS_UNKNOWN)
	}

	if *password == "" && *password_file == "" {
		fmt.Fprintf(os.Stderr, "ERROR: No password or password file specified\n")
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
		//
	} else if *check_installed_cpus != "" {
		//
	} else if *check_thermal {
		//
	} else if *check_psu != "" {
		//
	} else if *check_general {
		status, err = CheckGeneralHealth(rf, *system_id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
			os.Exit(NAGIOS_UNKNOWN)
		}
	}

	rc, msg := ProcessStatus(status)
	fmt.Println(msg)

	os.Exit(rc)
}
