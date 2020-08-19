## registry-creds operator

This operator can be used to propagate a single ImagePullSecret to all namespaces within your cluster, so that images can be pulled with authentication.

### Why is this operator required?

The primary reason for creating this operator, is to make it easier for users of Kubernetes to consume images from the Docker Hub after [recent pricing and rate-limiting changes](https://www.docker.com/pricing) were brought in, an authenticated account is now required to pull images.

* Unauthenticated users: 100 layers / 6 hours
* Authenticated users: 200 layers / 6 hours
* Paying, authenticated users: unlimited downloads

See also: [Docker Hub rate limits & pricing](https://www.docker.com/pricing)

Pulling images with authentication is required in two scenarios:
* To extend the Docker Hub anonymous pull limits to a practical number
* To access private registries or repos on the Docker Hub

The normal process is as follows, which becomes tedious and repetitive when you have more than one namespace in a cluster.

* Create a secret
* Edit your service account, and add the name of the secret to `imagePullSecrets`

## Status

This is a very early, working prototype. Do not use it in production, you can create a test cluster very quickly with something like [KinD](https://kind.sigs.k8s.io/docs/user/quick-start/).

Backlog (done):
- [x] Create secrets in each namespace at start-up
- [x] Use a "seed" secret via an object reference
- [x] Watch new namespaces and create new secrets
- [x] Update the ImagePullSecret list for the default ServiceAccount in each namespace
- [x] Add an exclude annotation for certain namespaces `alexellis.io/registry-creds.ignore`
- [x] Add Docker image for `x86_64`
- [x] Test and update kustomize

Todo:
- [ ] Use `apierrors.IsNotFound(err)` everywhere instead of assuming an error means not found
- [ ] Support alterations/updates to the primary `ClusterPullSecret`
- [ ] Add multi-arch Docker image for `x86_64` and arm
- [ ] Add helm chart
- [ ] Add one-liner with an arkade app - `arkade install registry-creds --username $USERNAME --password $PASSWORD`

## Installation

Only development instructions are available at this time

You can use the [arkade project](https://get-arkade.dev) to get CLIs the easy way, or find your way to the releases page of each application required.

If you don't have a local Kubernetes cluster, you can create one with k3d, or KinD

```bash
arkade get kind
kind create cluster
```

Get the pre-reqs: kubectl and kustomize

```bash
arkade get kubectl
arkade get kustomize
```

Install with:

```bash
git clone https://github.com/alexellis/registry-creds
cd registry-creds
make install
make run
```

> Note, you can also run `make install deploy` to try running in-cluster.

## Usage

To use this operator create a `ClusterPullSecret` CustomResource and apply it to your cluster.

Create a secret so that it can be referenced by the ClusterPullSecret. You can customise the name, and namespace as per your own preference.

```bash
export USERNAME=username
export PW=mypassword
export EMAIL=me@example.com

kubectl create secret docker-registry registry-creds-secret \
  --namespace kube-system
  --docker-username=$USERNAME \
  --docker-password=$PW \
  --docker-email=$EMAIL
```

If you're not using the Docker Hub, then add `--docker-password`

Now create a `ClusterPullSecret` YAML file, and populate the `secretRef` with the secret name and namespace from above.

```yaml
apiVersion: ops.alexellis.io/v1
kind: ClusterPullSecret
metadata:
  name: dockerhub
spec:
  secretRef:
    name: registry-creds-secret
    namespace: kube-system
```

## Testing it out

Do you want to see it all in action, but don't have time to waste? You're in luck.

You can use the [arkade project](https://get-arkade.dev) to install a self-hosted registry, with authentication enabled and TLS.

```bash
arkade install docker-registry
arkade install docker-registry-ingress \
 --email me@example.com \
 --domain reg.example.com
```

Then go ahead and deploy something like OpenFaaS, create a function, and push it to your registry:

```bash
arkade install openfaas
arkade get faas-cli

# Follow the login instructions

faas-cli new --lang go --prefix reg.example.com/functions awesome-api

faas-cli up -f awesome-api.yml
```

You'll see `reg.example.com/functions/awesome-api:latest` being built, pushed and deployed to your cluster.

Check the event-stream to see the image being pulled and started:

```
kubectl get event -n openfaas-fn -w
```
