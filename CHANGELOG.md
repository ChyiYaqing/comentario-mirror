# Comentario changelog

## v2.2.3

This release brings no extra functionality to Comentario, but rather concentrates on the automated build pipeline, stability, and [documentation](https://docs.comentario.app/).

We're now using Cypress for end-to-end (e2e) tests (the proper tests will follow).

## v2.2.2

**Changes:**

* Helm chart: add `comentario.indexHtmlConfigMapName` config value (073c0b8)
* Serve favicon at root (a56ea0f)
* Tie style to Comentario colours (e1b21f4)
* Fix: Vue error in dashboard (ac4993f)

## v2.2.1

**Changes:**

* Allow serving `index.html` at root when present (20bb3db)
* Fix: comment voting turned score into `NaN` for zero-score comment (bca19a3)
* Allow moderator edit others' comments (resolves #2) (84c5ec1)
* Allow interrupting connection process with `SIGINT` (0a0e83e, 40c13b8)

## v2.2.0

* This release features a major backend overhaul: Comentario is now using server code generated with [go-swagger](https://goswagger.io/) based on the [Swagger/OpenAPI spec](swagger/swagger.yml).
* All available federated authentication options are fully functional again: GitHub, GitLab, Google, Twitter, and SSO.
* This is the last Comentario version that's fully compatible (meaning, backward- and forward-compatible) with Commento database v1.8.0.
* It's also *almost* compatible with Commento API, with the exception that it consumes `application/json` instead of `application/x-www-form-urlencoded`.

**Changes:**

* Twitter OAuth re-added (9446502, ab1f244)
* Fix: avatar handling and resizing for all identity providers (59c8643)
* Fix: federated auth completion (proper HTML gets served) (a0c4626)
* OAuth flows refactored (2533eda, af56d81, dc2c9c6)
* Gzip producer for downloads (4c8df85)
* Comentario Helm chart and image updates (802dddb, 9d0a645, 4f06183, 968059c, a89a99a)
* Backend refactoring: OpenAPI code generator used (26e099c, 1b0ab10, 27b9e6f, b127050, f82c1be, 1ae87f4, e57dc4c, fe2306d, 8139ae4, e8ebe29, c84828a, dd03b35, 6c99df9, 90c095c, b3ac79c)

## v2.1.0

**Changes:**

* Bump ci-tools v2, Go 1.20, Postgres 15-alpine (cf574c1)
* Restyle error box (f7b2b6b)
* Hide all controls when page load has failed (f7b2b6b)
* Add Helm chart (508a72f, 0a029ab, 2ea9354, 4696d6e, c464c8f, 89232e3, 8a8b29d, 4e17bb2, 945d8e8, c529653, 57b2b8e)
* Rebranding Commento → Comentario (f143215, 8803b26, 5e7d5ea)
* Highlight and scroll to added comment (161222b)
* Move card options to the bottom (4655d3f)
* Validate and submit forms using Ctrl+Enter (a30c430)
* Close dialogs with Esc (82e4163)
* Visual input validation (9271bf6)
* Popup confirmation dialog on comment delete (2a539ea)
* Ditch Makefiles and prod/devel targets (d255a86)
* Blur/animate backdrop (82e4163)
* Add Popper, redesign dialogs & make them responsive (b81d555, 4260dcd)
* DB connect: use a progressive delay and up to 10 attempts (29c0df8)
* Add `nofollow noopener noreferrer` to profile links (c398f5a)
* Move version to console message appearing upon init (6f050af)
* Fix: anonymous checkbox (00939d0)
* Fix: footer overlapping with following content (2918264)
* Fix: Comentario load when session token invalid (e64fa8a)
* Refactor the frontend into components and DSL pattern (5de1790, 3e2fc44, ca9643f, dea5fd9, 4fd1d02, 64b1903, 6776ed1, 7d71261, 33e0d4b, 23808de, 8ce6def)
* docs: reflow the license text (8f7916b)

## v2.0.1

This is the very first public release of Comentario, a successor of (seemingly discontinued) [Commento](https://gitlab.com/commento/commento) (resolves commento/commento#414).

**Changes:**

* Add this changelog (resolves commento/commento#344)
* Modernise all code and its dependencies. Migrate to Go 1.19, Node 18 (62d0ff0, 6818638, c6db746, e9beec9; resolves commento/commento#407, commento/commento#331, resolves commento/commento#421)
* Drop support for non-ES6 browsers (Chrome 50-, Firefox 53-, Edge 14-, Safari 9-, Opera 37-, IE 11-) (62d0ff0)
* Resolve potential resource leak in api/version.go (62d0ff0)
* Place login/signup fields on a form and add `autocomplete` attribute. Submit the login or the signup with Enter. This must enable proper support for password managers, it also eliminates a browser warning about password field not contained by a form (f477a71, 0923f96; resolves commento/commento#138)
* Fix doubling comment on login via OAuth2 (c181c2e; resolves commento/commento#342) and locally (582455c)
* Force nofollow and target="_blank" on external links (d90b8bd; resolves commento/commento#341)
* Remove Twitter OAuth 1 as obsolete and dysfunctional (e9beec9)
* Migrate commento.js to TypeScript + Webpack (a22ed44, ca4ee7b, ef37fd4, dafb8ac, f575dc0, e349806)
* Backend: handle errors properly (4d92d4f)
* Backend: filter out deleted comments (1672508)
* Reimplement build pipeline for `dev` or tags (f654924, e3e55a6, 02a9beb, 6aa9f58, 9a65b3d, f7f6628)
* Other, internal changes.
