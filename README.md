# Comentario

**[Comentario](https://comentario.app)** is a fork of [Commento](https://gitlab.com/comentario/comentario) by Adhityaa Chandrasekar, an open source web comment server that has been discontinued.

* [Homepage](https://comentario.app)

Comentario is a platform that you can embed in your website to allow your readers to add comments. It's reasonably fast lightweight. Supports markdown, import from Disqus, voting, automated spam detection, moderation tools, sticky comments, thread locking, OAuth login, single sign-on, and email notifications.

## FAQ

### How is this different from Disqus, Facebook Comments, and the rest?

Most other products in this space do not respect your privacy; showing ads is their primary business model and that nearly always comes at the users' cost. Comentario has no ads; you're the customer, not the product.

Comentario is also orders of magnitude lighter than alternatives.

### Why should I care about my readers' privacy?

For starters, your readers value their privacy. Not caring about them is disrespectful, and you will end up alienating your audience; they won't come back. Disqus still isn't GDPR-compliant (according to their <a href="https://help.disqus.com/terms-and-policies/privacy-faq" title="At the time of writing (28 December 2018)" rel="nofollow">privacy policy</a>). Disqus adds megabytes to your page size; what happens when a random third-party script that is injected into your website turns malicious?

## Installation

### Deploying in Kubernetes

1. Create a new namespace: `kubectl create namespace ys-comentario`
2. Edit the values in `k8s/comentario-secrets.template.yaml` and save it as `k8s/comentario-secrets.yaml`
3. Create the secret: `kubectl create -f k8s/comentario-secrets.yaml`

### Backing up the database

```bash
kubectl exec -t -n ys-comentario \
    $(kubectl get -n ys-comentario pods -l app.kubernetes.io/instance=comentario-postgres -o name) \
    -- pg_dump -U postgres -d comentario > /path/to/comentario.sql
```

### Restoring the database from backup

We cannot send it via the pipe directly (dunno why), so we copy it over first and clean up afterwards.

```bash
PG_POD=$(kubectl get -n ys-comentario pods -l app.kubernetes.io/instance=comentario-postgres -o 'jsonpath={.items..metadata.name}')
kubectl cp -n ys-comentario /path/to/comentario.sql $PG_POD:/tmp/c.sql
kubectl exec -t -n ys-comentario $PG_POD -- psql -U postgres -d comentario -f /tmp/c.sql
kubectl exec -t -n ys-comentario $PG_POD -- rm /tmp/c.sql
```
