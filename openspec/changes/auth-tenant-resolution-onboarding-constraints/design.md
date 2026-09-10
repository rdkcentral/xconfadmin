# Authorization Tenant Resolution and Onboarding Design

## Goal

Define a deterministic auth-boundary flow for tenant resolution and tenant onboarding that:
- preserves existing SAT v2 capability and tenant-scope checks
- constrains tenant header usage by auth path
- restricts tenant auto-creation to privileged SAT v2 callers

## Current Flow (Observed)

1. Auth middleware determines auth type (SAT v2, SAT legacy, login-token).
2. Middleware resolves tenantId from header with fallback to default tenant for all auth types.
3. Middleware writes tenantId into context.
4. Middleware checks tenants table; missing tenant may be auto-created for SAT v2 or test-only mode.
5. Downstream authorization reads tenantId from context.

## Target Flow

1. Determine auth type.
2. Resolve tenantId by auth-type-specific rules.
3. Write resolved tenantId to context.
4. Check tenant existence.
5. If missing tenant, apply auth-type-specific onboarding rules.
6. Continue to downstream authorization and handlers only if checks pass.

## Tenant Resolution Algorithm

Inputs:
- authType
- request header tenantId
- configured default_tenant_id
- login-token tenant-header feature flag

Algorithm:
1. If authType is SAT v2:
- tenantId = header tenantId if present else default_tenant_id.
2. Else if authType is SAT legacy:
- tenantId = default_tenant_id.
3. Else if authType is login-token:
- If feature flag is enabled:
  - tenantId = header tenantId if present else default_tenant_id.
- Else:
  - tenantId = default_tenant_id.
4. Store tenantId in context.

## Tenant Onboarding Algorithm

Inputs:
- authType
- resolved tenantId
- tenant existence lookup
- SAT v2 capabilities
- required system write capability
- testOnly mode

Algorithm:
1. If tenant exists: continue.
2. If authType is SAT v2:
- If required system write capability absent: deny.
- Else attempt onboarding.
3. If authType is login-token: deny (no onboarding).
4. If authType is SAT legacy: deny (no onboarding).
5. testOnly support can exist in dedicated tests but not as normal runtime config behavior.

## Capability Requirement

Required capability for SAT v2 tenant auto-creation is xconf:system:readwrite.

## Config Placement Requirement

Login-token tenant-header feature flag belongs in xconf block near default_tenant_id.

Final key name is pending product decision and must be documented in auth contract and sample config.

## Failure Semantics

Must preserve fail-fast behavior:
- No downstream handler execution after auth/authz/tenant-check failure.

Final status/body for tenant-missing without onboarding permission must be explicitly standardized by contract update.

## Security Notes

- Restricting tenant header usage by auth path reduces risk of unauthorized tenant context selection.
- Preventing login-token/legacy-SAT onboarding reduces chance of accidental tenant provisioning.
- Requiring explicit SAT v2 write capability for onboarding enforces least privilege.
