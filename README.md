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

This is a very early, working prototype, do not use it in production. When you move to production, you can use something like Flux, Argo or Terraform (see appendix) for managing secrets across namespaces.

The primary purpose of this tool is to ease the every-day lives of developers and new-comers to Kubernetes.

You can create a test cluster very quickly with something like [KinD](https://kind.sigs.k8s.io/docs/user/quick-start/) to try it out.

Backlog (done):
- [x] Create secrets in each namespace at start-up
- [x] Use a "seed" secret via an object reference
- [x] Watch new namespaces and create new secrets
- [x] Update the ImagePullSecret list for the default ServiceAccount in each namespace
- [x] Add an exclude annotation for certain namespaces `alexellis.io/registry-creds.ignore`
- [x] Add Docker image for `x86_64`
- [x] Test and update kustomize
- [x] Add multi-arch Docker image for `x86_64` and arm

Todo:
- [ ] Use `apierrors.IsNotFound(err)` everywhere instead of assuming an error means not found
- [ ] Support alterations/updates to the primary `ClusterPullSecret`
- [ ] Add helm chart
- [ ] Add one-liner with an arkade app - `arkade install registry-creds --username $USERNAME --password $PASSWORD`

## Installation

![Diagram](./diagram.jpg)

* The operator requires CRUD permission on secrets within all namespaces, and uses a ClusterRole.
* A custom resource is also installed called `ClusterPullSecret`, scoped to the global cluster level.
* You'll create a "seed secret", which the ClusterPullSecret will point to
* The operator will clone your seed secret to each namespace, and append it to the default ServiceAccount's imagePullSecrets list

### Deploy to a cluster

Apply the YAML for the manifest.

```bash
kubectl apply -f https://raw.githubusercontent.com/alexellis/registry-creds/master/mainfest.yaml
```

### Or make locally

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

### Create a seed secret and `ClusterPullSecret`

To use this operator create a `ClusterPullSecret` CustomResource and apply it to your cluster.

Create a "seed" secret so that it can be referenced by the ClusterPullSecret. You can customise the name, and namespace as per your own preference.

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

### Rotate your seed secret and `ClusterPullSecret`

If you want to update your `ClusterPullSecret`, then update your main "seed" secret, delete the `ClusterPullSecret` entry, and create it again. The owner references will garbage collect the older secrets and re-create them again.

### Exclude a namespace from being updated

Disable:

```bash
kubectl create ns alex
kubectl annotate ns alex alexellis.io/registry-creds.ignore=1
```

Enable:

```bash
kubectl annotate ns alex alexellis.io/registry-creds.ignore=0 --overwrite
```

## Testing it out

Do you want to see it all in action, but don't have time to waste? You're in luck, [OpenFaaS](https://www.openfaas.com/) provides a very easy to use workflow for creating a quick Docker image that servers HTTP traffic, and that can be deployed to Kubernetes.

The easiest way to test the controller is with a private image on the Docker Hub, but you can also use the [arkade project](https://get-arkade.dev) to install a self-hosted registry, with authentication enabled and TLS.

```bash
arkade install ingress-nginx
arkade install cert-manager

arkade install docker-registry
arkade install docker-registry-ingress \
 --email me@example.com \
 --domain reg.example.com
```

Run the above on a computer with a public IP, or use the [inlets-operator](https://github.com/inlets/inlets-operator) to expose your local registry on the Internet, and to get a TLS certificate for it.

Then go ahead and deploy something like OpenFaaS, create a function, and push it to your registry.

```bash
# Get the OpenFaaS CLI, and put it in `PATH`
arkade get faas-cli

# Install openfaas, and follow the login instructions
arkade install openfaas

# Create a new function and push it to your registry
faas-cli new --lang go --prefix reg.example.com/functions awesome-api
faas-cli up -f awesome-api.yml
```

You'll see `reg.example.com/functions/awesome-api:latest` being built, pushed and deployed to your cluster.

Check the event-stream to see the image being pulled and started:

```bash
kubectl get event -n openfaas-fn -w

LAST SEEN   TYPE     REASON              OBJECT                      MESSAGE
8s          Normal   Scheduled           pod/api-577d87c687-knd54    Successfully assigned openfaas-fn/api-577d87c687-knd54 to kind-control-plane
7s          Normal   Pulling             pod/api-577d87c687-knd54    Pulling image "alexellis2/http-api"
8s          Normal   SuccessfulCreate    replicaset/api-577d87c687   Created pod: api-577d87c687-knd54
8s          Normal   ScalingReplicaSet   deployment/api              Scaled up replica set api-577d87c687 to 1
0s          Normal   Pulled              pod/api-577d87c687-knd54    Successfully pulled image "alexellis2/http-api"
0s          Normal   Created             pod/api-577d87c687-knd54    Created container api
0s          Normal   Started             pod/api-577d87c687-knd54    Started container api
```

## Appendix

Terraform:

* [Snippet to imperatively apply a secret throughout your cluster](https://gist.github.com/phumberdroz/81885c01c2207d578c17635afce1b033) by Pierre Humberdroz

Tools to manage Kubernetes configuration, declaratively:

* [ArgoCD](https://argoproj.github.io/argo-cd/) - install via `arkade install argocd`
* [Flux](https://fluxcd.io)
