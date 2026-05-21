#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="${ROOT_DIR}/bin/testlogs"
mkdir -p "${LOG_DIR}"

TS="$(date +%Y%m%d_%H%M%S)"
JSON_LOG="${LOG_DIR}/go-test-${TS}.jsonl"
SUMMARY_LOG="${LOG_DIR}/go-test-${TS}.summary.txt"

echo "Running tests with JSON logging..."
echo "Log file: ${JSON_LOG}"
echo "Summary file: ${SUMMARY_LOG}"

ulimit -n 10000

set +e
go test ./... -json -p 1 -parallel 1 -cover -count=1 -timeout=45m | tee "${JSON_LOG}"
TEST_EXIT=${PIPESTATUS[0]}
set -e

python3 - "${JSON_LOG}" <<'PY' | tee "${SUMMARY_LOG}"
import json
import sys
from collections import defaultdict

path = sys.argv[1]

stats = defaultdict(lambda: {
    "run": 0,
    "pass": 0,
    "fail": 0,
    "skip": 0,
    "elapsed": 0.0,
    "pkg_status": "unknown",
    "failing_tests": []
})

with open(path, "r", encoding="utf-8", errors="replace") as f:
    for line in f:
        line = line.strip()
        if not line:
            continue
        try:
            evt = json.loads(line)
        except Exception:
            continue

        pkg = evt.get("Package")
        if not pkg:
            continue

        action = evt.get("Action")
        test = evt.get("Test")

        if test and action == "run":
            stats[pkg]["run"] += 1
        elif test and action == "pass":
            stats[pkg]["pass"] += 1
        elif test and action == "fail":
            stats[pkg]["fail"] += 1
            stats[pkg]["failing_tests"].append(test)
        elif test and action == "skip":
            stats[pkg]["skip"] += 1

        if not test and action in ("pass", "fail"):
            stats[pkg]["pkg_status"] = action
            if "Elapsed" in evt:
                stats[pkg]["elapsed"] = float(evt["Elapsed"])

if not stats:
    print("\nNo package stats found in JSON log.")
    sys.exit(0)

print("\n=== Test Metrics By Package ===")
header = f"{'Package':70} {'Run':>6} {'Pass':>6} {'Fail':>6} {'Skip':>6} {'Elapsed(s)':>11} {'Status':>8}"
print(header)
print("-" * len(header))

total_run = total_pass = total_fail = total_skip = 0
for pkg in sorted(stats.keys()):
    s = stats[pkg]
    total_run += s["run"]
    total_pass += s["pass"]
    total_fail += s["fail"]
    total_skip += s["skip"]
    print(f"{pkg:70} {s['run']:6d} {s['pass']:6d} {s['fail']:6d} {s['skip']:6d} {s['elapsed']:11.3f} {s['pkg_status']:>8}")

print("-" * len(header))
print(f"{'TOTAL':70} {total_run:6d} {total_pass:6d} {total_fail:6d} {total_skip:6d}")

failing_pkgs = [p for p, s in stats.items() if s["pkg_status"] == "fail" or s["fail"] > 0]
if failing_pkgs:
    print("\n=== Failing Tests ===")
    for pkg in sorted(failing_pkgs):
        tests = stats[pkg]["failing_tests"]
        uniq = []
        seen = set()
        for t in tests:
            if t not in seen:
                uniq.append(t)
                seen.add(t)
        print(f"{pkg}")
        if uniq:
            for t in uniq:
                print(f"  - {t}")
        else:
            print("  - package failed before running explicit tests")

print(f"\nFull JSON log: {path}")
PY

echo "Summary log: ${SUMMARY_LOG}"

exit ${TEST_EXIT}
