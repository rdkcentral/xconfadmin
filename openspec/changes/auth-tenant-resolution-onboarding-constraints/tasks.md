## 1. OpenSpec Artifacts
- [x] 1.1 Create proposal with implementation audit and mismatch analysis.
- [x] 1.2 Create design for tenant resolution and tenant onboarding constraints.
- [x] 1.3 Create spec delta with normative tenant and onboarding rules.
- [x] 1.4 Record explicit unknowns requiring product/security decisions.

## 2. Contract Updates (Auth Contract)
- [x] 2.1 Add auth-path-specific tenantId resolution requirements.
- [x] 2.2 Add login-token tenant-header feature flag default-off requirements.
- [x] 2.3 Add tenant auto-creation allow/deny matrix by auth type.
- [x] 2.4 Add SAT v2 capability prerequisite for tenant auto-creation.
- [x] 2.5 Add testOnly isolation requirement (test-only support not runtime-config enabled).
- [x] 2.6 Standardize status/body semantics for missing-tenant disallow paths.

## 3. Implementation Status
- [x] 3.1 Align auth middleware tenantId resolution with updated contract.
- [x] 3.2 Add login-token feature flag in xconf config block near default_tenant_id.
- [x] 3.3 Enforce SAT v2-only auto-creation with required write capability check.
- [x] 3.4 Prevent auto-creation on login-token and SAT legacy paths.
- [x] 3.5 Preserve fail-fast behavior for all denial branches.

## 4. Validation
- [x] 4.1 Add/adjust middleware unit tests for tenantId resolution matrix across auth types.
- [x] 4.2 Add/adjust tests for feature flag enabled/disabled login-token behavior.
- [x] 4.3 Add/adjust tests for SAT v2 capability-gated onboarding.
- [ ] 4.4 Add/adjust authenticated middleware tests proving no onboarding for SAT legacy.
- [ ] 4.5 Add/adjust authenticated middleware tests proving no onboarding for login-token regardless of feature flag.
- [x] 4.6 Add/adjust tests for testOnly isolation behavior.
