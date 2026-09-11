## 1. OpenSpec Artifacts
- [x] 1.1 Create proposal for SAT tenant onboarding partner check.
- [x] 1.2 Define the onboarding flow, bypass boundaries, and lifecycle in design.
- [x] 1.3 Define normative requirements and validation matrix.

## 2. Contract Updates
- [x] 2.1 Add the allowed-partner onboarding prerequisite to the authentication contract.
- [x] 2.2 Clarify the relationship between the onboarding check and general SAT RBAC v2 tenant-scope enforcement.

## 3. Implementation
- [x] 3.1 Gate SAT RBAC v2 onboarding of a missing tenant on allowed-partner membership.
- [x] 3.2 Preserve the system write capability requirement and fail-fast behavior.

## 4. Validation
- [x] 4.1 Test listed partner permits onboarding.
- [x] 4.2 Test unlisted partner returns `403` without onboarding.
- [x] 4.3 Test missing system write capability remains denied.
- [x] 4.4 Test an existing tenant retains its existing authorization behavior.