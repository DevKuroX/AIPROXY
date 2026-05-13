# docs/FRONTEND.md — Next.js Dashboard

> **Status:** scaffold.

## Stack
- Next.js 15, App Router
- TypeScript strict
- Tailwind v4
- shadcn/ui components
- Zustand (state) + TanStack Query (server state)
- recharts (charts)

## Build mode
**Static export.** `next.config.mjs` sets `output: 'export'`. The build emits
`web/out/` which is copied to `assets/web/` and embedded into the Go binary
via `//go:embed`. The Go server serves it under `/dashboard/*` (with index
fallback for client-side routing).

## Routes mirror 9router

| 9router page | This dashboard |
|---|---|
| `/` (overview) | `/(dashboard)` |
| `/providers` | `/(dashboard)/providers` |
| `/provider-nodes` | `/(dashboard)/provider-nodes` |
| `/keys` | `/(dashboard)/keys` |
| `/combos` | `/(dashboard)/combos` |
| `/aliases` | `/(dashboard)/aliases` |
| `/pricing` | `/(dashboard)/pricing` |
| `/usage` | `/(dashboard)/usage` |
| `/endpoint` | `/(dashboard)/endpoint` |
| `/logs` | `/(dashboard)/logs` |
| `/settings` | `/(dashboard)/settings` |
| `/login` | `/(auth)/login` |

## Auth
- `(auth)/login` posts to `/api/auth/login`; on success the cookie is set
  by the server and the page redirects to `/`.
- `web/middleware.ts` checks the cookie presence for `(dashboard)/*` group
  and redirects to `/login` if absent.

## API client
- `web/lib/api.ts` is a thin `fetch` wrapper that:
  - Always uses `credentials: 'include'`.
  - Returns typed responses using generated types in `web/lib/types.gen.ts`.
  - Surfaces `errs.OpenAIError` envelopes as typed throws.
- TanStack Query hooks live in `web/hooks/` (one file per resource).

## Codegen
- Source of truth: `docs/openapi.yaml`.
- `task gen` runs `openapi-typescript docs/openapi.yaml -o web/lib/types.gen.ts`.
- **Never** hand-edit `types.gen.ts`.

## Styling
- shadcn/ui generated under `web/components/ui/`.
- Tailwind config at `web/tailwind.config.ts`.
- Dark mode via class strategy; toggle stored in `themeStore` (Zustand) +
  cookie for SSR consistency.

## Charts
- recharts only. No d3 direct usage.
- Color palette matches 9router (see `src/app/.../charts` for reference).

## Dev workflow
- `cd web && npm run dev` runs Next.js on a different port (3000 by default).
- API requests go to `http://localhost:20128` via `next.config.mjs` rewrites.
- `task dev` runs both concurrently.

## Hard rules
- ❌ No Next.js API routes (`web/app/api/`). Go is the only backend.
- ❌ No `getServerSideProps` / `Route Handlers` for data. Use TanStack Query.
- ❌ No CSS-in-JS. Tailwind + CSS modules only.
