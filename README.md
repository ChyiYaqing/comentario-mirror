# Comentario

**[Comentario](https://comentario.app)** is a fork of [Commento](https://gitlab.com/commento/commento) by Adhityaa Chandrasekar, an open source web comment server that has been discontinued: see a list of major differences below.

[Homepage](https://comentario.app) * [Demo](https://demo.comentario.app) * [Author's blog](https://yktoo.com)

**Comentario** is a platform that you can embed in your website to allow your readers to add comments. It's lightweight and fast.

Comentario supports Markdown syntax, import from Disqus, comment voting, automated spam detection, moderation tools, sticky comments, thread locking, OAuth login, single sign-on, and email notifications.

## FAQ

### How is this different from Disqus, Facebook Comments, and the rest?

Most other products in this space do not respect your privacy; showing ads is their primary business model and that nearly always comes at the users' cost. Comentario has no ads; you're the customer, not the product.

Comentario is also orders of magnitude lighter than alternatives.

### Why should I care about my readers' privacy?

For starters, your readers value their privacy. Not caring about them is disrespectful, and you will end up alienating your audience; they won't come back. Disqus still isn't GDPR-compliant (according to their <a href="https://help.disqus.com/terms-and-policies/privacy-faq" title="At the time of writing (28 December 2018)" rel="nofollow">privacy policy</a>). Disqus adds megabytes to your page size; what happens when a random third-party script that is injected into your website turns malicious?

### How does Comentario differ from its predecessor Commento?

There are quite a few major points (and counting):

* Comentario is running the latest and greatest software versions: Go 1.20, Postgres 15.x (older version supported down to 9+), ES6 and so on.
* The "embeddable" part (`comentario.js`) is a complete rewrite:
    * Code is modernised and reimplemented using Typescript.
    * Layouts are optimised for all screen sizes (300 pixels up).
    * Login, Signup, and Markdown Help are made popup dialogs (we're using [Popper](https://popper.js.org/) for correct positioning).
    * Login, Signup, and Comment Editor are using HTML `form` element and proper `autocomplete` attribute values, which makes them compatible with password managers.
    * Improvements for WCAG (accessibility), including keyboard navigation.
    * Subtle animations are added.
    * Keyboard-enabled dialogs (<kbd>Escape</kbd> cancels, <kbd>Enter</kbd> (<kbd>Ctrl</kbd><kbd>Enter</kbd> in a multiline field) submits the dialog).
    * Tons of other issues and inconsistencies have been fixed.
* Dropped support for local service installation. Instead, we're recommending deploying Comentario in the cloud, ideally into a Kubernetes cluster — there's a Helm chart for that.

## Installation

### Deploying into a Kubernetes cluster

#### Prerequisites

1. Comentario is installed using the [Helm package manager](https://helm.sh/) 3.x.
2. We're using [certmanager](https://cert-manager.io/) for dealing with SSL certificates in the cluster: requesting and renewing.
3. Once you have `certmanager` up and running, create a new `ClusterIssuer` for Let's Encrypt. Or, even better, two issuers: `letsencrypt-staging` for experimenting with your installation (so that you don't hit Let's Encrypt usage limits) and `letsencrypt-prod` for production usage.
4. The Helm chart does not include PostgreSQL: it has to be installed separately (have a look at [.gitlab-ci.yml](.gitlab-ci.yml) for inspiration).

#### Deployment

1. Create a new namespace (in these examples I'll refer to it as `$NAMESPACE`): `kubectl create namespace $NAMESPACE`
2. Edit the values in `k8s/comentario-secrets.yaml` as required. Don't forget to base64-encode the values as the last step.
3. Create the secret: `kubectl create -f k8s/comentario-secrets.yaml -n $NAMESPACE`
4. Install Comentario using Helm (adjust the values as you see fit):
```bash
helm upgrade --install \
    --namespace $NAMESPACE \                            # The same namespace value as above
    --set "clusterIssuer=letsencrypt-staging" \         # Replace with letsencrypt-prod when you're ready for production
    --set "image.repository=registry.gitlab.com/comentario/comentario" \
    --set "image.tag=<VERSION>" \                       # Use the desired Comentario version here
    --set "comentario.secretName=comentario-secrets" \  # This is the name of the secret from k8s/comentario-secrets.yaml
    --set "comentario.smtpHost=mail.example.com" \      # Name of the SMTP host you're using for emails
    --set "comentario.smtpFromAddress=x@example.com" \  # Email to set in the Reply field
    --set "ingress.host=comment.example.com" \          # Domain where your Comentario instance should be reachable on 
    my-comentario \                                     # Name of your instance (and Helm release)
    helm/comentario
```

### Backing up the database

```bash
kubectl exec -t -n $NAMESPACE \
    $(kubectl get -n $NAMESPACE pods -l app.kubernetes.io/instance=comentario-postgres -o name) \
    -- pg_dump -U postgres -d comentario > /path/to/comentario.sql
```

### Restoring the database from backup

We cannot send it via the pipe directly (dunno why), so we copy it over first and clean up afterwards.

```bash
PG_POD=$(kubectl get -n $NAMESPACE pods -l app.kubernetes.io/instance=comentario-postgres -o 'jsonpath={.items..metadata.name}')
kubectl cp -n $NAMESPACE /path/to/comentario.sql $PG_POD:/tmp/c.sql
kubectl exec -t -n $NAMESPACE $PG_POD -- psql -U postgres -d comentario -f /tmp/c.sql
kubectl exec -t -n $NAMESPACE $PG_POD -- rm /tmp/c.sql
```
