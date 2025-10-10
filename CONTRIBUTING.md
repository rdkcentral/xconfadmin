Contributing to xconfadmin
==========================

If you would like to contribute code to this project you can do so through GitHub by forking the repository and sending a pull request.

**Before RDK accepts your code into the project you must sign the RDK Contributor License Agreement (CLA).**

## Branch Management

### Branch Structure
- **`main`** - Production-ready code. All releases are tagged from this branch.
- **`develop`** - Integration branch for new features. All feature development should be based on this branch.


### Workflow and Branch Rules
- **Feature PRs**: Target `develop` branch
- **Bug fix PRs**: Target `develop` branch (unless critical hotfix)
- No direct commit to develop and main branch,create a feature or bugfix branch for your changes and submit a pull request (PR) to merge into the develop or main branches as appropriate."

## Pull Request Guidelines

### PR Title Format
- Always begin the title with the ticket number
- Title should succinctly describe the PR purpose
- Remove this section when you creating a PR

### PR Description Template
```
## Summary
• [Link to ticket](ticket-url)
• BUG/FEATURE: Brief description of the change

## Details
• Provide 5-6 lines of details unless the PR is trivial
• Use bulleted lists for better readability
• If it's a new feature, this section must be filled out
• Include technical approach and implementation notes
• Mention any breaking changes or migration steps

## Checklist
PR Reviewers should ensure this section is completed:
- [ ] is Unit tests included
- [ ] Code coverage from unit tests before this PR: _%
- [ ] Code coverage from unit tests after this PR: _%
- [ ] Does this change the db schema? If yes, flag for review
```

### Requirements Before Creating PR
- Ensure your branch is up-to-date with target branch
- Run all tests locally: `make test`
- Write/update tests for your changes
- Code coverage should not decrease significantly

## Code Standards
- Follow standard Go conventions and use `gofmt`
- Add comments for exported functions and complex logic
- Handle errors appropriately
- Use meaningful variable and function names

## Review Process
- **For `develop` branch**: At least 1 approval from maintainer
- **For `main` branch**: At least 2 approvals from maintainers
- All CI checks must pass including CLA verification
- Code coverage must be maintained
