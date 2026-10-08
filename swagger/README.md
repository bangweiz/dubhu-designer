# OpenAPI specification

Use `openapi.yaml` as the entry point.

- `paths/`: endpoint definitions grouped by resource.
- `schemas/`: request and response schemas grouped by resource, plus shared error schemas in `common.yaml`.

The entry point controls endpoint and tag ordering. Variables and environments appear last.

## Authentication and organisations

1. `POST /api/v1/organisations` with organisation `name`, `description`, and `rootAccount: {name, email, password}`. Registration atomically creates the organisation and its sole root account.
2. `POST /api/v1/organisations/{organisationId}/auth/login` with `email` and `password` to obtain `accessToken` and `expiresAt`.
3. Send `Authorization: Bearer <accessToken>` on organisation-scoped API calls. Tokens expire after 24 hours. `POST /auth/logout` within the organisation prefix revokes the current token; `GET /auth/me` returns the current account.
4. Only root may `POST /api/v1/organisations/{organisationId}/accounts` with `{name, email, password, role}`. Role must be `admin` or `user`.

All existing resource endpoints are now under `/api/v1/organisations/{organisationId}`; the old unscoped URLs are not registered. A deny-by-default permission middleware protects the entire API group, including future endpoints. Root/admin can write and every role can read. Registration and login are explicitly public; every authenticated role can revoke its own session. Names of resources and account emails are unique within an organisation. Organisation names are globally unique.

Passwords must have at least 15 characters and at most 72 UTF-8 bytes. They are hashed with bcrypt (cost 12), never returned, and never included in validation errors. Session tokens contain 256 random bits; only their SHA-256 hashes are stored. Database queries and aggregation lookups require authenticated organisation context and fail closed without it.

The application creates organisation-scoped indexes and replaces legacy global unique-name indexes at startup. Existing documents without `organisation_id` remain stored but inaccessible. They require an explicit ownership migration; they are not automatically assigned to a new organisation.

Deploy behind HTTPS. The built-in authentication limit is 30 requests/minute/IP per process; multi-instance deployments need a shared rate limiter at the gateway. Proxy headers are not trusted by default; configure an explicit trusted-proxy allowlist before relying on forwarded client IPs. Public organisation registration is intentional onboarding; restrict it at the gateway if registration should be invitation-only. Password reset, email verification, MFA, and account lifecycle management are not implemented in this initial scope.

The Postman collection includes authentication requests. Set `rootPassword` and `accountPassword` locally (minimum 15 characters), then create an organisation and log in. Organisation ID and access token are captured automatically; do not commit real passwords or tokens.

Keep this directory structure when loading or publishing the specification: relative `$ref` references require the associated files to be available. Serve the entire `swagger` directory for browser-based documentation tools; uploading only `openapi.yaml` will not include its referenced files.

## Editable concierges and immutable versions

A concierge stores its sole editable configuration, including `agents`, directly. Agent endpoints use `/concierges/{conciergeId}/agents`; no draft ID or `draft` URL selector exists. `GET /concierges/{conciergeId}` returns agent summaries, `nextVersion`, and `conciergeVersions` containing only saved snapshot references. Use agent get-by-ID to resolve current instruction and tool assignments.

`POST /concierge/{conciergeId}:save` snapshots the concierge name, description, agents, and referenced instructions/tools in a transaction. It increments `nextVersion` and appends the saved ID to `conciergeVersions`. Saved versions expose only `GET /concierges/{conciergeId}/concierge-versions/{conciergeVersionId}`; they have no nested agent or mutation endpoints. The former `saved` flag is unnecessary and removed. Concierge ETags are derived from `updatedAt` and change on metadata updates, agent mutations, and saves independently of snapshot sequence numbers. Usage responses identify the current concierge and agent without a draft version ID or number.

### Migrating existing drafts

Before deploying this change, back up the database and stop application writes. Run `mongosh "$MONGODB_URI" --file scripts/migrate-concierge-drafts.js` from the repository root, selecting the application database in the URI. A replica set is required. The script transfers each legacy draft's agents and version sequence into its parent concierge and removes its draft reference. Saved snapshot documents remain unchanged. It preserves the legacy draft collection for rollback and skips already migrated parents on reruns. Existing saved snapshots predate the name/description fields and may omit them.

The old version-scoped agent URLs are no longer registered. In Postman, set `conciergeVersionId` to an ID from a save response or `conciergeVersions`; it is only used for reading saved snapshots.

## Timestamp-based optimistic updates

Editable resources have no numeric revision field. Their strong `ETag` is the quoted UTC `updatedAt` timestamp, for example `"2026-10-08T00:00:00.123Z"`. Send the returned header unchanged as `If-Match` on updates. Old numeric ETags are invalid; fetch the resource again before updating. Invalid or stale timestamps return HTTP 412.

Writes compare the persisted timestamp atomically and advance it by at least one millisecond, including embedded-agent mutations. Agent responses include `createdAt` and `updatedAt`. Saved concierge `version`, saved references, and `nextVersion` remain snapshot sequence numbers, not optimistic revision fields. Legacy numeric revision properties in existing documents are ignored; no database migration is required for timestamp concurrency.
