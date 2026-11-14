# Authentication Contract (xconfadmin)

## Purpose
This document defines the authentication behavior provided by
xconfadmin as a standalone library.

## Scope
This specification describes:
- credential validation
- identity resolution
- authentication success and failure outcomes
- routing-based authorization selection across SAT and IDP
- fail-fast termination after authentication or authorization failure

This specification does not describe:
- business-specific policy enforcement
- application type or partner enforcement policy
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

### Authorization Routing Selection

Authorization selection SHALL be deterministic and route credentials by
credential type, not by evaluating IDP and SAT in precedence order.

Normative behavior:

- If the `Authorization` header is present, the request SHALL be treated
	as SAT-authenticated and SAT processing SHALL be selected.
- Else, if an IDP login token is present (header `token` or cookie `token`),
	the system SHALL authorize using IDP permissions.
- Else, the system SHALL return `401 Unauthorized`.

When both SAT and IDP credentials are present, `Authorization`-header
routing MUST win; SAT selection SHALL be used and IDP SHALL NOT be
evaluated for that request.

### HTTP Status Semantics

The system SHALL use:

- `401 Unauthorized` only for missing or invalid authentication.
- `403 Forbidden` for authenticated-but-not-authorized requests.


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