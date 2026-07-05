## 1. Spec And Contract Updates
- [x] 1.1 Update `openspec/specs/auth/auth-contract.md` with normative SAT RBAC v2 requirements.
- [x] 1.2 Add routing-based selection requirements: Authorization header -> SAT path; else Xerxes path.
- [x] 1.3 Add SAT v2 detection requirement via `xconf:` capability prefix.
- [x] 1.4 Add route-based domain classification requirements with ordered rules.
- [x] 1.5 Add access classification requirements with `/filtered` override then HTTP method mapping.
- [x] 1.6 Add SAT v2 deny-by-default requirement for unclassifiable requests.
- [x] 1.7 Add HTTP status semantics requirement (`401` auth failure, `403` authz denial).
- [x] 1.8 Add metrics domain constraint (readonly only; no `xconf:metrics:readwrite`).

## 2. Mapping Registry And Classification
- [x] 2.1 Implement a central route-to-domain mapping registry used by SAT v2 authorization.
- [x] 2.2 Ensure mapping evaluation is ordered and first-match-wins.
- [x] 2.3 Seed registry with representative patterns for `core`, `tagging`, `system`, `metrics`.
- [x] 2.4 Verify POST-based read endpoints (e.g. `/filtered`) use read authorization paths.
  - Audited all POST `/filtered` handlers.
  - All matched handlers use `auth.CanRead`.
  - No POST `/filtered` handlers use `auth.CanWrite`.
  - No POST `/filtered` handlers are missing auth checks.
  - No separate readonly override registry needed at this time because access semantics are already determined by handler auth method.


## 3. Authorization Flow Integration
- [x] 3.1 Integrate routing-based selection in auth middleware: Authorization header selects SAT path; otherwise Xerxes path.
- [x] 3.2 Add SAT v2 detection logic based on any capability prefixed with `xconf:`.
- [x] 3.3 Ensure SAT v2 deny-by-default on unclassifiable `(domain, access)` requests.
- [x] 3.4 Enforce metrics as readonly-only during SAT v2 capability checks.
- [x] 3.5 Preserve legacy SAT behavior unchanged when SAT v2 is not detected.

## 4. HTTP Semantics And Error Handling
- [x] 4.1 Ensure `401` is returned only for missing/invalid authentication.
- [x] 4.2 Ensure `403` is returned for authenticated-but-not-authorized requests.
- [x] 4.3 Ensure SAT v2 classification and capability failures consistently return `403`.
- [x] 4.4 Verify fail-fast termination remains enforced after `401`/`403` responses.

## 5. Validation
- [ ] 5.1 Add tests for routing behavior (Authorization header selects SAT path; SAT v2 vs legacy SAT selection on SAT path; Xerxes path only when Authorization is absent).
  - Deferred: requires SAT/login-token middleware test harness with local JWT/key validation or test seams.
- [x] 5.2 Add tests for SAT v2 detection by `xconf:` prefix presence.
- [x] 5.3 Add tests for ordered route classification and first-match behavior.
- [x] 5.4 Verify POST /filtered endpoints use read authorization (`CanRead`) rather than write authorization.
- [x] 5.5 Add tests for deny-by-default on unclassified SAT v2 routes.
- [x] 5.6 Add tests for metrics readonly-only behavior and no readwrite functionality.
- [ ] 5.7 Add tests for `401` vs `403` semantics across auth/authz scenarios.
  - 403 authorization semantics covered by permission_service tests.
  - 401 middleware-level coverage deferred with 5.1 pending middleware test harness.

