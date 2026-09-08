# SAT Tenant Onboarding Partner Check Specification

## Spec Delta Summary

This change adds an allowed-partner prerequisite to SAT RBAC v2 tenant
auto-creation.

## Normative Behavior

### SAT RBAC v2 Tenant Auto-Creation Partner Check

For a SAT RBAC v2 request whose resolved `tenantId` is absent from the tenants
table, the system SHALL verify that `allowedResources.allowedPartners` contains
the resolved `tenantId` before invoking tenant onboarding.

When the allowed-partner claim does not contain the resolved `tenantId`, the
system SHALL:

- return `403 Forbidden`;
- not invoke tenant onboarding; and
- terminate before downstream authorization, handler logic, or post-failure
  side effects execute.

This feature SHALL apply only to attempted onboarding of a missing tenant. It
SHALL NOT modify authorization semantics for an existing tenant.

## Validation Matrix

| Tenant exists | SAT v2 system write | Tenant in allowedPartners | Expected result |
|---|---|---|---|
| No | Yes | Yes | Onboarding is attempted |
| No | Yes | No | `403`; onboarding is not invoked |
| No | No | Yes | `403`; onboarding is not invoked |
| Yes | Any | No | Existing tenant authorization behavior is unchanged |