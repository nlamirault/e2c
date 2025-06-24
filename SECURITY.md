# Security Policy

## Prevention

- [Dependabot](https://docs.github.com/en/code-security/dependabot/dependabot-security-updates/about-dependabot-security-updates) helps us keep our dependencies up-to-date to patch vulnerabilities as soon as possible by creating awareness and automated PRs.
- [Whitesource Bolt for GitHub](https://www.whitesourcesoftware.com/free-developer-tools/bolt/) helps us with identifying vulnerabilities in our dependencies to raise awareness.
- [GitHub's security features](https://github.com/features/security) are constantly monitoring our repo and dependencies:
  - All pull requests (PRs) are using CodeQL to scan our source code for vulnerabilities
  - Dependabot will automatically identify vulnerabilities based on the GitHub Advisory Database and open PRs with patches
- The [Scorecard GitHub Action](https://github.com/ossf/scorecard-action) automates the process by running security checks on the GitHub repository.
