Status: Proposed
Applied to: openspec/specs/auth/auth-contract.md

## Why

xconfadmin now supports SAT RBAC v2 capability authorization, but capability checks alone are not sufficient for multi-tenant protection. A SAT token may have the correct domain/access capability while still being scoped only to specific partners. Without tenant-aware enforcement, a SAT-authenticated request can be authorized at capability level without proving access to the request tenant.

This change introduces tenant-aware SAT authorization as a separate phase after the base SAT RBAC v2 rollout. It adds explicit enforcement using request header tenantId and SAT token claim allowedResources.allowedPartners.

## What Changes

- Define tenant source for SAT tenant enforcement:
  - Request header `tenantId`
- Define SAT partner scope source:
  - SAT claim `allowedResources.allowedPartners`
- Define SAT tenant authorization rule (applies after SAT RBAC v2 capability success):
  - Read `tenantId` from request header.
  - Read partner allowlist from `allowedResources.allowedPartners`.
  - Allow only when `allowedPartners` contains `tenantId`.
- Define SAT tenant failure behavior:
  - If `tenantId` is missing when tenant-scoped SAT authorization is required, return `403 Forbidden`.
  - If `allowedResources.allowedPartners` is missing or empty for a SAT request that requires tenant enforcement, return `403 Forbidden`.
  - If `allowedPartners` does not contain `tenantId`, return `403 Forbidden`.
- Clarify compatibility and sequencing:
  - Existing SAT RBAC v2 capability checks remain unchanged and still run first.
  - Tenant enforcement runs only after capability authorization succeeds.
  - Legacy SAT behavior remains unchanged.
  - Login token / IDP service behavior remains unchanged.
  - Capability strings remain unchanged.

## Non-Goals

- No changes to SAT RBAC v2 capability names or capability detection.
- No changes to legacy SAT authorization behavior.
- No changes to Xerxes/login-token/IDP service behavior.
- No endpoint contract changes beyond documented `403` authorization outcomes for tenant-scope denials.

## Capabilities

### Modified Capabilities
- `auth`: Adds SAT RBAC v2 tenant-aware authorization enforcement using request header `tenantId` and SAT claim `allowedResources.allowedPartners` after capability authorization succeeds.

## Impact

- Affected specs: openspec/specs/auth/auth-contract.md.
- Affected implementation areas (future apply phase): SAT authorization middleware and SAT claim extraction/validation flow.
- API behavior impact: SAT RBAC v2 requests now return `403 Forbidden` when tenant scope checks fail.
- Compatibility impact: legacy SAT and login token flows are unchanged.
