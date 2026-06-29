# apps/web

React + TypeScript + Vite frontend for the Enterprise File Ingestion Platform.

## Stack

- React 18
- TypeScript
- Vite
- Tailwind CSS
- React Router
- TanStack Query
- Zustand
- Recharts
- TanStack Table

## Folder structure

```text
apps/web/src/
├── app/
├── components/
│   ├── layout/
│   ├── ui/
│   ├── upload/
│   ├── files/
│   ├── jobs/
│   ├── search/
│   └── dashboard/
├── pages/
├── services/
├── hooks/
├── types/
├── stores/
└── utils/
```

## Pages

- Dashboard (`/dashboard`)
- Upload (`/upload`)
- Files (`/files`)
- File detail (`/files/:file_id`)
- Jobs (`/jobs`)
- Job detail (`/jobs/:job_id`)
- Search (`/search`)
- Admin settings (`/admin/settings`)
- Audit logs (`/audit-logs`)

## Development

From repository root:

```bash
cd apps/web
npm install
npm run dev -- --host 0.0.0.0 --port 5173
```

The app expects `VITE_API_BASE_URL` to point to `http://localhost:8080/api/v1`.

To override in this package:

```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1 npm run dev
```

Build production bundle:

```bash
npm run build
```

Preview:

```bash
npm run preview
```

## Compose service

The root `docker-compose.yml` defines `web` and uses `apps/web/Dockerfile`.

Access the UI at:

- `http://localhost:5173`

## Quick API smoke checks from browser or terminal

- List files:

```bash
curl -s http://localhost:8080/api/v1/files | jq
```

- Search parsed records:

```bash
curl -s "http://localhost:8080/api/v1/search?q=error&sort=created_at&page=1&page_size=10" | jq
```

- Search suggestions:

```bash
curl -s "http://localhost:8080/api/v1/search/suggestions?q=10.0.0.1" | jq
```

- Submit archive password:

```bash
curl -s -X POST "http://localhost:8080/api/v1/files/PASTE_FILE_ID_HERE/password" \
  -H "Content-Type: application/json" \
  -d '{"password":"secret"}' | jq
```
