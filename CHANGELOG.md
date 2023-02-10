# Comentario changelog

## v2.1.0

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
