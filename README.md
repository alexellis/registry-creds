## registry-creds operator

[![CI status](https://github.com/alexellis/registry-creds/actions/workflows/ci-only.yaml/badge.svg)](https://github.com/alexellis/registry-creds/actions/workflows/ci-only.yaml)

This operator can be used to propagate a single ImagePullSecret to all namespaces within your cluster, so that images are pulled using authentication.

See also: [ROADMAP.md](/ROADMAP.md)

### Use-case: Propagate a private registry secret to all namespaces

The second use-case for this operator is to take an authentication token which is required to pull images from a private registry, and to make sure it's available and configured for each and every namespace.

For example, if you were running a multi-tenant service, where customers had their own namespaces, and every Pod was pulled from a common private registry. You could use this operator to automate what would otherwise be a manual and error-prone process.

### Use-case: Docker Hub Rate Limits

The original need for this operator, was to make it easier for users of Kubernetes to consume images from the Docker Hub after [recent pricing and rate-limiting changes](https://www.docker.com/pricing) were brought in, an authenticated account is now required to pull images.

These are the limits as understood at time of writing:

* Unauthenticated users: 100 pulls / 6 hours
* Authenticated users: 200 pulls / 6 hours
* Paying, authenticated users: unlimited downloads

Read also: [Docker Hub rate limits & pricing](https://www.docker.com/pricing)

Pulling images with authentication is required in two scenarios:
* To extend the Docker Hub anonymous pull limits to a practical number
* To access private registries or repos on the Docker Hub

The normal process is as follows, which becomes tedious and repetitive when you have more than one namespace in a cluster.

* Create a secret
* Edit your service account, and add the name of the secret to `imagePullSecrets`

## Getting Started

* [Install the tool](GUIDE.md)

## License

MIT
