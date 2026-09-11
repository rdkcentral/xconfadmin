# SAT Tenant Onboarding Partner Check

## Problem Statement

SAT RBAC v2 requests with `xconf:system:readwrite` may currently onboard a
resolved tenant that is absent from the tenants table. Tenant onboarding is a
provisioning side effect and must be constrained to the partners authorized by
the SAT token's `allowedResources.allowedPartners` claim.

## Desired Behavior

- Before onboarding a missing tenant for a SAT RBAC v2 request, verify the
  resolved `tenantId` is in `allowedResources.allowedPartners`.
- Keep `xconf:system:readwrite` as a separate required prerequisite.
- Deny the request with `403 Forbidden` and do not invoke onboarding when the
  partner check fails.
- Do not alter the established authorization behavior for an existing tenant.

## Scope

This change applies only to SAT RBAC v2 auto-creation of a tenant that does not
already exist. It does not change legacy SAT or login-token onboarding policy,
SAT capability requirements, or general SAT RBAC v2 tenant-scope enforcement.

## Security Considerations

Tenant onboarding is a provisioning side effect and is allowed only when the
resolved tenant is included in the SAT token's `allowedResources.allowedPartners` 
claim and the caller has `xconf:system:readwrite`. If the partner check fails, 
the request returns `403 Forbidden` and tenant onboarding is not invoked. 
Existing-tenant authorization behavior is unchanged.