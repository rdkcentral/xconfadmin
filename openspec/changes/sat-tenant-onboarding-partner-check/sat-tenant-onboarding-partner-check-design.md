# SAT Tenant Onboarding Partner Check Design

## Goal

Prevent SAT RBAC v2 callers from provisioning a missing tenant outside the
partner scope granted by their SAT token.

## Onboarding Flow

For a request whose resolved tenant does not exist:

1. Deny requests that are not SAT RBAC v2.
2. Deny SAT RBAC v2 requests without `xconf:system:readwrite`.
3. Deny if `allowedResources.allowedPartners` does not contain the resolved
   `tenantId`.
4. When all applicable prerequisites pass, invoke tenant onboarding.
5. Continue to downstream authorization and handlers only after onboarding
   succeeds.

For a tenant that already exists, this feature does not add or remove any
authorization behavior. Existing SAT RBAC v2 tenant-scope enforcement remains
independent from this onboarding gate.

## Failure Semantics

An allowed-partner failure returns `403 Forbidden`, does not invoke the
onboarding function, and terminates request processing. The response should
describe the authorization failure without exposing unnecessary tenant state.