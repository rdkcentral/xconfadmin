Status: Proposed
Applied to: openspec/specs/auth/auth-contract.md

## Why

xconfadmin must introduce SAT RBAC v2 capability names while preserving all legacy SAT behavior for backward compatibility. SAT token shape is unchanged (capabilities list), but capability values now include a new xconf-prefixed namespace. We need a clear and deterministic routing-based authorization selection contract that supports SAT v2 when present on the SAT path, falls back to legacy SAT semantics on the SAT path when no xconf capability is present, and uses Xerxes authorization on its own path.

This change documents the transition contract so implementation can safely evolve without breaking existing SAT clients.

## What Changes

- Define SAT RBAC v2 capability namespace and names:
  - xconf:core:readonly / xconf:core:readwrite
  - xconf:tagging:readonly / xconf:tagging:readwrite
  - xconf:system:readonly / xconf:system:readwrite
  - xconf:metrics:readonly
- Define SAT RBAC v2 detection:
  - SAT v2 SHALL be detected by the presence of at least one capability with prefix "xconf:"
- Define routing-based authorization selection:
  - If `Authorization` header is present:
    - Run existing SAT validation logic.
    - If SAT is valid:
      - If SAT contains any capability starting with "xconf:", authorize using SAT RBAC v2.
      - Else, authorize using legacy SAT behavior unchanged.
    - Else, return 401 Unauthorized.
  - Else, if token header/cookie `token` is present:
    - Run existing Xerxes validation and authorization.
  - Else, return 401 Unauthorized.
- Define initial SAT v2 domain mapping seed set:
  - Core: firmware, firmware rules, firmware templates, features, feature rules, telemetry, dcm
  - Metrics: penetration metrics, future metrics APIs
  - System: download location round robin filter
  - Tagging: all tagging APIs
- Define SAT v2 request classification:
  - SAT v2 authorization SHALL classify requests by API route/path into one of {core, tagging, system, metrics}.
  - SAT v2 domains are not equivalent to Xerxes entity types; classification is based on admin functionality (route/path), not entity.
  - Classification SHALL use an ordered ruleset where the first matching rule determines the domain.
  - Classification rules SHALL match on stable route substrings and/or route templates (when available) (e.g., /queries/firmware, /firmware, /dcm, /telemetry, /tagging, /roundrobinfilter, /metrics), with precedence determined by rule order.
- Define SAT v2 access classification:
  - SAT v2 access level SHALL be determined using the following precedence:
    1. If the route matches a known read-only override pattern (e.g., path contains the segment "/filtered"), access SHALL be treated as "readonly".
    2. Else, HTTP method SHALL determine access:
       - GET, HEAD → readonly
       - POST, PUT, PATCH, DELETE → readwrite
  - Certain endpoints use POST for read operations (e.g., filtered search APIs). These MUST be explicitly classified as readonly.
- Define SAT v2 deny-by-default behavior:
  - If a request cannot be classified into (domain, access) requirements, SAT v2 authorization SHALL deny with 403.
- Define HTTP status semantics (Option A):
  - 401 Unauthorized only for missing/invalid authentication.
  - 403 Forbidden for authenticated-but-not-authorized requests, including unmapped SAT v2 operations/domains.

## Non-Goals

- No tenant or partner enforcement in this phase (phase 2).
- No changes to legacy SAT capability names or authorization semantics.
- No appType field in SAT RBAC v2.
- No redesign of Xerxes authentication or permission model.

## Capabilities

### Modified Capabilities
- `auth`: Authorization outcome semantics clarified to use 401 for authentication failures and 403 for authorization denials, including SAT v2 unmapped operations.
- `auth`: SAT RBAC v2 capability-name model, detection by `xconf:` prefix, ordered route/path classification, access classification, and routing-based selection contract across SAT and Xerxes credential paths.

## Impact

- Affected specs: openspec/specs/auth/auth-contract.md.
- Affected implementation areas (future apply phase): SAT validation/authorization middleware and API domain-to-capability mapping logic.
- API behavior impact: no endpoint shape changes; authorization outcomes become explicitly specified for SAT v2 capability evaluation.
- Compatibility impact: legacy SAT behavior remains unchanged unless SAT v2 capabilities are explicitly present.
