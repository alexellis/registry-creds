## Roadmap & status

The primary purpose of this tool is to ease the every-day lives of developers and new-comers to Kubernetes. When you move to production, you can use something like Flux, Argo or Terraform (see appendix) for managing secrets across namespaces.

Disclaimer: see the [license of this project](/LICENSE) before deploying or using it.

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
- [x] Add one-liner with an arkade app - `arkade install registry-creds --username $DOCKER_USERNAME --password $PASSWORD`
- [x] ~~Add helm chart~~ - static manifest available instead
- [x] Use `apierrors.IsNotFound(err)` everywhere instead of assuming an error means not found
- [x] Support additional ServiceAccounts beyond the `default` account in each namespace

Todo:
- [ ] Remove pull secret reference from ServiceAccounts upon ClusterPullSecret deletion
- [ ] Propagate alterations/updates to the primary `ClusterPullSecret` in each namespace when the secret value changes (the work-around is to delete and re-create the ClusterPullSecret)

