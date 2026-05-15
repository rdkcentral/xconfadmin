# SAT RBAC v2 Design

## Overview

This document describes the architectural approach to implementing SAT RBAC v2 capability names and authorization logic in xconfadmin while maintaining full backward compatibility with legacy SAT behavior.

## Design Principles

1. **Xerxes-First**: If a valid Xerxes token is present, use Xerxes authorization exclusively; SAT is not consulted.
2. **Graceful Fallback**: If SAT contains xconf-prefixed capabilities, authorize via SAT v2; otherwise, fall back to legacy SAT unchanged.
3. **Deny-by-Default for SAT v2**: Any request that cannot be classified into (domain, access) requirements is denied.
4. **Route-Based Classification**: Domain and access level are determined from HTTP method and API route/path, not from request entity type.
5. **Backward Compatibility**: Existing SAT tokens without xconf-prefixed capabilities continue to work exactly as before.

## Authorization Precedence

The authorization system follows this strict precedence:

```
┌─────────────────────────────────┐
│ Request arrives with credentials │
└─────────────────┬───────────────┘
                  │
          ┌───────▼────────┐
          │ Xerxes present?│
          └───────┬────────┘
                  │
        ┌─────────┴─────────┐
       NO                   YES
        │                    │
        │          ┌─────────▼─────────┐
        │          │ Xerxes valid?     │
        │          └─────────┬─────────┘
        │                    │
        │          ┌─────────┴─────────┐
        │         NO                   YES
        │          │                    │
        │          │          ┌─────────▼─────────────┐
        │          │          │ Authorize via Xerxes  │
        │          │          │ (Skip SAT entirely)   │
        │          │          └───────────────────────┘
        │          │
        ▼          ▼
   ┌────────────────────────┐
   │ SAT present and valid? │
   └────────────┬───────────┘
                │
        ┌───────┴────────┐
       NO                YES
        │                 │
        │      ┌──────────▼──────────┐
        │      │ Has xconf: prefix?  │
        │      └──────────┬──────────┘
        │                 │
        │         ┌───────┴──────┐
        │        NO              YES
        │         │               │
        │         │   ┌───────────▼──────────────┐
        │         │   │ Authorize via SAT RBAC v2│
        │         │   │ (domain + access check)  │
        │         │   └──────────────────────────┘
        │         │
        ▼         ▼
   ┌────────────────────────┐
   │ Authorize via legacy   │
   │ SAT (unchanged)        │
   └────────────────────────┘
```

## Request Classification

### Route-to-Domain Mapping

Request domain is determined by matching the HTTP route/path against an ordered ruleset. The first matching rule determines the domain.

Classification rules SHALL match on stable route substrings and/or route templates. The following is a representative seed set of patterns:
- `/queries/firmware`, `/firmware` → **core**
- `/dcm`, `/telemetry` → **core**
- `/tagging` → **tagging**
- `/roundrobinfilter` → **system**
- `/metrics` → **metrics**

The ruleset ordering is critical because the first match wins. More specific routes must appear before more general patterns.

**Note on Mapping Scope**: This seed set represents initial patterns. The authoritative route-to-domain classification will be maintained in a central mapping registry in code and extended iteratively as new endpoints are added. SAT v2 authorization remains deny-by-default for any request that cannot be classified.

### Access-Level Determination

Access level (readonly or readwrite) is determined by the following precedence:

1. **Route Override**: If the request path contains the segment `/filtered`, treat as readonly (these are filtered search endpoints that use POST).
2. **HTTP Method**: Otherwise, classify by HTTP method:
   - GET, HEAD → **readonly**
   - POST, PUT, PATCH, DELETE → **readwrite**

Post-based read endpoints (e.g., /filtered searches) are explicitly classified as readonly via the override pattern to avoid misclassification based on method alone.

## Integration Architecture

### Authorization Middleware Position

The SAT RBAC v2 authorization logic is invoked after credential validation, within the existing auth middleware stack:

```
Request
  │
  ├─> Credential Extraction & Validation
  │     (Xerxes token, SAT token, etc.)
  │
  ├─> Xerxes Authorization (if present)
  │
  ├─> SAT RBAC v2 Authorization (if xconf: prefix detected)
  │     ├─> Route Classification (domain + access)
  │     ├─> Capability Matching
  │     └─> Allow/Deny Decision
  │
  ├─> Legacy SAT Authorization (fallback)
  │
  └─> Handler Execution (if authorized)
```

### Capability Matching

Given a classified request (domain, access), SAT RBAC v2 checks whether the SAT token contains a matching capability:

- Request (core, readonly) requires any of: xconf:core:readonly, xconf:core:readwrite
- Request (core, readwrite) requires: xconf:core:readwrite
- Request (tagging, readonly) requires any of: xconf:tagging:readonly, xconf:tagging:readwrite
- Request (tagging, readwrite) requires: xconf:tagging:readwrite
- Request (system, readonly) requires any of: xconf:system:readonly, xconf:system:readwrite
- Request (system, readwrite) requires: xconf:system:readwrite
- Request (metrics, readonly) requires: xconf:metrics:readonly
- No readwrite functionality exists for the metrics domain (no xconf:metrics:readwrite).

### Unclassifiable Requests

If a request cannot be classified into (domain, access) requirements—for example, if a new API route is added but the classification rules are not yet updated—the request is denied with 403 Forbidden.

This deny-by-default approach ensures that new endpoints are secure by default and prevents accidental authorization leakage.

## Backward Compatibility

Legacy SAT tokens (without xconf-prefixed capabilities) continue to work exactly as before:

1. If no xconf: capabilities are detected, the authorization flow immediately falls back to legacy SAT logic.
2. Legacy SAT behavior (appType-based authorization, existing capability names) is unchanged.
3. No new validation rules are applied to legacy SAT tokens.
4. Operators can mix legacy and SAT v2 tokens in the same deployment; each is authorized according to its own rules.

## HTTP Status Code Semantics

The auth middleware uses the following status codes:

- **401 Unauthorized**: Sent when credential extraction or validation fails (missing token, invalid signature, expired token, etc.). No authenticated identity is available.
- **403 Forbidden**: Sent when the request is authenticated but not authorized for the requested operation. This includes:
  - SAT v2 capability mismatch (e.g., readonly SAT v2 token attempting a write)
  - Unclassifiable SAT v2 requests (route/domain mapping missing)
  - Any other authorization denial after successful authentication

## Future Extensibility

### Phase 2: Tenant/Partner Enforcement
Once domain-to-capability mapping is stable, Phase 2 will add tenant or partner scoping to authorization. Tenant/partner enforcement will be implemented using separate SAT claims (e.g., partner scope) and/or request metadata (e.g., tenantId header), independent of capability strings. The capability strings themselves (e.g., `xconf:core:readwrite`) will remain unchanged.

### Metrics Domain
The metrics domain is read-only. The only supported capability for the metrics domain is `xconf:metrics:readonly`. No write capability exists for the metrics domain.

### Additional Domains
As new domain areas emerge, new capability names (e.g., xconf:foo:readonly) can be added without affecting existing capabilities or fallback logic.
