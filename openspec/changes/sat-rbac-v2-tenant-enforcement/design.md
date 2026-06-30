# SAT RBAC v2 Tenant Enforcement Design

## Overview

This document defines tenant-aware SAT authorization for xconfadmin as a follow-on phase to SAT RBAC v2 capability authorization. The goal is to enforce partner/tenant scope for SAT-authenticated requests in multi-tenant deployments.

Capability authorization remains the first gate. Tenant enforcement is a second gate that runs only after SAT RBAC v2 capability checks succeed.

## Design Principles

1. Keep credential-path routing unchanged.
2. Keep SAT RBAC v2 capability names and capability matching unchanged.
3. Enforce tenant scope for all SAT RBAC v2 requests.
4. Deny with `403 Forbidden` for all tenant-enforcement failures.
5. Preserve legacy SAT and login-token behavior.

## Authorization Flow

1. Choose auth path
- If `Authorization` header is present, select SAT path.
- Else, if login token (`token` header/cookie) is present, select login/Xerxes path.
- Else, return `401 Unauthorized`.

2. Validate credentials
- SAT path: validate SAT token.
- Login/Xerxes path: validate login token as currently implemented.

3. Detect SAT mode on SAT path
- If SAT includes at least one capability with prefix `xconf:`, classify request as SAT RBAC v2.
- Otherwise, use legacy SAT behavior unchanged and stop this design flow.

4. Perform SAT RBAC v2 capability authorization
- Classify request to `(domain, access)`.
- Evaluate SAT capabilities.
- If capability authorization fails, return `403 Forbidden`.

5. Enforce SAT tenant scope (new phase)
- For all SAT RBAC v2 requests:
  - Read request tenant from header `tenantId`.
  - Read allowed partner list from SAT claim `allowedResources.allowedPartners`.
  - Allow only when `allowedPartners` contains `tenantId`.

6. Complete request
- If tenant scope check passes, continue to handler.
- If tenant scope check fails, return `403 Forbidden` and terminate.

## Tenant Enforcement Decision Table

| Condition | Result |
|---|---|
| `tenantId` header missing | `403 Forbidden` |
| `allowedResources.allowedPartners` missing | `403 Forbidden` |
| `allowedResources.allowedPartners` empty | `403 Forbidden` |
| `tenantId` not found in `allowedPartners` | `403 Forbidden` |
| `tenantId` found in `allowedPartners` | Allow |

* If the token validator enforces allowedResources.allowedPartners as a required claim (per auth-contract.md SAT Token Validation Requirements), missing/empty values result in 401 Unauthorized during token validation and do not reach this phase.

## Error Semantics

- `401 Unauthorized` remains only for missing/invalid authentication.
- `403 Forbidden` is used for authenticated authorization failures, including:
  - SAT capability denials
  - SAT tenant-scope denials

## Compatibility

- SAT RBAC v2 capability checks remain unchanged and execute first.
- Legacy SAT behavior remains unchanged.
- Login token / IDP service behavior remains unchanged.
- Capability strings remain unchanged; tenant enforcement uses `tenantId` plus SAT claim values.

## Implementation Notes

- Tenant enforcement should be implemented in SAT RBAC v2 authorization flow after capability success and before handler execution.
- Tenant enforcement logic should be isolated and reusable to minimize risk to existing capability validation code paths.
- Failure responses for tenant checks should be explicit authorization failures (`403`) and should keep fail-fast behavior.
