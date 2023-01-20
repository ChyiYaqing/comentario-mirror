# Comentario changelog

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
