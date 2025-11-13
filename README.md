# Community Report Worker

A modular Go-based worker for generating PDF reports and processing images, designed for community organizations, educational projects, or any portfolio use. Easily switch between local, Docker Compose, or cloud infrastructure.

---

## Features
- **PDF Report Generation**: Community activity, participant demographics, program impact, financial summary (with charts)
- **Image Processing**: Resize and convert images to WebP
- **Storage Abstraction**: Save files locally or to Cloudflare R2 (S3-compatible)
- **MongoDB & Redis Integration**: Flexible connection (local, Docker, or cloud)
- **Configurable Organization Name**: Set via `.env`, appears in all reports
- **Docker & Compose Ready**: Easy deployment and local development

---

## Quick Start

### 1. Clone the Repository
```sh
git clone https://github.com/padiil/Community-Report-Worker.git
cd Community-Report-Worker
```

### 2. Environment Setup
Copy `.env.example` to `.env` and fill in as needed:
```sh
cp .env.example .env
```

**Key variables:**
- `ORG_NAME` — Organization name for reports (default: "Community Organization")
- `REDIS_URI` — Redis connection string (e.g. `localhost:6379` or `queue:6379` for Compose)
- `MONGO_URI` — MongoDB connection string (e.g. `mongodb://localhost:27017/<db>` or `mongodb://db:27017/<db>` for Compose)
- `STORAGE_PROVIDER` — `local` (default) or `r2` for Cloudflare R2
- R2 credentials (if using cloud storage)

### 3. Run with Docker Compose (Recommended)
```sh
docker compose up --build
```
- Default services: `worker`, `db` (MongoDB), `queue` (Redis)
- All connections are managed via environment variables

### 4. Run Locally (Without Docker)
- Start MongoDB and Redis on your machine
- Set `REDIS_URI` and `MONGO_URI` in `.env` to point to your local services
- Build and run:
```sh
go build -o worker ./cmd/worker
./worker
```

### 5. Run with Cloud Services
- Set `REDIS_URI` and `MONGO_URI` to your cloud endpoints in `.env`
- For Cloudflare R2, set all R2 credentials and `STORAGE_PROVIDER=r2`

---

## Usage

### Tracking Document Pattern
This project uses the **Tracking Document Pattern** for job dispatch:
- **Step 1:** Insert a job document into MongoDB (`reports` or `image_jobs` collection).
- **Step 2:** Push only the job ID to Redis (`task_queue`).
- **Step 3:** Worker fetches job details from MongoDB, processes, and updates status/output fields.

### Enqueue Report Generation (Recommended)
1. **Insert job document to MongoDB (collection: reports):**
```json
{
  "_id": { "$oid": "655500a1f12a3d0f3c5a1001" },
  "type": "community_activity",
  "status": "pending",
  "fileURL": "",
  "errorMsg": "",
  "filters": {
    "community_name": "Community A",
    "start_date": "2025-01-01T00:00:00Z",
    "end_date": "2025-01-31T23:59:59Z"
  },
  "createdAt": { "$date": "2025-11-13T14:00:00.000Z" },
  "updatedAt": { "$date": "2025-11-13T14:00:00.000Z" }
}
```
2. **Push job ID to Redis:**
```sh
redis-cli
> LPUSH task_queue '{"task_type":"generate_report","payload":{"reportID":"655500a1f12a3d0f3c5a1001"}}'
```

### Enqueue Image Processing (Recommended)
1. **Insert job document to MongoDB (collection: image_jobs):**
```json
{
  "_id": { "$oid": "655501d2f12a3d0f3c5a1002" },
  "status": "PENDING",
  "sourceImageURL": "https://your-bucket.r2.dev/uploads/originals/test-image.jpeg",
  "outputImageURL": "",
  "errorMsg": "",
  "createdAt": { "$date": "2025-11-13T15:00:00.000Z" },
  "updatedAt": { "$date": "2025-11-13T15:00:00.000Z" }
}
```
2. **Push job ID to Redis:**
```sh
redis-cli
> LPUSH task_queue '{"task_type":"process_image","payload":{"imageJobID":"655501d2f12a3d0f3c5a1002"}}'
```

---

## Folder Structure
```
cmd/worker/                # Main worker entrypoint
internal/config/           # Configuration and environment loading
internal/domain/           # Domain models and types
internal/processor/image/  # Image processing logic
internal/processor/report/ # PDF report generation logic
internal/queue/            # Redis queue helpers
internal/repository/       # MongoDB data access
internal/storage/          # Storage abstraction (local/R2)
reports/                   # Output folder for generated files
```

---

## Environment Variables
See `.env.example` for all options. Key variables:
- `ORG_NAME` — Organization name for branding (optional)
- `REDIS_URI` — Redis connection string
- `MONGO_URI` — MongoDB connection string
- `STORAGE_PROVIDER` — `local` or `r2`
- `R2_ENDPOINT`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET_NAME`, `R2_PUBLIC_URL` — Cloudflare R2 credentials

---

## Customization & Branding
- Change `ORG_NAME` in `.env` to set your organization name in all reports
- No hardcoded branding—fully portfolio-friendly

---

## Troubleshooting
- **MongoDB/Redis connection errors**: Check your `.env` and service status
- **File not found**: Ensure paths are correct and accessible
- **Cloudflare R2 errors**: Double-check credentials and bucket permissions

---
