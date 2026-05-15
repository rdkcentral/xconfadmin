# SAT RBAC v2 Specification

## Spec Delta Summary

This change updates `openspec/specs/auth/auth-contract.md` with the
following normative clauses:

- `Authorization Precedence`: Xerxes MUST be evaluated first, SAT RBAC v2
  second, and legacy SAT third.
- `SAT RBAC v2 Detection`: SAT v2 SHALL be detected by presence of at
  least one capability with prefix `xconf:`.
- `SAT RBAC v2 Domain Classification`: domain SHALL be determined by
  ordered route/path rules; first matching rule MUST win.
- `SAT RBAC v2 Access Classification`: `/filtered` route override SHALL
  classify as readonly before method-based mapping.
- `SAT RBAC v2 Deny-By-Default`: unclassifiable SAT v2 requests SHALL be
  denied with `403 Forbidden`.
- `HTTP Status Semantics`: `401 Unauthorized` only for missing/invalid
  authentication; `403 Forbidden` for authenticated authorization denials.
- `Metrics Domain Constraint`: metrics SHALL be readonly-only; no
  `xconf:metrics:readwrite` functionality exists.

## Definitions

### SAT Token
A Structured Authorization Token (SAT) is a credential containing a list of capability strings. The token format is unchanged; this specification only defines new capability names.

### Capability
A capability is an opaque string in the SAT token indicating a permission or privilege. Legacy capabilities are arbitrary strings; SAT v2 capabilities follow the pattern `xconf:<domain>:<access>`.

### Domain
A domain is a logical grouping of xconfadmin APIs. Valid domains are:
- **core**: Firmware rules, firmware configs, firmware templates, features, feature control rules, telemetry profiles, telemetry rules, and DCM settings.
- **tagging**: Tagging APIs.
- **system**: System-level APIs (e.g., download location round robin filter).
- **metrics**: Metrics and analytics APIs.

### Access Level
An access level describes the scope of operations permitted by a capability.
- **readonly**: Only read operations (GET, HEAD queries) are permitted.
- **readwrite**: Both read operations (GET, HEAD) and write operations (POST, PUT, PATCH, DELETE) are permitted.

### SAT v2 Capability Names
The following capability names are defined for SAT RBAC v2:
- `xconf:core:readonly` - Read access to core domain
- `xconf:core:readwrite` - Read and write access to core domain
- `xconf:tagging:readonly` - Read access to tagging domain
- `xconf:tagging:readwrite` - Read and write access to tagging domain
- `xconf:system:readonly` - Read access to system domain
- `xconf:system:readwrite` - Read and write access to system domain
- `xconf:metrics:readonly` - Read access to metrics domain

No readwrite functionality exists for the metrics domain in Phase 1 (no `xconf:metrics:readwrite`).

### SAT v2 Detection
A SAT token is classified as SAT v2 if and only if it contains at least one capability with the prefix `xconf:`.

A SAT token without any xconf-prefixed capabilities is treated as a legacy SAT token and authorized according to legacy SAT semantics.

## Authorization Algorithm

### Input
- HTTP method (GET, POST, HEAD, PUT, PATCH, DELETE)
- HTTP request path (e.g., `/queries/firmware`, `/dcm/devices`)
- Xerxes token (if present)
- SAT token (if present)

### Output
- Authorization decision: ALLOW or DENY
- HTTP status code: 401 Unauthorized (auth failure) or 403 Forbidden (authz failure)

### Precedence

1. **Xerxes Authorization**
   - If Xerxes token is present AND Xerxes validation succeeds:
     - Authorize based on Xerxes permissions (Xerxes-specific logic)
     - If Xerxes permissions allow the operation:
       - ALLOW request
       - Skip all SAT authorization
     - If Xerxes permissions do not allow the operation:
       - DENY request
       - Return 403 Forbidden
   - If Xerxes token is present but validation fails:
     - DENY request
     - Return 401 Unauthorized

2. **SAT RBAC v2 Authorization**
   - If SAT token is present AND valid AND is classified as SAT v2 (has xconf: prefix):
     - Classify request into (domain, access) pair (see below)
     - If request cannot be classified:
       - DENY request
       - Return 403 Forbidden
     - Check SAT token capabilities for matching capability
     - If matching capability exists:
       - ALLOW request
     - If no matching capability:
       - DENY request
       - Return 403 Forbidden
     - Skip legacy SAT authorization

3. **Legacy SAT Authorization**
   - If SAT token is present AND valid AND is NOT classified as SAT v2:
     - Authorize using legacy SAT semantics (unchanged from prior xconfadmin behavior)
   - If legacy SAT authorization allows:
     - ALLOW request
   - If legacy SAT authorization denies:
     - DENY request
     - Return 403 Forbidden

4. **No Credentials**
   - If no Xerxes token and no SAT token:
     - DENY request
     - Return 401 Unauthorized

## Request Classification

### Algorithm

Given an HTTP method and request path, classify the request as (domain, access):

**Step 1: Determine Access Level**

```
if path contains segment "/filtered":
    access = readonly
else if method in {GET, HEAD}:
    access = readonly
else if method in {POST, PUT, PATCH, DELETE}:
    access = readwrite
else:
    # Unknown method; deny for safety
    DENY with 403
```

**Step 2: Determine Domain**

Apply the following ruleset in order. The first matching rule determines the domain:

| Rule # | Path Pattern | Domain | Notes |
|--------|-------------|--------|-------|
| 1 | contains `/metrics` | metrics | Metrics domain |
| 2 | contains `/roundrobinfilter` | system | Round robin filter for download locations |
| 3 | contains `/tagging` | tagging | Tagging APIs |
| 4 | contains `/telemetry` | core | Telemetry profiles and rules |
| 5 | contains `/dcm` | core | Device Configuration Management |
| 6 | contains `/queries/firmware` or `/firmware` | core | Firmware rules and configs |
| 7 | contains `/feature` | core | Feature and feature control rules |
| 8 | _default_ | _unclassified_ | No matching rule |

If the default rule matches (no prior rules matched), the request is unclassifiable and SHALL be DENIED with 403.

### Classification Examples

| HTTP Method | Path | Domain | Access | Capabilities Required |
|-------------|------|--------|--------|----------------------|
| GET | /queries/firmware | core | readonly | xconf:core:readonly, xconf:core:readwrite |
| POST | /queries/firmware/filtered | core | readonly | xconf:core:readonly, xconf:core:readwrite |
| POST | /firmware | core | readwrite | xconf:core:readwrite |
| PUT | /dcm/device-settings | core | readwrite | xconf:core:readwrite |
| GET | /tagging/operations | tagging | readonly | xconf:tagging:readonly, xconf:tagging:readwrite |
| DELETE | /tagging/operations/123 | tagging | readwrite | xconf:tagging:readwrite |
| GET | /roundrobinfilter | system | readonly | xconf:system:readonly, xconf:system:readwrite |
| GET | /metrics/penetration | metrics | readonly | xconf:metrics:readonly |
| POST | /unknown-api | _unclassified_ | - | DENY 403 |

## Capability Matching

After classifying a request as (domain, access), check whether the SAT token contains at least one capability matching the requirement:

**For readonly requests**:
- Required capabilities: `xconf:<domain>:readonly` OR `xconf:<domain>:readwrite`
- Both are acceptable because readwrite implies readonly

**For readwrite requests**:
- Required capability: `xconf:<domain>:readwrite`
- Only readwrite is acceptable; readonly is insufficient

### Matching Examples

| SAT Capabilities | Request | Result |
|------------------|---------|--------|
| `xconf:core:readonly` | GET /queries/firmware | ALLOW |
| `xconf:core:readonly` | POST /firmware | DENY 403 |
| `xconf:core:readwrite` | POST /firmware | ALLOW |
| `xconf:core:readwrite` | GET /queries/firmware | ALLOW |
| `xconf:tagging:readonly, xconf:core:readwrite` | GET /tagging/ops | ALLOW |
| `xconf:tagging:readonly, xconf:core:readwrite` | DELETE /tagging/ops/1 | DENY 403 |
| `xconf:metrics:readonly` | GET /metrics/penetration | ALLOW |
| `xconf:core:readonly` | POST /unknown-api | DENY 403 (unclassifiable) |

## Backward Compatibility

### Legacy SAT Tokens

Tokens without any xconf-prefixed capability are treated as legacy SAT tokens and are authorized using the existing xconfadmin legacy SAT semantics. No new rules or restrictions are applied.

Example: A SAT token with capabilities `["admin", "firmware-operator"]` (without xconf: prefix) is evaluated using legacy logic and is unaffected by SAT v2 authorization.

### Migration Path

1. **Phase 1 (Current)**: Operators issue both legacy SAT tokens and SAT v2 tokens. Legacy tokens work unchanged; SAT v2 tokens use new classification rules.
2. **Phase 2**: Tenant/partner enforcement is introduced using separate SAT claims and/or request metadata (e.g., tenantId header), independent of capability strings.
3. **Phase 3** (future): Legacy SAT tokens are deprecated and eventually removed.

During the transition, no action is required by operators for existing tokens to continue working.

## Error Responses

### 401 Unauthorized
Returned when:
- No credentials (Xerxes or SAT) are provided
- Xerxes token validation fails
- SAT token signature validation fails
- SAT token is expired

Response body SHALL include an error code and message suitable for debugging.

### 403 Forbidden
Returned when:
- Xerxes token is valid but does not grant permission for the requested operation
- SAT RBAC v2 token does not contain a matching capability for the request (domain, access)
- SAT RBAC v2 request cannot be classified into a valid (domain, access) pair
- Legacy SAT authorization denies the request

Response body SHALL include an error code and message suitable for debugging.

## Implementation Notes

- Route classification rules MUST be ordered as specified; the first matching rule determines the domain.
- The `/filtered` segment check MUST be performed before HTTP method inspection to correctly classify POST-based read operations.
- Capability matching MUST be case-sensitive (xconf:core:readwrite is not equivalent to xconf:CORE:READWRITE).
- When multiple domains could match a path (e.g., a path containing both `/telemetry` and `/tagging`), the first rule in the ruleset that matches SHALL determine the domain.
- SAT v2 authorization SHALL NOT alter or inspect the request entity; classification is based solely on HTTP method and route/path.
- The SAT v2 domains are independent of Xerxes entity types; a request to access a "firmware" entity via SAT v2 is classified by route (core domain) regardless of entity semantics.
