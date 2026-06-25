## 1. OpenSpec Artifacts
- [x] 1.1 Create proposal for SAT RBAC v2 tenant enforcement phase.
- [x] 1.2 Create design describing auth flow and tenant-scope gate ordering.
- [x] 1.3 Create spec delta for tenantId and allowedPartners enforcement.
- [x] 1.4 Define implementation and validation task plan.

## 2. Contract Update
- [x] 2.1 Update openspec/specs/auth/auth-contract.md with SAT tenant-scope enforcement requirements.
- [x] 2.2 Add tenant source definition (`tenantId` header).
- [x] 2.3 Add SAT partner scope source definition (`allowedResources.allowedPartners`).
- [x] 2.4 Add mandatory `403 Forbidden` outcomes for missing/empty/mismatch tenant scope.
- [x] 2.5 Clarify ordering: capability checks first, tenant checks second.

## 3. Implementation
- [ ] 3.1 Add SAT tenant-enforcement step in SAT RBAC v2 auth flow after capability success.
- [ ] 3.2 Read and validate request header `tenantId`.
- [ ] 3.3 Read and validate SAT claim `allowedResources.allowedPartners`.
- [ ] 3.4 Enforce membership check (`tenantId` in `allowedPartners`).
- [ ] 3.5 Return `403 Forbidden` for tenant authorization failures with fail-fast behavior.
- [ ] 3.6 Keep legacy SAT and login-token/IDP flows unchanged.

## 4. Validation
- [ ] 4.1 Add tests for tenantId missing -> `403` (SAT tenant-enforced route).
- [ ] 4.2 Add tests for allowedPartners missing -> `403`.
- [ ] 4.3 Add tests for allowedPartners empty -> `403`.
- [ ] 4.4 Add tests for tenant mismatch -> `403`.
- [ ] 4.5 Add tests for tenant match -> allow.
- [ ] 4.6 Add tests confirming capability failure still evaluated first.
- [ ] 4.7 Add tests confirming legacy SAT and login-token paths are unchanged.
