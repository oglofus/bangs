# Accessibility

Oglofus Bangs aims to be usable by people regardless of disability, assistive technology, device, or environment. Accessibility is an ongoing part of project quality and contributor review.

## Goals

- Follow WCAG 2.2 AA guidance where it applies to the web experience.
- Keep the redirect service and its documentation understandable without relying on color, position, motion, or visual-only instructions.
- Make contribution and issue-reporting workflows usable with keyboard navigation and assistive technology.
- Treat accessibility reports as valuable user expertise and prioritize barriers that prevent core tasks.

## Supported scope

The project includes a Cloudflare Worker and a small web entry point. Accessibility expectations apply to the web entry point, documentation, issue forms, pull request templates, and contributor tooling. The redirect response itself is intentionally minimal and may send users to third-party search destinations whose accessibility is outside this project's control.

## Known limitations

Third-party destinations and search result pages are not controlled by Oglofus Bangs. If a destination creates a barrier, report the issue to that service as well as to this project when the redirect or project-controlled interface contributes to the problem.

## Reporting an accessibility barrier

Use the [accessibility issue form](https://github.com/oglofus/bangs/issues/new?template=accessibility.yml). Include the expected and actual behavior, reproduction steps, your operating system, browser, assistive technology and versions when relevant, severity, and any workaround. Do not include personal or sensitive information.

If the barrier is a security vulnerability, follow [SECURITY.md](SECURITY.md) instead.

## Contributor checks

For user-facing changes, check keyboard-only operation, visible focus, labels and error messages, screen-reader behavior, zoom/reflow, contrast, and reduced motion where applicable. Record what you checked in the pull request.
