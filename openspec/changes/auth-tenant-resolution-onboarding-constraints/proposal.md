Status: Proposed
Applied to: openspec/specs/auth/auth-contract.md

## Problem Statement

xconfadmin tenant handling and tenant auto-onboarding behavior at the authentication boundary currently allows broader tenant selection and tenant creation than intended for production safety.

The implementation recently moved tenantId resolution into auth middleware request-context setup, which changed behavior across auth types. OpenSpec currently documents SAT v2, legacy SAT, and login-token tenant behavior, but does not yet fully specify tenant auto-creation guardrails and does not reflect all current middleware behavior.

## Background and Current Behavior (Audit)

Audited implementation files:
- http/webconfig_server.go
- http/auth.go
- adminapi/auth/permission_service.go
- http/auth_validation_middleware_test.go

Audited OpenSpec files:
- openspec/specs/auth/auth-contract.md
- openspec/changes/sat-rbac-capabilities-v2/*
- openspec/changes/sat-rbac-v2-tenant-enforcement/*

Verified current implementation behavior:
- Auth type is determined as SAT v2, SAT legacy, or login token in middleware.
- tenantId is resolved once in middleware from request header tenantId with fallback to default tenant.
- Resolved tenantId is written to request context for all auth paths, not SAT v2 only.
- Tenant existence is checked in middleware before downstream handlers.
- If tenant is missing:
  - SAT v2 requests can auto-create tenant via OnboardTenantFunc.
  - Non-SAT-v2 requests can also auto-create tenant when testOnly is true.
  - Non-SAT-v2 requests return unauthorized tenant-not-found when testOnly is false.
- SAT v2 tenant auto-create is not currently gated by a system write capability check.

Verified current OpenSpec behavior:
- SAT v2 tenant scope enforcement uses tenantId and allowedResources.allowedPartners.
- Tenant resolution by auth path is documented as:
  - SAT v2 reads tenantId from header.
  - SAT legacy resolves to default tenant.
  - Login-token resolves to default tenant.
- Auto-creation policy and capability prerequisites for tenant onboarding are not explicitly specified.

## Mismatches Between Implementation and OpenSpec

1. TenantId resolution mismatch:
- OpenSpec says SAT legacy and login-token resolve tenantId to default tenant.
- Implementation currently resolves tenantId from header for all auth paths and stores it in context.

2. Missing auto-creation contract:
- OpenSpec does not currently define runtime rules for tenant auto-creation by auth path.
- Implementation currently allows SAT v2 auto-creation and testOnly-based non-SAT-v2 auto-creation.

3. Missing capability gate contract for tenant onboarding:
- Desired behavior requires write-capability-based gating for SAT v2 tenant auto-creation.
- Current implementation does not enforce this gate.

## Desired Behavior

1. TenantId resolution:
- SAT v2: read tenantId from header with fallback to default tenant when header missing.
- SAT legacy: always use default tenant; ignore tenantId header.
- Login-token: use default tenant by default; optionally allow tenantId header only when dedicated feature flag is enabled.

2. Tenant auto-creation:
- Allowed only for SAT v2 requests and only when required write capability is present.
- Not allowed for login-token requests, regardless of login-token tenant-header feature flag state.
- Not allowed for SAT legacy requests.
- Test mode support may remain limited to isolated unit/integration tests and must not be enabled by normal runtime config.

3. Tenant validation failure handling
- Requests may proceed only when:
- the resolved tenant exists, or
- tenant onboarding succeeds.
- If the tenant does not exist and the request is not permitted to onboard the tenant, request processing must terminate immediately.
- Downstream authorization checks, business logic, and handlers must not execute after tenant validation failure.

4. Authorization sequence requirement:
- Auth path and tenant resolution must be deterministic before downstream authorization checks consume tenantId from request context.

## TenantId Resolution Rules

- Rule R1 (SAT v2): resolve tenantId from header tenantId; if missing/blank, use configured default tenant.
- Rule R2 (SAT legacy): resolve tenantId to configured default tenant only.
- Rule R3 (Login-token default): resolve tenantId to configured default tenant when feature flag is disabled.
- Rule R4 (Login-token opt-in): when login-token tenant-header feature flag is enabled, resolve tenantId from header tenantId if present, else fallback to default tenant.
- Rule R5: resolved tenantId must be written to request context and used consistently by downstream authorization and handlers.

## Tenant Auto-Creation Rules

- Rule C1: tenant existence check occurs after tenantId resolution and before downstream handler execution.
- Rule C2: auto-create missing tenant only for SAT v2 requests.
- Rule C3: SAT v2 auto-create requires required system write capability.
- Rule C4: login-token requests must never auto-create tenant.
- Rule C5: SAT legacy requests must never auto-create tenant.
- Rule C6: test-only onboarding support is allowed only for isolated test execution mode and is out of normal runtime configuration scope.

## Authorization and Capability Requirements

- SAT v2 tenant auto-creation must require capability xconf:system:readwrite.

## Feature Flag Behavior (Login-Token Tenant Header Support)

- Feature flag default is disabled when absent from config.
- When disabled, login-token requests always use default tenant.
- When enabled, login-token requests may use header tenantId for lower-environment UI testing only.
- Production deployments must keep this flag disabled.
- This flag only affects tenantId resolution for login-token requests; it does not grant tenant auto-creation.

## Backward Compatibility Notes

- SAT v2 authorization and tenant-scope semantics remain unchanged except for added tenant auto-creation capability guard.
- SAT legacy behavior becomes strictly default-tenant-only as currently documented.
- Login-token default behavior remains default-tenant-only when flag is unset/false.
- Existing lower-environment UI workflows that rely on login-token tenant header require explicit opt-in via feature flag.

## Security Considerations

- Restricting header-based tenant selection by auth path reduces header spoofing impact on legacy SAT and login-token flows.
- Disallowing auto-creation for login-token and legacy SAT prevents unintended tenant provisioning by less constrained auth paths.
- Requiring explicit SAT v2 system write capability for auto-creation limits tenant provisioning to explicitly privileged SAT callers.
- Default-off feature flag preserves safe production posture.

## Test Plan

Required tests:
- SAT v2 tenantId resolution:
  - Header present => context tenantId equals header.
  - Header missing => context tenantId equals default tenant.
- SAT legacy tenantId resolution:
  - Header present => context tenantId equals default tenant.
- Login-token tenantId resolution:
  - Flag disabled/absent + header present => context tenantId equals default tenant.
  - Flag enabled + header present => context tenantId equals header.
  - Flag enabled + header missing => context tenantId equals default tenant.
- Tenant auto-create gating:
  - SAT v2 + missing tenant + required capability => onboarding attempted.
  - SAT v2 + missing tenant + missing required capability => onboarding not attempted and request denied per policy.
  - Login-token + missing tenant (flag enabled or disabled) => onboarding not attempted.
  - SAT legacy + missing tenant => onboarding not attempted.
- Test-only behavior:
  - testOnly pathways remain isolated to test configuration and are not activated by runtime config.

## Explicit Open Questions (No Assumptions)

1. Feature flag config key:
- Desired location is in xconf section near default_tenant_id.
- Key name is not finalized by this proposal. Candidate examples:
  - xconfwebconfig.xconf.enable_ui_tenant_header
  - xconfwebconfig.xconf.login_token_allow_tenant_header
- Decision needed: final key name for spec and implementation.

2. Failure status for disallowed auto-create paths:
- Current implementation has mixed responses for tenant-not-found paths.
- Decision needed: exact status/body contract for tenant missing when auto-create is not permitted.
