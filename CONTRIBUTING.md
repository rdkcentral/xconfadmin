Contributing
============
If you would like to contribute code to this project you can do so through GitHub by forking the repository
and sending a pull request.
Before RDK accepts your code into the project you must sign the RDK Contributor License Agreement (CLA).

### Branch Structure
- **`main`** - Production-ready code. All releases are tagged from this branch.
- **`develop`** - Any feature development/bugfixes must first go into this branch and once changes are validated/tested and then raise PR to main branch.
- All contributors should first raise pull requests against the develop branch after receiving the PR approvals. Once the changes are validated and tested, a cherry-pick to main branch should be performed. This ensures that production remains stable and only verified changes are submitted.


### Workflow and Branch Rules
- **Feature PRs**: Target `develop` branch
- **Bug fix PRs**: Target `develop` branch (unless critical hotfix)
- **`main`** - branch is the latest stable release.
- No direct commit to develop and main branch, instead create a feature or bugfix branch for your changes and submit a pull request (PR).