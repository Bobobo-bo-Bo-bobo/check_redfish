package main

import "fmt"

const help string = `redfish-tool version 1.0.0
Copyright (C) 2018 - 2019 by Andreas Maus <maus@ypbind.de>
This program comes with ABSOLUTELY NO WARRANTY.

redfish-tool is distributed under the Terms of the GNU General
Public License Version 3. (http://www.gnu.org/copyleft/gpl.html)

Usage: check_redfish -host=<host> -user=<user> -password=<pass>|-password-file=<pwdfile> [-insecure-ssl] 
    [-chassis-id=<id>] [-system-id=<id>] [-check-installed-memory=<mem_gb>] [-timeout=<sec>] 
    [-check-installed-cpus=<cpu>] [-check-termal] [-check-psu=<warn>,<crit>] [-check-general-health]

    -host=<host>
        Hostname or IP address of management board
    -user=<user>
        Username for authentication
    -password=<pass>
        Password for authentication
    -password-file=<pwdfile>
        Read password for authentication from file, password is exepected in the first line of the file
    -insecure-ssl
        Don't verifiy SSL certificate. Default: Verify SSL certificate
    -chassis-id=<id>
        Check specific chassis. Default: First chassis reported will be checked
    -system-id=<id>
        Check specific system. Default: First system reported will be checked
    -timeout=<sec>
        Connection timeout in seconds. Default: 60
    -check-installed-memory=<mem_gb>
        Check if installed memory is recognized with <mem_gb> GByte of memory
    -check-installed-cpus=<cpu>
        Check number of installed cpus
    -check-thermal
        Check thermal subsystem
    -check-fans
        Check system fans
    -check-psu=<warn>,<crit>
        Check installed PSU, report <warn>/<crit> if <warn>/<crit> or less working PSUs are reported
    -check-voltages
        Check system voltage readings
    -check-general-health
        Check general health. This is the default when no check has been requested
`

func ShowUsage() {
	fmt.Println(help)
}
