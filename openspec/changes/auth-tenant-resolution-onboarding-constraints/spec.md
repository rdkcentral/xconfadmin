# Authorization and Tenant Onboarding Constraints Specification

## Spec Delta Summary

This proposal updates openspec/specs/auth/auth-contract.md with additional normative requirements for:
- tenantId resolution by auth path
- tenant auto-creation constraints
- login-token tenant-header feature flag behavior
- SAT v2 capability prerequisite for tenant auto-creation

## Normative Behavior

### TenantId Resolution By Auth Path

1. SAT v2 requests:
- The system SHALL resolve tenantId from request header tenantId.
- If header tenantId is missing or blank, the system SHALL resolve tenantId to configured default tenant.

2. SAT legacy requests:
- The system SHALL resolve tenantId to configured default tenant.
- The system SHALL NOT use request header tenantId for SAT legacy requests.

3. Login-token requests:
- By default, the system SHALL resolve tenantId to configured default tenant.
- The system MAY use request header tenantId only when dedicated login-token tenant-header feature flag is enabled.
- If feature flag is enabled and header tenantId is missing or blank, the system SHALL resolve tenantId to configured default tenant.

4. Context propagation:
- After resolution, the system SHALL store resolved tenantId in request context for downstream authorization and handlers.

### Tenant Auto-Creation

1. General rule:
- The system SHALL verify whether resolved tenant exists before downstream handler logic proceeds.

2. Allowed path:
- The system SHALL allow tenant auto-creation only for SAT v2 requests.

3. Capability prerequisite:
- For SAT v2 tenant auto-creation, caller SHALL possess capability xconf:system:readwrite.

4. Failure Handling
- If the resolved tenant does not exist, the system SHALL attempt onboarding only when all onboarding prerequisites are satisfied.
- If onboarding is not permitted or onboarding fails, request processing SHALL terminate immediately.
- No downstream authorization checks, business logic, or request handlers SHALL execute after tenant validation failure.
- The request SHALL return an authorization error indicating that the tenant is unavailable and the caller is not permitted to onboard it.

5. Disallowed paths:
- The system SHALL NOT auto-create tenants for login-token requests.
- The system SHALL NOT auto-create tenants for SAT legacy requests.

6. Test-mode isolation:
- Test-only onboarding support MAY exist for isolated unit/integration tests.
- Test-only onboarding support SHALL NOT be enabled via normal runtime configuration.

### Feature Flag Requirements

1. Login-token tenant-header feature flag:
- The feature flag SHALL default to disabled when absent from configuration.
- The feature flag SHALL only control login-token tenantId resolution from header.
- The feature flag SHALL NOT permit tenant auto-creation for login-token requests.

2. Production posture:
- Production deployments SHALL keep this feature flag disabled.

## Compatibility

- SAT v2 routing, detection, and tenant-scope authorization remain intact.
- SAT legacy default-tenant behavior is preserved and tightened against tenant header override.
- Login-token default-tenant behavior is preserved unless feature flag is explicitly enabled.
- Existing clients that do not send tenantId are unaffected due to default-tenant fallback.

## Security Considerations

- Auth-path-specific tenant resolution reduces cross-tenant header injection risk.
- Auto-create restrictions prevent unintended tenant provisioning through login-token or legacy SAT paths.
- Capability-gated SAT v2 onboarding enforces least privilege for tenant provisioning actions.

## Validation Matrix

| Auth Type | Feature Flag | Header tenantId | Missing Tenant | Capability Present | Expected Tenant Resolution | Auto-Create |
|---|---|---|---|---|---|---|
| SAT v2 | N/A | Yes | Yes | Yes | Header tenant | Allowed |
| SAT v2 | N/A | Yes | Yes | No | Header tenant | Denied |
| SAT v2 | N/A | No | Yes | Yes | Default tenant | Allowed |
| SAT legacy | N/A | Yes | Yes | Any | Default tenant | Denied |
| Login-token | Disabled | Yes | Yes | Any | Default tenant | Denied |
| Login-token | Enabled | Yes | Yes | Any | Header tenant | Denied |
| Login-token | Enabled | No | Yes | Any | Default tenant | Denied |

## Explicit Unknowns Requiring Decision

1. Final feature flag key name in xconf config block.
2. Exact failure status/body when tenant is missing and auto-create is not allowed.
