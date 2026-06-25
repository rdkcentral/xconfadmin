# SAT RBAC v2 Tenant Enforcement Specification

## Spec Delta Summary

This change updates openspec/specs/auth/auth-contract.md with the following normative clauses:

- SAT RBAC v2 capability authorization remains unchanged and runs first.
- For all SAT RBAC v2 requests, enforce tenant authorization using:
  - request header `tenantId`
  - SAT claim `allowedResources.allowedPartners`
- Authorization SHALL allow only when `allowedPartners` contains `tenantId`.
- Authorization SHALL deny with `403 Forbidden` when:
  - `tenantId` is missing
  - `allowedResources.allowedPartners` is missing
  - `allowedResources.allowedPartners` is empty
  - `allowedPartners` does not contain `tenantId`
- Legacy SAT behavior and login token/IDP behavior remain unchanged.
- Capability strings remain unchanged.

## Definitions

### Request Tenant
The request tenant is the value of HTTP header `tenantId`.

### Allowed Partners Claim
The SAT partner scope claim is `allowedResources.allowedPartners`, represented as a list of tenant/partner identifiers.

Example:

{
  "allowedResources": {
    "allowedPartners": ["comcast"]
  }
}

### Tenant-Scoped SAT Authorization
Tenant-scoped SAT authorization is an authorization gate evaluated after SAT RBAC v2 capability authorization succeeds.

## Normative Behavior

### SAT Authorization Sequence

For SAT RBAC v2 requests, authorization SHALL execute in this order:

1. SAT authentication and SAT RBAC v2 capability authorization.
2. Tenant-scope authorization (this change).

If step 1 fails, deny according to existing SAT RBAC v2 rules.
If step 1 succeeds but step 2 fails, deny with `403 Forbidden`.

### Tenant-Scope Authorization Rule

For all SAT RBAC v2 requests:

1. Read `tenantId` from request header.
2. Read `allowedResources.allowedPartners` from SAT token.
3. Allow only when `allowedPartners` contains the `tenantId` value.

For this phase, the `tenantId` request header value SHALL be compared against the values in `allowedResources.allowedPartners` using case-insensitive membership matching. No separate tenant-to-partner translation is performed.

### Tenant-Scope Failure Outcomes

For all SAT RBAC v2 requests, the system SHALL return `403 Forbidden` when any of the following is true:

- `tenantId` header is missing.
- `allowedResources.allowedPartners` is missing.
- `allowedResources.allowedPartners` is present but empty.
- `allowedResources.allowedPartners` does not contain `tenantId`.

## Compatibility

- SAT RBAC v2 capability checks remain unchanged and run before tenant checks.
- Legacy SAT authorization behavior remains unchanged; token validation
  does not enforce tenant or partner claims. Request processing still
  supports multi-tenancy, and in this phase resolves `tenantId` to the
  default tenant.
- Login token / IDP service behavior remains unchanged; token validation
  does not enforce tenant or partner claims. Request processing still
  supports multi-tenancy, and in this phase resolves `tenantId` to the
  default tenant.
- Capability names remain unchanged; tenant enforcement does not add or modify capability strings.
- Multi-tenant authorization guarantees in this change apply only to
  SAT RBAC v2 requests.

## HTTP Status Semantics

- `401 Unauthorized` is reserved for missing or invalid authentication.
- `403 Forbidden` is used for authenticated authorization denials, including tenant-scope enforcement failures.

## Validation Scenarios

| SAT RBAC v2 capability check | tenantId header | allowedPartners claim | Contains tenantId | Result |
|---|---|---|---|---|
| Fail | Any | Any | Any | Existing deny behavior |
| Pass | Missing | Present | N/A | `403 Forbidden` |
| Pass | Present | Missing | N/A | `403 Forbidden` |
| Pass | Present | Empty | N/A | `403 Forbidden` |
| Pass | Present | Present | No | `403 Forbidden` |
| Pass | Present | Present | Yes | Allow |
