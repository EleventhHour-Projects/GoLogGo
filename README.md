<p align="center">
  <img
    src="https://github.com/user-attachments/assets/c58bb217-e48d-4c40-aab4-16c8d6d3fed2"
    width="480"
    alt="GoLogGo - Universal Event Processing & Normalization Platform"
  >
</p>

GoLogGo is a universal event processing and normalization platform for ingesting, fingerprinting, parsing, and exploring logs from different sources. It identifies recurring log structures, reuses known parsers, and generates new Go-compatible parsing rules for previously unseen formats.

## 1. Project Information

- **Project Title:** GoLogGo - Universal Event Processing and Normalization Platform
- **PS ID:** SIH26156
- **PS Title:** Universal Log Pre-processing Framework
- **Category:** Software
- **Theme:** Blockchain & Cybersecurity
- **Team:** EleventhHour

## 2. Team Details

**Team Name:** EleventhHour

**Team Leader:** [@ManavSharma142](https://github.com/ManavSharma142)

**Team Members:**

- 2024UCI8028 - [@ManavSharma142](https://github.com/ManavSharma142)
- 2024UCI8013 - [@Not-Dhananjay-Mishra](https://github.com/Not-Dhananjay-Mishra)
- 2024UCI8016 - [@Vishesh1328t](https://github.com/Vishesh1328t)
- 2024UCI6532 - [@aryan-kazuha](https://github.com/aryan-kazuha)
- 2024UCI6643 - [@DevikaJain16](https://github.com/DevikaJain16)
- 2024UCI8029 - [@advikarathore](https://github.com/advikarathore)

## Project Links

- **SIH Presentation:** TODO
- **Video Demonstration:** [Youtube](https://www.youtube.com/watch?v=V8vLanBAm0E)
- **Live Deployment:** [go-log-go.vercel.app](https://go-log-go.vercel.app/)
- **Source Code:** [GitHub Repository](https://github.com/EleventhHour-Projects/GoLogGo/tree/main/code)

## 3. Problem Statement

Security, infrastructure, and application systems produce logs in many formats, including JSON, key-value, Syslog, CEF, LEEF, CSV, XML, and free-form text. Building and maintaining a separate parser for every format is slow, difficult to scale, and vulnerable to changes in log structure.

## 4. Proposed Solution

GoLogGo provides an asynchronous log processing pipeline that:

1. Accepts log events through an authenticated API.
2. Sends log events to a worker pool for concurrent processing.
3. Detects the log format and extracts structural features.
4. Generates a deterministic fingerprint so equivalent log schemas can share parsers.
5. Reuses cached and persisted parsers when a matching schema is known.
6. Sends unknown formats to a parser-generation service powered by an LLM.
7. Normalizes parsed events and makes logs, parser rules, and processing status available through the web dashboard.

## 5. Key Features

- Multi-format log fingerprinting for JSON, key-value, Syslog, CEF, LEEF, CSV, XML, multiline, and text logs
- Structural normalization of volatile values such as timestamps, IP addresses, UUIDs, and identifiers
- Deterministic SHA-256 fingerprints for fast parser lookup
- Asynchronous processing with RabbitMQ and a concurrent Go worker pool
- Parser caching with Redis and persistence with MongoDB
- LLM-assisted generation of Go-compatible parser rules through a FastAPI service
- JWT-based authentication and protected log endpoints
- Web dashboard for log exploration, log details, and parser management
- Health endpoints for the backend and parser-generation service

## 6. Technology Stack

- **Frontend:** Next.js 16, React 19, Tailwind CSS, shadcn/ui components
- **Backend API and workers:** Go 1.26.5
- **ML / parser generation service:** Python 3.11, FastAPI, Groq API
- **Database:** MongoDB
- **Message broker:** RabbitMQ 3
- **Cache:** Redis 7
- **Authentication:** JWT and Google OAuth client integration
- **Deployment:** Docker Compose, GCP

## 7. Architecture
See `assets/screenshot/architecture.png` for complete architecture.

## 8. Repository Structure

```text
GoLogGo/
├── README.md
├── docker-compose.yml
└── code/
    ├── backend/
    │   ├── cmd/server/              # Go API entrypoint
    │   └── internal/
    │       ├── api/                 # HTTP handlers
    │       ├── auth/                # JWT authentication
    │       ├── database/            # MongoDB access
    │       ├── fingerprint/          # Format detection and hashing
    │       ├── parser/               # Parser execution
    │       ├── parsergen/            # Parser-generation jobs
    │       ├── rabbitmq/             # Message broker integration
    │       ├── redis/                # Cache integration
    │       └── worker/               # Concurrent log processing
    ├── frontend/gologgo-client/      # Next.js web application
    └── ml/
        └── app/                      # FastAPI parser rule synthesizer
```

## 9. Final Presentation

The final SIH presentation is included in the repository.

You can find the presentation in:

`submission/PRESENTATION.pptx`

## 10. Demo Video

The project demonstration video is included in the repository documentation.

The YouTube link is available in:

`submission/DEMO.md`

Direct Link : [Youtube](https://www.youtube.com/watch?v=V8vLanBAm0E)

## 11. Screenshots / Prototype Photos

Important project screenshots and prototype photos are included in:

`assets/screenshots/`

These include relevant screenshots of the project, system interface, and implementation.


## 12. Installation

### Prerequisites

- Docker and Docker Compose
- Go 1.26.5 or later for local backend development
- Node.js and pnpm for local frontend development
- A reachable MongoDB instance
- A Groq API key for LLM-assisted parser generation

Clone the repository and enter its root directory:

```bash
git clone https://github.com/EleventhHour-Projects/GoLogGo.git
cd GoLogGo
```

Create local environment files from the checked-in examples, then replace every placeholder with a real local value:

```bash
cp code/backend/.env.example code/backend/.env
cp code/ml/.env.example code/ml/.env
```

At minimum, configure `MONGODB_URI`, `MONGODB_DATABASE_NAME`, `RABBITMQ_URL`, `REDIS_URL`, `JWT_KEY`, `ML_API_URL`, and `GROQ_API_KEY`. Do not commit `.env` files or credentials.

## 13. Run

### Run services with Docker Compose

From the repository root:

```bash
docker compose up --build
```

The services are available at:

- Frontend: `http://localhost:3000` when started locally
- Backend API: `http://localhost:9000`
- Backend health: `http://localhost:9000/health`
- ML service: `http://localhost:9001`
- ML health: `http://localhost:9001/health`
- RabbitMQ management UI: `http://localhost:15672`

### Run the frontend locally

```bash
cd code/frontend/gologgo-client
pnpm install
pnpm dev
```

The frontend uses `http://localhost:9000` as the default backend URL. Set `BACKEND_URL` and `NEXT_PUBLIC_GOOGLE_CLIENT_ID` in a local frontend `.env` file when needed.

## 14. Future Scope

- Add native ingestion connectors for common security and observability platforms.
- Support schema versioning and parser approval workflows.
- Add richer event correlation, alerting, and operational analytics.
- Auto scaling of worker pool size.
- Fine-tune an LLM for local use without an internet dependency.
- Securely store logs using encryption and blockchain technology.
- Expand parser evaluation datasets and automated regression testing.
