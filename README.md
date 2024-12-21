## registry-creds operator

[![Sponsor this](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&link=https://github.com/sponsors/alexellis)](https://github.com/sponsors/alexellis) [![build](https://github.com/alexellis/registry-creds/actions/workflows/build.yaml/badge.svg)](https://github.com/alexellis/registry-creds/actions/workflows/build.yaml) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


This operator can be used to propagate a single ImagePullSecret to all namespaces within your cluster, so that images are pulled using authentication.

See also: [ROADMAP.md](/ROADMAP.md)

### Use-case 1: Propagate a private registry secret to all namespaces

The second use-case for this operator is to take an authentication token which is required to pull images from a private registry, and to make sure it's available and configured for each and every namespace.

For example, if you were running a multi-tenant service, where each tenant has their own namespaces, and every image is sourced from a common private registry. You could use this operator to propagate the pull secret for each namespace.

### Use-case 2: Docker Hub Rate Limits

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

## Do you use `registry-creds`?

`k3sup` was created by [Alex Ellis](https://github.com/users/alexellis/sponsorship) - the founder of [OpenFaaS &reg;](https://www.openfaas.com/) & [inlets](https://inlets.dev/). 

<a href="https://github.com/sponsors/alexellis/">
<img alt="Sponsor this project" src="https://github.com/alexellis/alexellis/blob/master/sponsor-today.png" width="90%">
</a>

Want to see continued development? [Sponsor alexellis on GitHub](https://github.com/users/alexellis/sponsorship)

## License

MIT
