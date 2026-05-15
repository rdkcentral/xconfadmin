## 1. Spec And Contract Updates
- [x] 1.1 Update `openspec/specs/auth/auth-contract.md` with normative SAT RBAC v2 requirements.
- [x] 1.2 Add precedence requirements: Xerxes -> SAT v2 -> legacy SAT.
- [x] 1.3 Add SAT v2 detection requirement via `xconf:` capability prefix.
- [x] 1.4 Add route-based domain classification requirements with ordered rules.
- [x] 1.5 Add access classification requirements with `/filtered` override then HTTP method mapping.
- [x] 1.6 Add SAT v2 deny-by-default requirement for unclassifiable requests.
- [x] 1.7 Add HTTP status semantics requirement (`401` auth failure, `403` authz denial).
- [x] 1.8 Add metrics domain constraint (readonly only; no `xconf:metrics:readwrite`).

## 2. Mapping Registry And Classification
- [ ] 2.1 Implement a central route-to-domain mapping registry used by SAT v2 authorization.
- [ ] 2.2 Ensure mapping evaluation is ordered and first-match-wins.
- [ ] 2.3 Seed registry with representative patterns for `core`, `tagging`, `system`, `metrics`.
- [ ] 2.4 Add explicit readonly override patterns for POST-based read endpoints (including `/filtered`).

## 3. Authorization Flow Integration
- [ ] 3.1 Integrate precedence logic in auth middleware: Xerxes first, then SAT v2, then legacy SAT.
- [ ] 3.2 Add SAT v2 detection logic based on any capability prefixed with `xconf:`.
- [ ] 3.3 Ensure SAT v2 deny-by-default on unclassifiable `(domain, access)` requests.
- [ ] 3.4 Enforce metrics as readonly-only during SAT v2 capability checks.
- [ ] 3.5 Preserve legacy SAT behavior unchanged when SAT v2 is not detected.

## 4. HTTP Semantics And Error Handling
- [ ] 4.1 Ensure `401` is returned only for missing/invalid authentication.
- [ ] 4.2 Ensure `403` is returned for authenticated-but-not-authorized requests.
- [ ] 4.3 Ensure SAT v2 classification and capability failures consistently return `403`.
- [ ] 4.4 Verify fail-fast termination remains enforced after `401`/`403` responses.

## 5. Validation
- [ ] 5.1 Add tests for precedence behavior (Xerxes over SAT; SAT v2 over legacy SAT).
- [ ] 5.2 Add tests for SAT v2 detection by `xconf:` prefix presence.
- [ ] 5.3 Add tests for ordered route classification and first-match behavior.
- [ ] 5.4 Add tests for access classification (`/filtered` override and method-based fallback).
- [ ] 5.5 Add tests for deny-by-default on unclassified SAT v2 routes.
- [ ] 5.6 Add tests for metrics readonly-only behavior and no readwrite functionality.
- [ ] 5.7 Add tests for `401` vs `403` semantics across auth/authz scenarios.
