# NAStree

A WizTree/WinDirStat-style disk usage visualizer for a NAS, built to run in Docker
and serve a web dashboard — so you can see what's eating your storage from any
browser instead of installing a Windows-only app on the NAS itself.

- **Backend**: Go. Walks a mounted path on an interval (or on demand), stores
  results in SQLite, serves a small JSON API.
- **Frontend**: Vue 3 + ECharts. A WizTree-style nested treemap, a pie chart,
  a scan-wide file search with duplicate detection, and the usual folder/file-type
  tables — all navigable by clicking into folders.

## Features

**Three visualizations, one click apart** — a toolbar next to the theme picker
switches between:
- **Treemap** — a dense, recursively-nested mosaic (files as small leaf boxes,
  folders as stacked header bands), colored by directory vs. file extension,
  with a subtle gradient and a white hover outline.
- **Pie chart** — the same folder's contents as a donut chart; small slices
  automatically fold into "Other" so one huge file doesn't turn everything
  else into an unreadable sliver.
- **File view** — a flat, scan-wide file search (not limited to the folder
  you're in), styled after WizTree's File View: search by filename or full
  path, a folders-only filter, and a **duplicate finder** — match by name+size
  or name+size+date, with a "duplicates only" toggle and Dup Count/Dup Size
  columns.

A **"Treemap / Pie View" toggle** in the top bar swaps the whole content area
for a maximized version of whichever visualization is active, for when you
want the chart to fill the screen.

**Navigation** — an expandable/collapsible directory tree sidebar (stays in
sync no matter which view you navigate from), a breadcrumb, an Up button, and
sortable columns on every table (click a header to sort, click again to flip
direction).

**Themes** — six built in (Dark, Light, High Contrast, VS Code, Monokai,
Solarized), picked from a dropdown and persisted locally. Every visualization
(treemap, pie, swatches) recolors live when you switch.

**Custom colors** — click any color swatch in the file-type table (including
the dedicated "Folders" row) to open a native color picker and override that
extension's or folders' color everywhere it's drawn. Choices persist across
reloads, with per-item and reset-all controls.

**Background scanning** — an initial scan runs on container startup, then
again on an interval (`SCAN_INTERVAL`), or on demand via the "Rescan" button.

## Run it

### Option A: build from source (dev machine)

Edit `docker-compose.yml` and point the volume at the path you want scanned:

```yaml
volumes:
  - /path/to/your/nas/share:/data:ro
  - nastree-db:/db
```

Then:

```sh
docker compose up -d --build
```

### Option B: pull the published image (NAS)

Every push to `main` builds a multi-arch (amd64 + arm64) image and publishes
it to GitHub Container Registry via `.github/workflows/docker-publish.yml`.

On the NAS, copy just `docker-compose.nas.yml` (no source checkout needed),
edit its volume path, then:

```sh
docker compose -f docker-compose.nas.yml up -d
```

To pull a newer version later: `docker compose -f docker-compose.nas.yml pull && docker compose -f docker-compose.nas.yml up -d`.

Open `http://<host>:8080` either way. The first scan starts immediately and
can take a while on a large share; the dashboard shows a "no scan yet"
message until it finishes.

## Configuration (environment variables)

| Variable | Default | Meaning |
|---|---|---|
| `SCAN_PATH` | `/data` | Path inside the container to scan (mount your NAS share here) |
| `SCAN_INTERVAL` | `6h` | How often to rescan in the background (Go duration string) |
| `RETAIN_SCANS` | `5` | How many past scans to keep in the DB (for future history/trend views) |
| `DB_PATH` | `/db/nastree.db` | SQLite database location |
| `LISTEN_ADDR` | `:8080` | HTTP listen address |

A manual rescan can also be triggered from the dashboard's "Rescan" button,
which hits `POST /api/scan/trigger`.

## Project layout

```
backend/    Go module: scanner, SQLite store, HTTP API, scan scheduler
frontend/   Vue 3 + Vite dashboard
docker/     Dockerfile (multi-stage: build frontend + backend, alpine runtime)
```

## Local development

Backend (requires Go 1.25+; not needed if you only use Docker):

```sh
cd backend
SCAN_PATH=/some/local/path go run ./cmd/server
```

Frontend (dev server proxies `/api` to `localhost:8080`, see `vite.config.ts`):

```sh
cd frontend
npm install
npm run dev
```
