## registry-creds operator

[![CI status](https://github.com/alexellis/registry-creds/actions/workflows/ci-only.yaml/badge.svg)](https://github.com/alexellis/registry-creds/actions/workflows/ci-only.yaml)

This operator can be used to propagate a single ImagePullSecret to all namespaces within your cluster, so that images can be pulled with authentication.

### Why is this operator required?

The primary reason for creating this operator, is to make it easier for users of Kubernetes to consume images from the Docker Hub after [recent pricing and rate-limiting changes](https://www.docker.com/pricing) were brought in, an authenticated account is now required to pull images.

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

See also: [ROADMAP.md](/ROADMAP.md)

## Getting started

### Set up your sponsorship

This tool requires time and effort to maintain, so if you use it at work, you should become a sponsor on GitHub:

* [Set-up a sponsorship now](https://github.com/sponsors/alexellis)

It's up to you to choose a tier or a custom amount based upon the value and time saving of the tool, plus its maintainance.

Alternative approaches are available if you don't agree to these terms.

### Install & configure the tool

* [Install the tool](GUIDE.md)

### Support

[Sponsors](https://github.com/sponsors/alexellis) may raise a GitHub issue to propose changes, new features and to request help with usage.
