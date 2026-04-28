Status: Accepted
Applied to: openspec/specs/auth/auth-contract.md

## Why

Authentication and authorization failures currently rely on handler authors to stop execution after an authentication or authorization failure is produced, which is easy to miss and can allow unintended logic to continue. We need an explicit contract that fail-fast termination is mandatory whenever an auth/authz failure is produced.

## What Changes

- Clarify the auth contract to require that request processing terminates immediately after an authentication or authorization failure is produced.
- Define this behavior as a normative guarantee so downstream handlers and middleware can rely on no post-failure side effects.
- Align xconfadmin auth behavior documentation with this fail-fast requirement so future implementations and refactors preserve the same safety property.

## Non‑Goals

- This change does not alter authentication mechanisms, token types,
  permission models, or authorization semantics.
- This change does not introduce new authentication or authorization
  capabilities.
- This change does not modify request/response formats or API contracts.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `auth`: Add explicit fail-fast requirements that handler execution MUST NOT continue after authentication or authorization failures are identified.

## Impact

- Affected specs: `openspec/specs/auth/auth-contract.md`.
- Affected implementation areas (future apply phase): auth middleware and handlers that write 401/403/related auth error responses.
- API behavior impact: no new endpoints; clarifies control-flow guarantees for existing auth/authz failure paths.
- System impact: improves safety by preventing accidental post-failure processing and side effects.