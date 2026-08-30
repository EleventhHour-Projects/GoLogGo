package fingerprint

import (
	"testing"
)

var (
	benchJSON = `{
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

	benchKV = `date=2026-08-30 time=13:02:11 devname=FG100D devid=FG100D3G15012345 type=traffic action=deny srcip=10.0.0.5 dstip=8.8.8.8 srcport=52341 dstport=443 sessionid=98231`

	benchSyslog = `<165>1 2026-08-30T13:02:11.003Z myhost.example.com myapp 1234 ID47 [exampleSDID@32473 iut="3" eventSource="Application"] User 10.0.0.1 authenticated`

	benchCEF = `CEF:0|SecurityCorp|NextGenFW|1.0|1001|Connection Blocked|6|src=10.0.0.5 dst=8.8.8.8 spt=52341 dpt=443 proto=TCP act=drop`

	benchPlainText = `Connection from 10.0.0.5:52341 to 192.168.1.10:443 accepted`

	benchMultiline = `java.lang.NullPointerException: Cannot invoke method on null object
	at com.example.service.PaymentService.processOrder(PaymentService.java:142)
	at com.example.controller.OrderController.handleRequest(OrderController.java:58)
	at javax.servlet.http.HttpServlet.service(HttpServlet.java:750)`
)

func BenchmarkExtractFingerprintFeatures_JSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ExtractFingerprintFeatures(benchJSON)
	}
}

func BenchmarkGenerateFingerprintHash_JSON(b *testing.B) {
	feat := ExtractFingerprintFeatures(benchJSON)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateFingerprintHash(feat)
	}
}

func BenchmarkPipeline_JSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		feat := ExtractFingerprintFeatures(benchJSON)
		_ = GenerateFingerprintHash(feat)
	}
}

func BenchmarkPipeline_KV(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		feat := ExtractFingerprintFeatures(benchKV)
		_ = GenerateFingerprintHash(feat)
	}
}

func BenchmarkPipeline_Syslog(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		feat := ExtractFingerprintFeatures(benchSyslog)
		_ = GenerateFingerprintHash(feat)
	}
}

func BenchmarkPipeline_CEF(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		feat := ExtractFingerprintFeatures(benchCEF)
		_ = GenerateFingerprintHash(feat)
	}
}

func BenchmarkPipeline_PlainText(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		feat := ExtractFingerprintFeatures(benchPlainText)
		_ = GenerateFingerprintHash(feat)
	}
}

func BenchmarkPipeline_Multiline(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		feat := ExtractFingerprintFeatures(benchMultiline)
		_ = GenerateFingerprintHash(feat)
	}
}
