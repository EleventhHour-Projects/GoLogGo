package fingerprint

import (
	"regexp"
	"strings"
)

var (
	// Vendor regex patterns
	ciscoAsaRegex     = regexp.MustCompile(`%(?:ASA|FTD|PIX)-\d-\d{6}`)
	ciscoIosRegex     = regexp.MustCompile(`%(?:SYS|LINK|LINEPROTO|SEC|IP)-\d-[A-Z0-9_]+`)
	fortigateRegex    = regexp.MustCompile(`(?i)\b(?:devid=FG|devname=FG|device_id=FG|type=(?:traffic|utm|event|virus|webfilter|ips)\b.*devname=)`)
	paloAltoRegex     = regexp.MustCompile(`\b(?:TRAFFIC|THREAT|CONFIG|SYSTEM|CORRELATION),(?:drop|allow|deny|alert|reset)`)
	juniperRegex      = regexp.MustCompile(`\b(?:RT_FLOW|RT_IDP|RT_UTM|RT_IPSEC|SSG|JUNOS)\b`)
	checkPointRegex   = regexp.MustCompile(`(?i)\b(?:CheckPoint|SmartDefense|fw_log)\b`)
	linuxSyslogRegex  = regexp.MustCompile(`\b(?:sshd|systemd|kernel|sudo|cron|auditd|dockerd|containerd)\[\d+\]:`)
	windowsEventRegex = regexp.MustCompile(`(?i)xmlns="http://schemas\.microsoft\.com/win/2004/08/events/event"|<EventID>\d+</EventID>|Microsoft-Windows-`)
	awsFlowRegex      = regexp.MustCompile(`\b(?:\d+ eni-[0-9a-f]+ \d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|aws:cloudtrail|aws_region)\b`)
	azureRegex        = regexp.MustCompile(`(?i)\b(?:AzureActivity|AzureDiagnostics|/SUBSCRIPTIONS/[0-9a-fA-F\-]+)\b`)
)

// DetectVendorProduct inspects the raw log and returns strong vendor/product hints if confidently identified.
func DetectVendorProduct(raw string) (vendor string, product string) {
	lower := strings.ToLower(raw)

	// Cisco ASA / IOS / FTD
	if match := ciscoAsaRegex.FindString(raw); match != "" {
		return "cisco", "asa"
	}
	if match := ciscoIosRegex.FindString(raw); match != "" {
		return "cisco", "ios"
	}

	// Fortinet FortiGate
	if fortigateRegex.MatchString(raw) || (strings.Contains(lower, "devid=\"fg") || strings.Contains(lower, "devname=\"fg")) {
		return "fortinet", "fortigate"
	}

	// Palo Alto PAN-OS
	if paloAltoRegex.MatchString(raw) || strings.Contains(raw, "PAN-OS") {
		return "palo_alto", "pan_os"
	}

	// Juniper
	if juniperRegex.MatchString(raw) {
		return "juniper", "junos"
	}

	// Check Point
	if checkPointRegex.MatchString(raw) {
		return "checkpoint", "firewall1"
	}

	// Windows Event Log
	if windowsEventRegex.MatchString(raw) {
		return "microsoft", "windows_event"
	}

	// AWS VPC Flow / CloudTrail
	if awsFlowRegex.MatchString(raw) {
		if strings.Contains(raw, "eni-") {
			return "aws", "vpc_flow"
		}
		return "aws", "cloudtrail"
	}

	// Azure
	if azureRegex.MatchString(raw) {
		return "microsoft", "azure"
	}

	// Linux Daemon / Syslog
	if linuxSyslogRegex.MatchString(raw) {
		if strings.Contains(raw, "sshd") {
			return "linux", "sshd"
		}
		if strings.Contains(raw, "systemd") {
			return "linux", "systemd"
		}
		if strings.Contains(raw, "sudo") {
			return "linux", "sudo"
		}
		if strings.Contains(raw, "cron") {
			return "linux", "cron"
		}
		if strings.Contains(raw, "auditd") {
			return "linux", "auditd"
		}
		return "linux", "system"
	}

	return "", ""
}
