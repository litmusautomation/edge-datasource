# Topic Autocomplete Design

**Date:** 2026-03-24
**Status:** Approved

## Overview

Add search-as-you-type autocomplete to the topic field in the query editor. When the user types, the plugin queries Litmus Edge's DeviceHub GraphQL API and suggests matching NATS topics. The feature is optional — it activates only when an API token is configured.

## Requirements

- Query the `ListRegistersFromAllDevices` GraphQL endpoint on Litmus Edge
- Filter results to topics with `Format == "Raw"` only
- Search-as-you-type with 300ms debounce, minimum 2 characters
- Allow free-form entry — user can type any topic, not just suggestions
- API token is optional; autocomplete degrades gracefully without it
- Contextual UX messages guide the user when the token is missing, invalid, or the API is unreachable

## Approach: Backend Resource Handler

The backend proxies GraphQL requests to Litmus Edge. The frontend calls `getResource()` to fetch suggestions. This keeps the API token server-side (secure), handles self-signed TLS certs, and follows the standard Grafana plugin pattern.

## Design

### 1. Configuration (Config Editor)

**New secure field: "API Token"**

- Stored in `secureJsonData.apiToken` (encrypted by Grafana)
- Optional — when not set, autocomplete is disabled but the plugin works normally
- Help text: explains that setting this token enables topic autocomplete via the DeviceHub API
- Placed below the existing Token field in the datasource settings page

**Types:**

```typescript
interface EdgeSecureJsonData {
  token?: string;      // existing — NATS proxy token
  apiToken?: string;   // new — DeviceHub GraphQL API token
}
```

The GraphQL endpoint URL is derived from the existing `hostname` field: `https://<hostname>/devicehub/v2`. No new jsonData fields needed.

### 2. Backend Resource Handler (Go)

**Endpoint:** `GET /topics?query=<search_term>`

**Flow:**

1. Frontend calls `datasource.getResource("topics", { query: "temp" })`
2. Backend `CallResourceHandler` receives the request
3. If `apiToken` is not configured, returns an appropriate error/empty result
4. Builds the GraphQL POST to `https://<hostname>/devicehub/v2`:
   - `operationName`: `"ListRegistersFromAllDevices"`
   - `variables.input.TagPattern`: search term from query param
   - `variables.input.TagPatternSearchOption`: `"CONTAINS"`
   - `variables.input.Limit`: `15`
5. Auth header: `Authorization: Basic base64(":" + apiToken)`
6. TLS: skip verify (self-signed certs, consistent with existing NATS client)
7. Parses GraphQL response, iterates registers, collects topics where `Format == "Raw"`
8. Returns JSON: `{ "topics": ["devicehub.alias.demo.bearing_temperature", ...] }`

**New files:**

- `pkg/plugin/resource.go` — resource handler implementation
- `pkg/edge/devicehub.go` — DeviceHub GraphQL HTTP client

**Changed files:**

- `pkg/plugin/datasource.go` — store `apiToken`, register `CallResourceHandler`

### 3. Frontend Query Editor (React)

**Component:** Replace `<Input>` with Grafana's `<AsyncSelect>` (creatable mode) from `@grafana/ui`.

- **Creatable:** allows free-form entry alongside suggestions
- **Async:** fetches suggestions from the backend resource endpoint
- **Debounce:** 300ms delay before firing requests
- **Min chars:** 2 characters before triggering search

**Contextual inline messages** (using Grafana `<Alert>` inline, small):

| State | Style | Message |
|-------|-------|---------|
| No API token configured | Info (blue) | "Set up an API Token in datasource settings to enable topic autocomplete" (links to config) |
| Token invalid / 401 | Warning (amber) | "Topic autocomplete unavailable — API token may be invalid or expired" |
| API unreachable | Warning (amber) | "Could not reach Litmus Edge API — autocomplete disabled" |
| No results | Neutral (gray) | "No matching topics found — you can still enter a topic manually" |
| Working normally | None | Clean autocomplete, no noise |

**Behavior:**

- On mount, a lightweight probe call (`getResource("topics?query=")`) detects token status
- Graceful degradation: field always works as normal text input when autocomplete is unavailable
- Messages dismiss automatically when a valid token is configured and detected
- Existing validation rules (no wildcards, valid dot-separated tokens) still apply

**Changed files:**

- `src/components/QueryEditor.tsx` — swap `<Input>` for `<AsyncSelect>` creatable + inline alerts
- `src/datasource.ts` — add `getTopics(query: string)` method calling `getResource`

## File Summary

| File | Action | Purpose |
|------|--------|---------|
| `pkg/plugin/resource.go` | New | Resource handler for `/topics` endpoint |
| `pkg/edge/devicehub.go` | New | DeviceHub GraphQL HTTP client |
| `pkg/plugin/datasource.go` | Modify | Store apiToken, register resource handler |
| `src/components/QueryEditor.tsx` | Modify | AsyncSelect + contextual messages |
| `src/datasource.ts` | Modify | Add `getTopics()` method |
| `src/components/ConfigEditor.tsx` | Modify | Add optional API Token field |
| `src/types.ts` | Modify | Add `apiToken` to `EdgeSecureJsonData` |
