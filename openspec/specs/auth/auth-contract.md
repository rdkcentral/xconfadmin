# Authentication Contract (xconfadmin)

## Purpose
This document defines the authentication behavior provided by
xconfadmin as a standalone library.

## Scope
This specification describes:
- credential validation
- identity resolution
- authentication success and failure outcomes
- routing-based authorization selection across SAT v2, legacy SAT, and Xerxes
- SAT v2 request classification requirements
- fail-fast termination after authentication or authorization failure

This specification does not describe:
- business-specific policy enforcement
- downstream tenant or partner enforcement policy outside SAT RBAC v2
- downstream extensions or constraints

## Guarantees

### Credential Validation
The system SHALL validate supplied credentials and determine
their validity deterministically.

### SAT Token Validation Requirements

For SAT token authentication, specific claims may be required for a token
to be considered valid.

Normative behavior:

- The system SHALL validate SAT tokens for structure, signature, and
  required claims.
- Required claims MAY include `allowedResources.allowedPartners`,
as enforced by the SAT token validation implementation.
- If required claims are missing or invalid, the token SHALL be rejected.
- Such failures SHALL result in `401 Unauthorized`.

This validation occurs during authentication and is independent of
SAT RBAC v2 authorization semantics.

Note: Missing or invalid `allowedResources.allowedPartners` MAY be treated
as an authentication failure during SAT token validation and result in
`401 Unauthorized`, depending on validator behavior.

### Authentication Result
On successful authentication, the system SHALL return an
identity representation suitable for downstream use.

### Failure Modes
Authentication failures SHALL result in defined error categories.
Failure handling is subject to the Fail-Fast Termination guarantee
defined below.

### Authorization Routing Selection

Authorization selection SHALL be deterministic and route credentials by
credential type, not by evaluating Xerxes and SAT in precedence order.

Normative behavior:

- If the `Authorization` header is present, the request SHALL be treated
	as SAT-authenticated and SAT processing SHALL be selected.
	- If SAT contains at least one capability with prefix `xconf:`, the
		system SHALL authorize using SAT RBAC v2 semantics.
	- If SAT does not contain any capability with prefix `xconf:`, the
		system SHALL authorize using legacy SAT behavior unchanged.
- Else, if a Xerxes token is present (header `token` or cookie `token`),
	the system SHALL authorize using Xerxes permissions.
- Else, the system SHALL return `401 Unauthorized`.

When both SAT and Xerxes credentials are present, `Authorization`-header
routing MUST win; SAT selection SHALL be used and Xerxes SHALL NOT be
evaluated for that request.

### SAT RBAC v2 Detection

SAT RBAC v2 SHALL be detected by the presence of at least one SAT
capability string with prefix `xconf:`.

SAT tokens without any `xconf:` capability SHALL be treated as legacy SAT.

### SAT RBAC v2 Domain Classification

For SAT RBAC v2 authorization, request classification SHALL be based on
API route/path (admin functionality), not entity type.

SAT RBAC v2 domains are `core`, `tagging`, `system`, and `metrics`.

Domain classification requirements:

- The system SHALL classify requests using an ordered ruleset.
- The first matching rule MUST determine the domain.
- Rules SHALL match on stable route substrings and/or route templates
	(when available), for example: `/queries/firmware`, `/firmware`,
	`/dcm`, `/telemetry`, `/tagging`, `/roundrobinfilter`, `/metrics`.

### SAT RBAC v2 Access Classification

For SAT RBAC v2 authorization, access level SHALL be determined with this
precedence:

1. Route override
2. HTTP method

Normative behavior:

- If the path contains the segment `/filtered`, access SHALL be
	`readonly`.
- Otherwise, access SHALL be method-based:
	- `GET`, `HEAD` => `readonly`
	- `POST`, `PUT`, `PATCH`, `DELETE` => `readwrite`
- Endpoints that use `POST` for read behavior (such as filtered searches)
	MUST be explicitly treated as `readonly` via the route override.

### SAT RBAC v2 Tenant Scope Enforcement

For SAT RBAC v2 authorization, tenant scope enforcement SHALL be applied
after SAT capability authorization succeeds.

Tenant scope sources:

- Request tenant: header `tenantId`.
- SAT partner scope: claim `allowedResources.allowedPartners`.

Normative behavior for SAT RBAC v2 requests requiring tenant scope:

- The system SHALL read tenant from request header `tenantId`.
- The system SHALL read allowed partner scope from
	`allowedResources.allowedPartners`.
- The request SHALL be authorized only when `allowedPartners` contains
	the request `tenantId` value.
- If `tenantId` is missing, authorization SHALL be denied with
	`403 Forbidden`.
- If `allowedPartners` does not contain `tenantId`, authorization SHALL
	be denied with `403 Forbidden`.

SAT RBAC v2 tenant scope enforcement SHALL rely only on request metadata and SAT claims and SHALL NOT modify capability strings.

### Tenant Resolution By Auth Path

Tenant resolution SHALL be path-specific in this phase:

- SAT RBAC v2 path:
	- The request tenant SHALL be read from header `tenantId`.
	- Authorization SHALL enforce membership against SAT claim
		`allowedResources.allowedPartners`.
- Legacy SAT path:
	- Legacy SAT authorization semantics remain unchanged, including any validation requirements for token claims such as allowedPartners.
	- Request processing SHALL continue to support multi-tenancy.
	- In this phase, request processing SHALL resolve `tenantId`
		to the default tenant.
- Login-token/Xerxes path:
	- Login-token/Xerxes authorization semantics remain unchanged.
	- Token validation SHALL NOT enforce tenant or partner claims.
	- Request processing SHALL continue to support multi-tenancy.
	- By default, request processing SHALL resolve `tenantId` to the
		default tenant.
	- Request processing MAY resolve `tenantId` from the request header
		only when `xconfwebconfig.xconf.enable_tenant_header_for_login_token`
		is enabled.
	- If the feature flag is enabled but the header is missing or blank,
		request processing SHALL resolve `tenantId` to the default tenant.
	- The feature flag SHALL default to disabled when absent from config
		and SHALL be disabled in production deployments.

After auth-path-specific resolution, the resolved `tenantId` SHALL be
stored in request context. Downstream authorization and handlers SHALL
use the context value rather than independently resolving the tenant
from the request header.

In this phase, multi-tenant authorization guarantees apply only to
SAT RBAC v2 requests.

### Tenant Auto-Creation

The system SHALL verify that the resolved tenant exists before invoking
downstream authorization or handlers.

Tenant auto-creation SHALL follow these rules:

- SAT RBAC v2 requests MAY auto-create a missing tenant only when the
	SAT capabilities include `xconf:system:readwrite`.
- Legacy SAT requests SHALL NOT auto-create tenants.
- Login-token/Xerxes requests SHALL NOT auto-create tenants, regardless
	of the login-token tenant-header feature flag.
- The login-token tenant-header feature flag SHALL control tenant
	resolution only; it SHALL NOT grant tenant-provisioning permission.
- `testOnly` SHALL NOT grant tenant auto-creation permission. Tests that
	require tenants MAY create them directly through the database or DAO
	layer.

If tenant validation or the onboarding authorization check fails, the
system SHALL return immediately. No downstream authorization, handler
logic, or post-failure side effect SHALL execute.

For an authenticated request whose tenant cannot be used:

- The system SHALL return `403 Forbidden` when onboarding is not
	permitted or the required SAT capability is absent.
- The response body SHOULD identify whether the tenant was not found or
	whether tenant onboarding was not authorized.

### SAT RBAC v2 Deny-By-Default

If a SAT RBAC v2 request cannot be classified into `(domain, access)`
requirements, authorization SHALL be denied with `403 Forbidden`.

### Metrics Domain Constraint

The metrics domain SHALL be read-only.

- `xconf:metrics:readonly` is the only supported metrics capability.
- No readwrite functionality exists for the metrics domain (no
	`xconf:metrics:readwrite`).

### HTTP Status Semantics

The system SHALL use:

- `401 Unauthorized` only for missing or invalid authentication.
- `403 Forbidden` for authenticated-but-not-authorized requests,
	including SAT RBAC v2 classification, capability, or tenant scope
	denials.


### Fail-Fast Termination

After an authentication failure produced by this system, or
an authorization failure surfaced through this system,
request processing MUST terminate immediately.

No downstream handler logic, middleware continuation, or
post-failure side effects SHALL occur after such a failure.

This contract defines authentication-boundary authorization semantics
for routing selection, classification, and failure handling; downstream
business policy remains outside scope.


## Extension Notice
Downstream systems (including xconfas) may impose additional
authentication or authorization constraints beyond this contract.
Those constraints are explicitly outside the scope of this specification.