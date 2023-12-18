## PR Title
- Always begin the title with the ticket number
- Title should succinctly describe the PR purpose
- Remove this section when you creating a PR

## Summary
- Two bullet summary of the PR.
- First bullet should be a link to the ticket
- First word of second bullet should be "BUG" or "FEATURE"
- e.g. BUG: divide by zero in webpa req percentage calculation
- e.g. 2 FEATURE: New metrics for wifiblaster latencies

## Details
 - Include more details, 5-6 lines at least unless the PR is truly trivial.
 - If it is a new feature, this section should be filled out.
 - A bulleted list is preferred

## Checklist
- PR Reviewers should make sure that this section is filled up. If a bullet does not have a checkmark, or NO, ask the dev to fill it up.
- PR Reviewers to ensure that code coverage is not significantly different before and after the PR. (Guideline - 1% less coverage may be allowed)
- [] Code coverage from unit tests before this PR
- [] Code coverage from unit tests after this PR
- [] Does this change the db schema? If yes, flag Venkata
- [] Is an Ansible PR needed? If yes, provide the link
- [] Is a MOPS entry needed? If yes, please provided the link to MOPS
- [] Are any metrics needed?
