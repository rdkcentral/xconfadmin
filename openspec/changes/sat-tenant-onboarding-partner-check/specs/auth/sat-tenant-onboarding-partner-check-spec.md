## ADDED Requirements

### Requirement: SAT RBAC v2 allowed-partner tenant onboarding authorization
The system SHALL verify that `allowedResources.allowedPartners` contains the
resolved `tenantId` before invoking tenant onboarding for a SAT RBAC v2 request
whose resolved `tenantId` is absent from the tenants table. This is an
additional prerequisite to the existing `xconf:system:readwrite` capability
requirement.

#### Scenario: Listed partner may be onboarded
- **WHEN** a SAT RBAC v2 request resolves to a missing tenant listed in `allowedResources.allowedPartners`
- **AND** the token has `xconf:system:readwrite`
- **THEN** the system attempts tenant onboarding

#### Scenario: Unlisted partner is denied
- **WHEN** a SAT RBAC v2 request resolves to a missing tenant not listed in `allowedResources.allowedPartners`
- **AND** the token has `xconf:system:readwrite`
- **THEN** the system returns `403 Forbidden`
- **AND** the system does not invoke tenant onboarding
- **AND** request processing terminates before downstream authorization, handler logic, or post-failure side effects execute

#### Scenario: Missing system write authorization remains denied
- **WHEN** a SAT RBAC v2 request resolves to a missing tenant
- **AND** the token lacks `xconf:system:readwrite`
- **THEN** the system returns `403 Forbidden`
- **AND** the system does not invoke tenant onboarding
