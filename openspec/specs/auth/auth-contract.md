# Authentication Contract (xconfadmin)

## Purpose
This document defines the authentication behavior provided by
xconfadmin as a standalone library.

## Scope
This specification describes:
- credential validation
- identity resolution
- authentication success and failure outcomes
- authorization precedence across Xerxes, SAT v2, and legacy SAT
- SAT v2 request classification requirements
- fail-fast termination after authentication or authorization failure

This specification does not describe:
- business-specific policy enforcement
- tenant or partner enforcement policy
- downstream extensions or constraints

## Guarantees

### Credential Validation
The system SHALL validate supplied credentials and determine
their validity deterministically.

### Authentication Result
On successful authentication, the system SHALL return an
identity representation suitable for downstream use.

### Failure Modes
Authentication failures SHALL result in defined error categories.
Failure handling is subject to the Fail-Fast Termination guarantee
defined below.

### Authorization Precedence

Authorization MUST be evaluated in this order:

1. Xerxes permissions
2. SAT RBAC v2
3. legacy SAT

Normative behavior:

- If a Xerxes token is present and valid, the system SHALL authorize
	using Xerxes permissions and SHALL NOT evaluate SAT v2 or legacy SAT.
- If no valid Xerxes authorization path is available and SAT contains at
	least one capability with prefix `xconf:`, the system SHALL authorize
	using SAT RBAC v2 semantics.
- If SAT does not contain any capability with prefix `xconf:`, the system
	SHALL authorize using legacy SAT behavior unchanged.

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
	including SAT RBAC v2 classification or capability denials.


### Fail-Fast Termination

After an authentication failure produced by this system, or
an authorization failure surfaced through this system,
request processing MUST terminate immediately.

No downstream handler logic, middleware continuation, or
post-failure side effects SHALL occur after such a failure.

This contract defines authentication-boundary authorization semantics
for precedence, classification, and failure handling; downstream
business policy remains outside scope.


## Extension Notice
Downstream systems (including xconfas) may impose additional
authentication or authorization constraints beyond this contract.
Those constraints are explicitly outside the scope of this specification.