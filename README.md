[Homepage](https://comentario.app) · [Demo](https://demo.comentario.app) · [Documentation](https://docs.comentario.app/) · [Author's blog](https://yktoo.com/)

# Comentario

**[Comentario](https://comentario.app)** is a platform that you can embed in your website to allow your readers to add comments. It's lightweight and fast.

Comentario supports Markdown syntax, import from Disqus, comment voting, automated spam detection, moderation tools, sticky comments, thread locking, OAuth login, single sign-on, and email notifications.

**Comentario** is a fork of [Commento](https://gitlab.com/commento/commento) by Adhityaa Chandrasekar, an open source web comment server that has been discontinued: see a list of major differences below.

## FAQ

### How is this different from Disqus, Facebook Comments, and the rest?

Most other products in this space do not respect your privacy; showing ads is their primary business model and that nearly always comes at the users' cost. Comentario has no ads; you're the customer, not the product.

Comentario is also orders of magnitude lighter than alternatives.

### Why should I care about my readers' privacy?

For starters, your readers value their privacy. Not caring about them is disrespectful, and you will end up alienating your audience; they won't come back. Disqus adds megabytes to your page size; what happens when a random third-party script that is injected into your website turns malicious?

### How does Comentario differ from its predecessor Commento?

There are quite a few major points (and counting):

* Comentario is running the latest and greatest software versions: Go 1.20, Postgres 15.x (older version supported down to 9.6), ES6 and so on.
* The "embeddable" part (`comentario.js`) is a complete rewrite:
    * Code is modernised and reimplemented using Typescript.
    * Layouts are optimised for all screen sizes (300 pixels up).
    * Login, Signup, and Markdown Help are made popup dialogs (we're using [Popper](https://popper.js.org/) for correct positioning).
    * Login, Signup, and Comment Editor are using HTML `form` element and proper `autocomplete` attribute values, which makes them compatible with password managers.
    * Improvements for WCAG (accessibility), including keyboard navigation.
    * Subtle animations are added.
    * Keyboard-enabled dialogs (<kbd>Escape</kbd> cancels, <kbd>Enter</kbd> (<kbd>Ctrl</kbd><kbd>Enter</kbd> in a multiline field) submits the dialog).
    * Tons of other issues and inconsistencies have been fixed.
* Dropped support for local service installation. Instead, we're recommending deploying Comentario in the cloud, ideally into a Kubernetes cluster — there's a [Helm chart](https://docs.comentario.app/en/getting-started/installation/helm-chart/) for that.
* The Comentario server ("backend") is using automated code generation from an Open API spec, with lots of extra checks and validations.
* Resolved all issues with OAuth identity providers (Google, GitHub, GitLab, Twitter), including user avatars.
* Every change is automatically end-to-end-tested using Cypress to prevent regressions.

## Getting started

Please refer to [Comentario documentation](https://docs.comentario.app/en/getting-started/) to learn how to install and configure it.
