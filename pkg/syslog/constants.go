package syslog

// Facility codes as defined in RFC 5424
const (
	FacilityKern     = 0  // kernel messages
	FacilityUser     = 1  // user-level messages
	FacilityMail     = 2  // mail system
	FacilityDaemon   = 3  // system daemons
	FacilityAuth     = 4  // security/authorization messages
	FacilitySyslog   = 5  // messages generated internally by syslogd
	FacilityLpr      = 6  // line printer subsystem
	FacilityNews     = 7  // network news subsystem
	FacilityUucp     = 8  // UUCP subsystem
	FacilityCron     = 9  // clock daemon
	FacilityAuthPriv = 10 // security/authorization messages
	FacilityFtp      = 11 // FTP daemon
	FacilityLocal0   = 16 // local use 0
	FacilityLocal1   = 17 // local use 1
	FacilityLocal2   = 18 // local use 2
	FacilityLocal3   = 19 // local use 3
	FacilityLocal4   = 20 // local use 4
	FacilityLocal5   = 21 // local use 5
	FacilityLocal6   = 22 // local use 6
	FacilityLocal7   = 23 // local use 7
)

// Severity codes as defined in RFC 5424
const (
	SeverityEmergency = 0 // system is unusable
	SeverityAlert     = 1 // action must be taken immediately
	SeverityCritical  = 2 // critical conditions
	SeverityError     = 3 // error conditions
	SeverityWarning   = 4 // warning conditions
	SeverityNotice    = 5 // normal but significant condition
	SeverityInfo      = 6 // informational messages
	SeverityDebug     = 7 // debug-level messages
)

// FacilityName returns the human-readable name for a facility code
func FacilityName(facility int) string {
	names := map[int]string{
		FacilityKern:     "kern",
		FacilityUser:     "user",
		FacilityMail:     "mail",
		FacilityDaemon:   "daemon",
		FacilityAuth:     "auth",
		FacilitySyslog:   "syslog",
		FacilityLpr:      "lpr",
		FacilityNews:     "news",
		FacilityUucp:     "uucp",
		FacilityCron:     "cron",
		FacilityAuthPriv: "authpriv",
		FacilityFtp:      "ftp",
		FacilityLocal0:   "local0",
		FacilityLocal1:   "local1",
		FacilityLocal2:   "local2",
		FacilityLocal3:   "local3",
		FacilityLocal4:   "local4",
		FacilityLocal5:   "local5",
		FacilityLocal6:   "local6",
		FacilityLocal7:   "local7",
	}
	if name, ok := names[facility]; ok {
		return name
	}
	return "unknown"
}

// SeverityName returns the human-readable name for a severity code
func SeverityName(severity int) string {
	names := map[int]string{
		SeverityEmergency: "emergency",
		SeverityAlert:     "alert",
		SeverityCritical:  "critical",
		SeverityError:     "error",
		SeverityWarning:   "warning",
		SeverityNotice:    "notice",
		SeverityInfo:      "info",
		SeverityDebug:     "debug",
	}
	if name, ok := names[severity]; ok {
		return name
	}
	return "unknown"
}
