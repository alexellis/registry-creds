## registry-creds operator

This operator can be used to propagate a single ImagePullSecret to all namespaces within your cluster, so that images can be pulled with authentication.

Pulling images with authentication is required in two scenarios:
* To extend the Docker Hub anonymous pull limits to a practical number
* To access private registries or repos

The normal process is as follows:

* Create a secret
* Edit your service account, and add the name of the secret to `imagePullSecrets`

## Installation

Only development instructions are available at this time

Get the pre-reqs: kubectl (`arkade get kubectl`) and [kustomize](https://github.com/kubernetes-sigs/kustomize/releases/tag/kustomize%2Fv3.5.4)

Install with:

```bash
git clone https://github.com/alexellis/registry-creds
cd registry-creds
make install
make run
```

## Usage

To use this operator create a `ClusterPullSecret` CustomResource and apply it to your cluster.

```yaml
apiVersion: ops.alexellis.io/v1
kind: ClusterPullSecret
metadata:
  name: dockerhub
spec:
  secret:
    data:
      .dockerconfigjson: base64-encodedtextgoeshere
    type: kubernetes.io/dockerconfigjson
```

You can obtain the text for the `.dockerconfigjson` field by running:

```bash
export USERNAME=username
export PW=mypassword
export EMAIL=me@example.com

kubectl create secret docker-registry temp-pull-secret \
  --docker-username=$USERNAME \
  --docker-password=$PW \
  --docker-email=$EMAIL \
  --dryrun -o yaml
```

If you're not using the Docker Hub, then add `--docker-password`

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
