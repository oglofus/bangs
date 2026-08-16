# Contributing to Oglofus Bangs

Thank you for helping improve this project. Contributions should be focused, documented, testable, and accessible to the people who read and use them.

## Before you start

1. Search open and recently closed issues for an existing request.
2. Open a detailed GitHub issue before making a change. The issue is the source of truth for scope, acceptance criteria, and tests.
3. For security vulnerabilities, follow [the security policy](SECURITY.md) instead of opening a public issue.
4. For accessibility barriers, use the [accessibility issue form](https://github.com/oglofus/bangs/issues/new?template=accessibility.yml).

Please follow the [Code of Conduct](CODE_OF_CONDUCT.md) when participating.

## Local setup

You need Git, Go 1.25 or later, Node.js 18 or later, pnpm, and the GitHub CLI for the complete validation workflow.

Clone the repository and run commands from its root:

```bash
git clone https://github.com/oglofus/bangs.git
cd bangs
pnpm install
```

## Make a change

Create a GitHub-linked branch from the issue when you have repository access:

```bash
gh issue develop ISSUE_NUMBER \
  --name issue-ISSUE_NUMBER-short-slug \
  --base main \
  --checkout
```

Keep changes focused and do not include unrelated local or generated files.

### Updating bang data

When `bangs.json` changes, regenerate the binary data and worker bundle:

```bash
go run ./preprocessor/main.go
pnpm run build
```

## Validation

Run the checks relevant to your change and record the exact commands and results in the issue and pull request:

```bash
go test ./...
pnpm run build
git diff --check
```

For documentation and issue-form changes, also check heading order, descriptive link text, valid YAML, keyboard-readable checklists, and whether every referenced file exists.

## Accessibility expectations

Accessibility is part of quality and the definition of done. Read [ACCESSIBILITY.md](ACCESSIBILITY.md) and the [Accessibility Best Practices for Your Project](https://opensource.guide/accessibility-best-practices-for-your-project/) guide.

For documentation and templates:

- Use logical heading levels, plain language, descriptive links, and real lists.
- Do not communicate meaning only through color, position, or styling.
- Provide text alternatives for meaningful images and diagrams.
- Keep instructions understandable without relying on visual layout.

For user-facing changes, check keyboard operation, visible focus, labels and error messages, color-independent meaning, reduced motion, screen-reader behavior, zoom/reflow, and contrast where applicable.

## Pull requests

Open the pull request against `main` after the issue's implementation and tests are complete. The pull request should:

- Explain what changed and why.
- Include `Closes #ISSUE_NUMBER` or an appropriate GitHub closing reference.
- List runnable test commands and their results.
- Include numbered manual-test instructions for anything that could not be run.
- Identify documentation, security, and accessibility considerations.
- Keep the review focused and respond constructively to feedback.

Do not merge your own pull request unless repository policy and the required review/checks allow it.

## License

By contributing, you agree that your contributions are provided under the repository's [Boost Software License 1.0](LICENSE).
