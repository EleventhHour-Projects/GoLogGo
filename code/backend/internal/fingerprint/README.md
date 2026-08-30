# GoLogGo Log Fingerprinting Subsystem

A high-performance, deterministic Go package that generates **structural fingerprints** for logs.

Instead of hashing the raw log (which changes constantly due to timestamps, IPs, and IDs), this package extracts the **structural schema** of a log and converts it into a 64-character SHA-256 hash.

---

## 🎯 Why Do We Need This?

When GoLogGo ingests logs at scale (from firewalls, servers, databases, cloud providers, and proprietary apps), it needs to know:

> *"Do I already have a parser that knows how to parse this specific log structure?"*

If we hashed raw logs directly:
```
Log 1: 10:00:01 User alice logged in from 10.0.0.1
Log 2: 10:00:02 User bob logged in from 192.168.1.50
```
Raw hashes would be completely different even though both logs have the **exact same structure**.

With GoLogGo Fingerprinting:
```
Both become -> User <USER> logged in from <IPV4>
Hash        -> 48ab70570eacc9bc4a113257af310c57cb882da6b372b6c15e4cb05f9f6246da
```
Both logs get the **exact same fingerprint hash**!

---

## ⚡ How It Works (2-Stage Pipeline)

```
┌──────────────────────────────────────────────────────────┐
│                         RAW LOG                          │
└────────────────────────────┬─────────────────────────────┘
                             │
                             ▼
 1. ExtractFingerprintFeatures(rawLog)
    ├── Detects format (JSON, Key-Value, Syslog, CEF, XML, etc.)
    ├── Extracts keys, nested schemas, delimiters, and attributes
    ├── Replaces volatile values (IPs, dates, UUIDs, numbers)
    └── Keeps schema constants (e.g. action=deny, version=2, status=200)
                             │
                             ▼
                    FingerprintFeatures
                             │
                             ▼
 2. GenerateFingerprintHash(features)
    ├── Sorts all keys and fields in strict alphabetical order
    ├── Builds a canonical string ("fp-v1:format=json|schema=...")
    └── Computes deterministic SHA-256 hex string
                             │
                             ▼
        "8e063f707f1ea5d5f2a1eb509eeab159f81eb55776dca33777d0dd6d2fdbcb74"
```

---

## 📖 Developer Usage Guide

### 1. Basic Usage: Get Hash for Any Log

```go
package main

import (
    "fmt"
    "github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/fingerprint"
)

func main() {
    rawLog := `{"timestamp": "2026-08-30T13:02:11Z", "src_ip": "10.0.0.1", "action": "deny", "port": 443}`

    // 1. Extract structural features
    features := fingerprint.ExtractFingerprintFeatures(rawLog)

    // 2. Generate the 64-character SHA-256 fingerprint hash
    hash := fingerprint.GenerateFingerprintHash(features)

    fmt.Printf("Log Format: %s\n", features.Format) // Output: json
    fmt.Printf("Hash:       %s\n", hash)            // Output: 64-character hex string
}
```

---

### 2. Convenience Helper Function

If you just need the hash directly in your code:

```go
// FingerprintLog is a convenience wrapper for one-liner hashing
func FingerprintLog(rawLog string) (hash string, format fingerprint.FormatType) {
    features := fingerprint.ExtractFingerprintFeatures(rawLog)
    return fingerprint.GenerateFingerprintHash(features), features.Format
}
```

---

### 3. Inspecting Extracted Structural Features

`ExtractFingerprintFeatures` gives you rich metadata about what was detected:

```go
features := fingerprint.ExtractFingerprintFeatures(rawLog)

fmt.Println("Format:     ", features.Format)      // json, key_value, syslog, cef, etc.
fmt.Println("Field Count:", features.FieldCount)  // Number of fields/columns
fmt.Println("Sorted Keys:", features.Keys)        // ["action", "port", "src_ip", "timestamp"]
fmt.Println("Vendor Hint:", features.VendorHint)  // "cisco", "fortinet", "palo_alto", etc.
fmt.Println("Product Hint:", features.ProductHint) // "asa", "fortigate", "pan_os", etc.
```

---

### 4. Debugging: Inspecting the Canonical String

If you want to see **why** two logs produced the same or different hashes, print `CanonicalRepresentation`:

```go
features := fingerprint.ExtractFingerprintFeatures(rawLog)
canonicalStr := fingerprint.CanonicalRepresentation(features)

fmt.Println("Canonical Representation:")
fmt.Println(canonicalStr)
// Example output:
// fp-v1:format=json|schema=object{action:string,port:number,src_ip:ipv4,timestamp:timestamp}
```

---

### 5. Worker Ingestion Pipeline Pattern

Here is how you typically use this inside GoLogGo worker goroutines to perform fast $O(1)$ parser candidate lookups:

```go
package worker

import (
    "context"
    "log"
    "github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/fingerprint"
)

type ParserRegistry interface {
    GetParsersByFingerprint(ctx context.Context, fingerprintHash string) ([]Parser, error)
}

func ProcessIncomingLog(ctx context.Context, registry ParserRegistry, rawLog string) error {
    // 1. Fast, zero-allocation feature extraction & hashing (< 50 microseconds)
    features := fingerprint.ExtractFingerprintFeatures(rawLog)
    fpHash := fingerprint.GenerateFingerprintHash(features)

    // 2. Query registry for registered candidate parsers
    candidateParsers, err := registry.GetParsersByFingerprint(ctx, fpHash)
    if err != nil {
        return err
    }

    if len(candidateParsers) == 0 {
        log.Printf("New unseen log schema detected [format=%s, hash=%s], routing to schema learner", features.Format, fpHash)
        // Route to GoLogGo parser generation / ML worker
        return nil
    }

    // 3. Execute matched parser
    matchedParser := candidateParsers[0]
    parsedEvent, err := matchedParser.Parse(rawLog)
    if err != nil {
        return err
    }

    log.Printf("Successfully parsed log with parser %s", matchedParser.Name())
    _ = parsedEvent
    return nil
}
```

---

## Supported Formats & Normalization

| Format | How It Is Processed | Example |
| :--- | :--- | :--- |
| **JSON** | Extracts recursive schema tree, sorts keys, infers field types (ipv4, uuid, timestamp, array, object). Key order does not matter. | `{"src": "10.0.0.1", "port": 443}` |
| **Key=Value** | Parses `k=v`, `k="v"`, sorts keys, extracts delimiters (`=`, ` `), detects field types. | `srcip=10.0.0.1 dstip=8.8.8.8 action=deny` |
| **Syslog (RFC 3164 & 5424)** | Extracts PRI, Facility, Severity, App/Tag, Structured Data keys; normalizes volatile hostnames, PIDs, and body. | `<34>Aug 30 13:02:11 host sshd[1234]: Failed password...` |
| **CEF (ArcSight)** | Extracts 7-part header metadata and sorted extension key-value pairs. | `CEF:0\|Vendor\|Product\|1.0\|1001\|Drop\|6\|src=10.0.0.1` |
| **LEEF (QRadar)** | Extracts LEEF header attributes, custom delimiters, and extension keys. | `LEEF:2.0\|Microsoft\|MSExchange\|2016\|15.1\|x09\|src=10.0.0.1` |
| **CSV / Tabular** | Detects delimiter (`,`, `\t`, `\|`), column count, header presence, and column data types. | `2026-08-30,10.0.0.1,8.8.8.8,443,ALLOW` |
| **XML** | Parses element hierarchy and attribute names without including volatile text bodies. | `<Event><System><EventID>4624</EventID></System></Event>` |
| **Multiline / Stack Traces** | Recognizes Java/Python/Go stack traces and captures exception class and frame templates. | `java.lang.NullPointerException\n at com.example...` |
| **Plain Text / Unknown** | Replaces volatile tokens with placeholders and preserves static message template. | `Connection from 10.0.0.5:52341 accepted` |

---

## Context-Aware Normalization

Unlike naive regex cleaners that blindly replace every number with `<NUMBER>`, GoLogGo uses **context-aware normalization**:

- **Preserves Schema Enums & Constants:**
  - `version=2` $\rightarrow$ keeps `2` (version number)
  - `severity=6` $\rightarrow$ keeps `6` (syslog/event severity)
  - `status=200` $\rightarrow$ keeps `200` (HTTP status code)
  - `proto=6` $\rightarrow$ keeps `6` (TCP protocol number)
  - `action=deny` vs `action=allow` $\rightarrow$ preserved to differentiate distinct event types
- **Replaces Volatile Tokens:**
  - `10.20.30.40` $\rightarrow$ `<IPV4>`
  - `2001:db8::1` $\rightarrow$ `<IPV6>`
  - `00:11:22:33:44:55` $\rightarrow$ `<MAC>`
  - `550e8400-e29b-41d4-a716-446655440000` $\rightarrow$ `<UUID>`
  - `2026-08-30T13:02:11Z` $\rightarrow$ `<TIMESTAMP>`
  - `inside:10.0.0.1/52341` $\rightarrow$ `inside:<IPV4>:<PORT>`
  - `session-abc123-98231` $\rightarrow$ `<DYNAMIC_ID>`

---

## How Unknown & Custom Logs are Handled

If a log does not match JSON, XML, or standard formats (e.g. custom IoT device or proprietary banking software):

1. It automatically routes to `FormatPlainText`.
2. All dynamic variables (IPs, dates, ports, sequence numbers, IDs) are abstracted into semantic tokens.
3. The remaining static boilerplate words form the template signature:
   ```
   Raw:      "TRANSACTION_TX node-9812: Transfer 500.00 from 10.0.1.5 to 10.0.2.9 completed in 42ms"
   Template: "TRANSACTION_TX <DYNAMIC_ID>: Transfer <NUMBER> from <IPV4> to <IPV4> completed in <NUMBER>ms"
   ```
4. Future logs with different node IDs, transfer amounts, or IPs will generate the **exact same hash**.

---

## Robustness & Zero-Panic Guarantee

- **No Panics:** Handles corrupted JSON, truncated XML, null bytes (`\x00`), and empty strings safely via recovery guards.
- **Fast:** Runs completely in-memory with zero network, database, or ML dependencies.
- **Microsecond Latency:** Takes under $50\mu\text{s}$ per log, ideal for high-throughput worker goroutines.

---

## Testing & Benchmarks

Run unit tests with race detection:
```bash
go test -v -race ./code/backend/internal/fingerprint/...
```

Run benchmarks:
```bash
go test -bench=. -benchmem ./code/backend/internal/fingerprint/...
```
