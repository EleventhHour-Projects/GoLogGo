package fingerprint

import (
	"fmt"
	"testing"
)

// TestJSONEquivalence verifies that JSON logs with identical structure but different volatile values produce identical fingerprints.
func TestJSONEquivalence(t *testing.T) {
	log1 := `{
		"timestamp": "2026-08-30T13:02:11Z",
		"src_ip": "10.0.0.1",
		"dst_ip": "8.8.8.8",
		"action": "deny",
		"port": 443,
		"request_id": "550e8400-e29b-41d4-a716-446655440000",
		"user": {
			"id": 1024,
			"email": "alice@example.com"
		},
		"tags": ["prod", "security"]
	}`

	log2 := `{
		"dst_ip": "1.1.1.1",
		"action": "deny",
		"user": {
			"email": "bob@company.org",
			"id": 9999
		},
		"tags": ["staging"],
		"port": 8080,
		"request_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"src_ip": "192.168.1.50",
		"timestamp": "2026-09-01T04:15:30.123+02:00"
	}`

	feat1 := ExtractFingerprintFeatures(log1)
	feat2 := ExtractFingerprintFeatures(log2)

	hash1 := GenerateFingerprintHash(feat1)
	hash2 := GenerateFingerprintHash(feat2)

	if feat1.Format != FormatJSON || feat2.Format != FormatJSON {
		t.Fatalf("expected FormatJSON, got %s and %s", feat1.Format, feat2.Format)
	}

	if hash1 != hash2 {
		t.Errorf("expected identical hash for structurally identical JSON logs,\nlog1 hash: %s\nlog2 hash: %s\ncanonical1: %s\ncanonical2: %s",
			hash1, hash2, CanonicalRepresentation(feat1), CanonicalRepresentation(feat2))
	}
}

// TestJSONDistinctness verifies that JSON logs with different schemas produce distinct hashes.
func TestJSONDistinctness(t *testing.T) {
	log1 := `{"event": "login", "user": "admin", "src_ip": "10.0.0.1"}`
	log2 := `{"event": "login", "user": "admin", "src_ip": "10.0.0.1", "session_id": "abc12345"}`

	feat1 := ExtractFingerprintFeatures(log1)
	feat2 := ExtractFingerprintFeatures(log2)

	hash1 := GenerateFingerprintHash(feat1)
	hash2 := GenerateFingerprintHash(feat2)

	if hash1 == hash2 {
		t.Errorf("expected distinct hashes for different JSON schemas, got %s", hash1)
	}
}

// TestKVEquivalence verifies Key-Value logs with volatile value and ordering differences produce identical hashes.
func TestKVEquivalence(t *testing.T) {
	log1 := `date=2026-08-30 time=13:02:11 devname=FG100D devid=FG100D3G15012345 type=traffic action=deny srcip=10.0.0.5 dstip=8.8.8.8 srcport=52341 dstport=443 sessionid=98231`
	log2 := `devid="FG100D3G99999999" devname="FG100D" type="traffic" date=2026-09-01 time=14:30:00 dstip=1.1.1.1 srcip=172.16.0.22 dstport=80 srcport=49152 sessionid=112233 action="deny"`

	feat1 := ExtractFingerprintFeatures(log1)
	feat2 := ExtractFingerprintFeatures(log2)

	hash1 := GenerateFingerprintHash(feat1)
	hash2 := GenerateFingerprintHash(feat2)

	if feat1.Format != FormatKeyValue || feat2.Format != FormatKeyValue {
		t.Fatalf("expected FormatKeyValue, got %s and %s", feat1.Format, feat2.Format)
	}

	if hash1 != hash2 {
		t.Errorf("expected identical hash for structurally identical KV logs,\nhash1: %s\nhash2: %s\ncan1: %s\ncan2: %s",
			hash1, hash2, CanonicalRepresentation(feat1), CanonicalRepresentation(feat2))
	}
}

// TestSyslogRFC5424Equivalence verifies RFC 5424 syslog logs.
func TestSyslogRFC5424Equivalence(t *testing.T) {
	log1 := `<165>1 2026-08-30T13:02:11.003Z myhost.example.com myapp 1234 ID47 [exampleSDID@32473 iut="3" eventSource="Application"] User 10.0.0.1 authenticated`
	log2 := `<165>1 2026-09-01T15:20:00.999Z otherhost.org myapp 5678 ID47 [exampleSDID@32473 iut="3" eventSource="Application"] User 192.168.1.10 authenticated`

	feat1 := ExtractFingerprintFeatures(log1)
	feat2 := ExtractFingerprintFeatures(log2)

	hash1 := GenerateFingerprintHash(feat1)
	hash2 := GenerateFingerprintHash(feat2)

	if feat1.Format != FormatSyslog || feat2.Format != FormatSyslog {
		t.Fatalf("expected FormatSyslog, got %s and %s", feat1.Format, feat2.Format)
	}

	if hash1 != hash2 {
		t.Errorf("expected identical hash for RFC5424 syslog,\nhash1: %s\nhash2: %s\ncan1: %s\ncan2: %s",
			hash1, hash2, CanonicalRepresentation(feat1), CanonicalRepresentation(feat2))
	}
}

// TestSyslogRFC3164Equivalence verifies BSD Syslog logs.
func TestSyslogRFC3164Equivalence(t *testing.T) {
	log1 := `<34>Aug 30 13:02:11 server01 sshd[12345]: Failed password for invalid user admin from 10.0.0.5 port 52341 ssh2`
	log2 := `<34>Sep  1 04:15:30 backup-srv sshd[67890]: Failed password for invalid user guest from 192.168.1.100 port 44321 ssh2`

	feat1 := ExtractFingerprintFeatures(log1)
	feat2 := ExtractFingerprintFeatures(log2)

	hash1 := GenerateFingerprintHash(feat1)
	hash2 := GenerateFingerprintHash(feat2)

	if feat1.Format != FormatSyslog || feat2.Format != FormatSyslog {
		t.Fatalf("expected FormatSyslog, got %s and %s", feat1.Format, feat2.Format)
	}

	if hash1 != hash2 {
		t.Errorf("expected identical hash for RFC3164 syslog,\nhash1: %s\nhash2: %s\ncan1: %s\ncan2: %s",
			hash1, hash2, CanonicalRepresentation(feat1), CanonicalRepresentation(feat2))
	}
}

// TestCEFEquivalence verifies ArcSight CEF logs.
func TestCEFEquivalence(t *testing.T) {
	log1 := `CEF:0|SecurityCorp|NextGenFW|1.0|1001|Connection Blocked|6|src=10.0.0.5 dst=8.8.8.8 spt=52341 dpt=443 proto=TCP act=drop`
	log2 := `CEF:0|SecurityCorp|NextGenFW|1.0|1001|Connection Blocked|6|dst=1.1.1.1 src=192.168.1.1 spt=49152 dpt=80 act=drop proto=TCP`

	feat1 := ExtractFingerprintFeatures(log1)
	feat2 := ExtractFingerprintFeatures(log2)

	hash1 := GenerateFingerprintHash(feat1)
	hash2 := GenerateFingerprintHash(feat2)

	if feat1.Format != FormatCEF || feat2.Format != FormatCEF {
		t.Fatalf("expected FormatCEF, got %s and %s", feat1.Format, feat2.Format)
	}

	if hash1 != hash2 {
		t.Errorf("expected identical hash for CEF logs with reordered extensions,\nhash1: %s\nhash2: %s\ncan1: %s\ncan2: %s",
			hash1, hash2, CanonicalRepresentation(feat1), CanonicalRepresentation(feat2))
	}
}

// TestLEEFEquivalence verifies IBM QRadar LEEF logs.
func TestLEEFEquivalence(t *testing.T) {
	log1 := `LEEF:2.0|Trend Micro|Deep Security|9.0|1000000|x09|src=10.0.0.1	dst=192.168.1.2	sev=5	action=blocked`
	log2 := `LEEF:2.0|Trend Micro|Deep Security|9.0|1000000|x09|action=blocked	dst=172.16.0.5	src=10.20.30.40	sev=5`

	feat1 := ExtractFingerprintFeatures(log1)
	feat2 := ExtractFingerprintFeatures(log2)

	hash1 := GenerateFingerprintHash(feat1)
	hash2 := GenerateFingerprintHash(feat2)

	if feat1.Format != FormatLEEF || feat2.Format != FormatLEEF {
		t.Fatalf("expected FormatLEEF, got %s and %s", feat1.Format, feat2.Format)
	}

	if hash1 != hash2 {
		t.Errorf("expected identical hash for LEEF logs,\nhash1: %s\nhash2: %s", hash1, hash2)
	}
}

// TestXMLEquivalence verifies XML logs.
func TestXMLEquivalence(t *testing.T) {
	log1 := `<Event xmlns="http://schemas.microsoft.com/win/2004/08/events/event">
		<System>
			<Provider Name="Microsoft-Windows-Security-Auditing" />
			<EventID>4624</EventID>
			<TimeCreated SystemTime="2026-08-30T13:02:11.000Z" />
		</System>
		<EventData>
			<Data Name="TargetUserName">Administrator</Data>
			<Data Name="IpAddress">10.0.0.5</Data>
		</EventData>
	</Event>`

	log2 := `<Event xmlns="http://schemas.microsoft.com/win/2004/08/events/event">
		<System>
			<Provider Name="Microsoft-Windows-Security-Auditing" />
			<EventID>4624</EventID>
			<TimeCreated SystemTime="2026-09-01T04:15:30.999Z" />
		</System>
		<EventData>
			<Data Name="TargetUserName">Alice</Data>
			<Data Name="IpAddress">192.168.1.50</Data>
		</EventData>
	</Event>`

	feat1 := ExtractFingerprintFeatures(log1)
	feat2 := ExtractFingerprintFeatures(log2)

	hash1 := GenerateFingerprintHash(feat1)
	hash2 := GenerateFingerprintHash(feat2)

	if feat1.Format != FormatXML || feat2.Format != FormatXML {
		t.Fatalf("expected FormatXML, got %s and %s", feat1.Format, feat2.Format)
	}

	if hash1 != hash2 {
		t.Errorf("expected identical hash for XML logs,\nhash1: %s\nhash2: %s", hash1, hash2)
	}
}

// TestCSVEquivalence verifies delimited tabular logs.
func TestCSVEquivalence(t *testing.T) {
	log1 := `2026-08-30T13:02:11Z,10.0.0.1,8.8.8.8,443,ALLOW,98234`
	log2 := `2026-09-01T04:15:30Z,192.168.1.10,1.1.1.1,80,ALLOW,12345`

	feat1 := ExtractFingerprintFeatures(log1)
	feat2 := ExtractFingerprintFeatures(log2)

	hash1 := GenerateFingerprintHash(feat1)
	hash2 := GenerateFingerprintHash(feat2)

	if feat1.Format != FormatCSV || feat2.Format != FormatCSV {
		t.Fatalf("expected FormatCSV, got %s and %s", feat1.Format, feat2.Format)
	}

	if hash1 != hash2 {
		t.Errorf("expected identical hash for CSV logs,\nhash1: %s\nhash2: %s", hash1, hash2)
	}
}

// TestPlainTextEquivalence verifies unstructured plain text templates.
func TestPlainTextEquivalence(t *testing.T) {
	log1 := `Connection from 10.0.0.5:52341 to 192.168.1.10:443 accepted`
	log2 := `Connection from 172.16.0.8:49152 to 10.20.30.40:8080 accepted`

	feat1 := ExtractFingerprintFeatures(log1)
	feat2 := ExtractFingerprintFeatures(log2)

	hash1 := GenerateFingerprintHash(feat1)
	hash2 := GenerateFingerprintHash(feat2)

	if feat1.Format != FormatPlainText || feat2.Format != FormatPlainText {
		t.Fatalf("expected FormatPlainText, got %s and %s", feat1.Format, feat2.Format)
	}

	if hash1 != hash2 {
		t.Errorf("expected identical hash for normalized plain text logs,\nhash1: %s\nhash2: %s\ncan1: %s\ncan2: %s",
			hash1, hash2, CanonicalRepresentation(feat1), CanonicalRepresentation(feat2))
	}
}

// TestMultilineEquivalence verifies multi-line application stack traces.
func TestMultilineEquivalence(t *testing.T) {
	log1 := `java.lang.NullPointerException: Cannot invoke method on null object
	at com.example.service.PaymentService.processOrder(PaymentService.java:142)
	at com.example.controller.OrderController.handleRequest(OrderController.java:58)
	at javax.servlet.http.HttpServlet.service(HttpServlet.java:750)`

	log2 := `java.lang.NullPointerException: Cannot invoke method on null object
	at com.example.service.PaymentService.processOrder(PaymentService.java:999)
	at com.example.controller.OrderController.handleRequest(OrderController.java:214)
	at javax.servlet.http.HttpServlet.service(HttpServlet.java:800)`

	feat1 := ExtractFingerprintFeatures(log1)
	feat2 := ExtractFingerprintFeatures(log2)

	hash1 := GenerateFingerprintHash(feat1)
	hash2 := GenerateFingerprintHash(feat2)

	if feat1.Format != FormatMultiline || feat2.Format != FormatMultiline {
		t.Fatalf("expected FormatMultiline, got %s and %s", feat1.Format, feat2.Format)
	}

	if hash1 != hash2 {
		t.Errorf("expected identical hash for multiline stack traces with different line numbers,\nhash1: %s\nhash2: %s", hash1, hash2)
	}
}

// TestVendorDistinctness verifies that different vendors (Cisco vs Fortinet) produce different fingerprints.
func TestVendorDistinctness(t *testing.T) {
	ciscoLog := `%ASA-4-106023: Deny tcp src inside:10.0.0.1/52341 dst outside:8.8.8.8/443 by access-group "acl_in"`
	fortinetLog := `devname="FG100D" type="traffic" action="deny" srcip=10.0.0.1 dstip=8.8.8.8 srcport=52341 dstport=443`

	feat1 := ExtractFingerprintFeatures(ciscoLog)
	feat2 := ExtractFingerprintFeatures(fortinetLog)

	hash1 := GenerateFingerprintHash(feat1)
	hash2 := GenerateFingerprintHash(feat2)

	if hash1 == hash2 {
		t.Errorf("expected distinct hashes for Cisco vs Fortinet logs, got %s", hash1)
	}
}

// TestRealWorldLogTypesEquivalence tests 10 distinct real-world log types and ensures volatile value invariance.
func TestRealWorldLogTypesEquivalence(t *testing.T) {
	tests := []struct {
		name        string
		log1        string
		log2        string
		expectedFmt FormatType
	}{
		{
			name:        "Cisco ASA Firewall",
			log1:        "%ASA-4-106023: Deny tcp src inside:10.0.0.1/52341 dst outside:8.8.8.8/443 by access-group acl_in [0x0, 0x0]",
			log2:        "%ASA-4-106023: Deny tcp src inside:172.16.1.50/49152 dst outside:1.1.1.1/80 by access-group acl_in [0x0, 0x0]",
			expectedFmt: FormatPlainText,
		},
		{
			name:        "Palo Alto PAN-OS Traffic CSV",
			log1:        "1,2026/08/30 13:02:11,001801000001,TRAFFIC,drop,1,2026/08/30 13:02:11,10.0.0.5,8.8.8.8,0.0.0.0,0.0.0.0,rule1,user1,,web-browsing,vsys1,trust,untrust,ethernet1/1,ethernet1/2,log_forward,2026/08/30 13:02:11,12345,1,52341,443,0,0,0x0,tcp,deny,100,100,0,1,2026/08/30 13:02:11,0,any,0,987654321,0x0,10.0.0.0-10.255.255.255,United States,0,1,0,drop,0,0,0,0,,PA-VM,from-policy",
			log2:        "1,2026/09/01 04:15:30,001801000002,TRAFFIC,drop,1,2026/09/01 04:15:30,192.168.1.50,1.1.1.1,0.0.0.0,0.0.0.0,rule1,user2,,web-browsing,vsys1,trust,untrust,ethernet1/1,ethernet1/2,log_forward,2026/09/01 04:15:30,67890,1,49152,80,0,0,0x0,tcp,deny,250,250,0,1,2026/09/01 04:15:30,0,any,0,123456789,0x0,192.168.0.0-192.168.255.255,Australia,0,1,0,drop,0,0,0,0,,PA-VM,from-policy",
			expectedFmt: FormatCSV,
		},
		{
			name:        "Fortinet FortiGate Key-Value",
			log1:        `date=2026-08-30 time=13:02:11 devname=FG100D devid=FG100D3G15012345 logid="0000000013" type="traffic" subtype="forward" level="notice" vd="root" srcip=10.0.0.1 dstip=8.8.8.8 srcport=52341 dstport=443 action="deny" proto=6`,
			log2:        `date=2026-09-01 time=15:45:00 devname=FG100D devid=FG100D3G99999999 logid="0000000013" type="traffic" subtype="forward" level="notice" vd="root" srcip=172.16.0.2 dstip=1.1.1.1 srcport=49152 dstport=80 action="deny" proto=6`,
			expectedFmt: FormatKeyValue,
		},
		{
			name:        "AWS VPC Flow Log",
			log1:        "2 123456789010 eni-1235b678 172.31.16.139 172.31.16.21 20641 22 6 20 4249 1418530010 1418530070 ACCEPT OK",
			log2:        "2 123456789010 eni-8765a432 10.0.1.50 10.0.2.20 54321 22 6 15 3120 1418540010 1418540070 ACCEPT OK",
			expectedFmt: FormatPlainText,
		},
		{
			name:        "Linux SSHD RFC3164 Syslog",
			log1:        `<34>Aug 30 13:02:11 auth-server sshd[12345]: Accepted publickey for user admin from 10.0.0.5 port 52341 ssh2: RSA SHA256:abc1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef`,
			log2:        `<34>Sep  1 05:22:19 backup-host sshd[67890]: Accepted publickey for user guest from 192.168.1.50 port 49152 ssh2: RSA SHA256:fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321`,
			expectedFmt: FormatSyslog,
		},
		{
			name:        "Nginx Web Access Log",
			log1:        `10.0.0.5 - user_1 [30/Aug/2026:13:02:11 +0000] "GET /api/v1/users HTTP/1.1" 200 4523 "https://example.com/dashboard" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"`,
			log2:        `192.168.1.10 - user_2 [01/Sep/2026:04:15:30 +0000] "GET /api/v1/users HTTP/1.1" 200 4523 "https://company.org/home" "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"`,
			expectedFmt: FormatPlainText,
		},
		{
			name:        "ArcSight CEF Security Log",
			log1:        `CEF:0|CheckPoint|VPN-1 & FireWall-1|CheckPoint|drop|Drop|High|src=10.0.0.1 dst=8.8.8.8 spt=52341 dpt=443 proto=6 act=drop cs1=Rule1`,
			log2:        `CEF:0|CheckPoint|VPN-1 & FireWall-1|CheckPoint|drop|Drop|High|cs1=Rule1 act=drop proto=6 dpt=80 spt=49152 dst=1.1.1.1 src=192.168.1.20`,
			expectedFmt: FormatCEF,
		},
		{
			name:        "IBM QRadar LEEF Log",
			log1:        `LEEF:2.0|Microsoft|MSExchange|2016|15.1.1466.3|x09|src=10.0.0.1	dst=192.168.1.2	usrName=user1@domain.com	sev=5`,
			log2:        `LEEF:2.0|Microsoft|MSExchange|2016|15.1.1466.3|x09|usrName=user2@company.org	sev=5	dst=172.16.0.5	src=10.20.30.40`,
			expectedFmt: FormatLEEF,
		},
		{
			name:        "Windows Event XML",
			log1:        `<Event xmlns="http://schemas.microsoft.com/win/2004/08/events/event"><System><Provider Name="Microsoft-Windows-Security-Auditing"/><EventID>4688</EventID><TimeCreated SystemTime="2026-08-30T13:02:11.000Z"/></System><EventData><Data Name="NewProcessName">C:\Windows\System32\cmd.exe</Data><Data Name="ProcessId">0x1234</Data></EventData></Event>`,
			log2:        `<Event xmlns="http://schemas.microsoft.com/win/2004/08/events/event"><System><Provider Name="Microsoft-Windows-Security-Auditing"/><EventID>4688</EventID><TimeCreated SystemTime="2026-09-01T04:15:30.999Z"/></System><EventData><Data Name="NewProcessName">C:\Windows\System32\powershell.exe</Data><Data Name="ProcessId">0x5678</Data></EventData></Event>`,
			expectedFmt: FormatXML,
		},
		{
			name:        "Java Application Stack Trace Multiline",
			log1:        "java.lang.IllegalArgumentException: Invalid transaction payload\n\tat com.payment.Gateway.execute(Gateway.java:105)\n\tat com.payment.Processor.handle(Processor.java:42)\nCaused by: java.io.IOException: Connection reset by peer\n\tat java.net.SocketInputStream.read(SocketInputStream.java:210)",
			log2:        "java.lang.IllegalArgumentException: Invalid transaction payload\n\tat com.payment.Gateway.execute(Gateway.java:210)\n\tat com.payment.Processor.handle(Processor.java:88)\nCaused by: java.io.IOException: Connection reset by peer\n\tat java.net.SocketInputStream.read(SocketInputStream.java:315)",
			expectedFmt: FormatMultiline,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feat1 := ExtractFingerprintFeatures(tt.log1)
			feat2 := ExtractFingerprintFeatures(tt.log2)

			if feat1.Format != tt.expectedFmt {
				t.Errorf("%s: expected format %s, got %s", tt.name, tt.expectedFmt, feat1.Format)
			}

			hash1 := GenerateFingerprintHash(feat1)
			hash2 := GenerateFingerprintHash(feat2)

			if hash1 != hash2 {
				t.Errorf("%s: expected identical hash for equivalent logs\nhash1: %s\nhash2: %s\ncan1:  %s\ncan2:  %s",
					tt.name, hash1, hash2, CanonicalRepresentation(feat1), CanonicalRepresentation(feat2))
			}
		})
	}
}

// TestMalformedInputsNeverPanic ensures that no malformed, truncated, or invalid input panics.
func TestMalformedInputsNeverPanic(t *testing.T) {
	malformedSamples := []string{
		"",
		"   \t\n  ",
		"{",
		"{\"broken_json\":",
		"{\"a\": [1, 2, }",
		"<Event><System>",
		"<",
		"<>",
		"CEF:0|Incomplete",
		"LEEF:2.0|",
		"<999999> Invalid Pri",
		"a= b= c=",
		",,,,",
		"\x00\x01\x02\xff\xfe",
		string(make([]byte, 100000)), // large payload of zeroes
		`{"nested": {"deep": {"level": {"key": 123}}}}`,
		"2026-08-30T13:02:11Z single_line",
	}

	for i, sample := range malformedSamples {
		t.Run(fmt.Sprintf("Sample_%d", i), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("ExtractFingerprintFeatures panicked on input %d: %v", i, r)
				}
			}()

			feat := ExtractFingerprintFeatures(sample)
			hash := GenerateFingerprintHash(feat)
			if len(hash) != 64 {
				t.Errorf("expected 64-char hex hash, got len %d: %s", len(hash), hash)
			}
		})
	}
}
