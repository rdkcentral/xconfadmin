## 1. OpenSpec Artifacts
- [x] 1.1 Create proposal with implementation audit and mismatch analysis.
- [x] 1.2 Create design for tenant resolution and tenant onboarding constraints.
- [x] 1.3 Create spec delta with normative tenant and onboarding rules.
- [x] 1.4 Record explicit unknowns requiring product/security decisions.

## 2. Contract Updates (Auth Contract)
- [ ] 2.1 Add auth-path-specific tenantId resolution requirements.
- [ ] 2.2 Add login-token tenant-header feature flag default-off requirements.
- [ ] 2.3 Add tenant auto-creation allow/deny matrix by auth type.
- [ ] 2.4 Add SAT v2 capability prerequisite for tenant auto-creation.
- [ ] 2.5 Add testOnly isolation requirement (test-only support not runtime-config enabled).
- [ ] 2.6 Standardize status/body semantics for missing-tenant disallow paths.

## 3. Implementation Follow-Up (Out of This Proposal)
- [ ] 3.1 Align auth middleware tenantId resolution with updated contract.
- [ ] 3.2 Add login-token feature flag in xconf config block near default_tenant_id.
- [ ] 3.3 Enforce SAT v2-only auto-creation with required write capability check.
- [ ] 3.4 Prevent auto-creation on login-token and SAT legacy paths.
- [ ] 3.5 Preserve fail-fast behavior for all denial branches.

## 4. Validation
- [ ] 4.1 Add/adjust middleware unit tests for tenantId resolution matrix across auth types.
- [ ] 4.2 Add/adjust tests for feature flag enabled/disabled login-token behavior.
- [ ] 4.3 Add/adjust tests for SAT v2 capability-gated onboarding.
- [ ] 4.4 Add/adjust tests proving no onboarding for SAT legacy.
- [ ] 4.5 Add/adjust tests proving no onboarding for login-token regardless of feature flag.
- [ ] 4.6 Add/adjust tests for testOnly isolation behavior.
