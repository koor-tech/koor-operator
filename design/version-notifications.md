---
title: version-notifications
target-version: v0.1.2
---

# Version Notifications

## Summary

### Goals

- Record current version of KSD and ceph
- Notify of the next version that is safe to upgrade to

### Non-Goals

- Upgrade the rook / ceph cluster - This will be handled by a future design document

## Proposal details

### API Changes
Add the following properties to KoorCluster:

```yaml
apiVersion: storage.koor.tech/v1alpha1
kind: KoorCluster
spec:
  ...
  upgradeOptions:
    mode: upgrade # or "disabled" or "notify"
    endpoint:  versions.koor.tech # the endpoint used to find ceph and rook versions
    schedule: 0 0 * * * # cron schedule for version check, defaults to midnight every day
status:
  ...
  currentVersions:
    ksd: v0.11.0 # current ksd version
    ceph: v17.2.5 # current ceph version
    koor-operator: v0.1.6 # current koor-operator version
  latestVersions:
    ksd: v0.11.1 # latest safe ksd version
    ceph: v17.2.6 # latest safe ceph version
    koor-operator: v0.1.7 # latest koor-operator version, will not update automatically.
```

### Controller changes
The controller will install a cronjob that queries the endpoint and updates the latest versions in the koorcluster status, then updates to the latest version.

### The endpoint
The endpoint is a server that when given the current versions returns the latest version that is safe to upgrade to. This is statically configured as JSON files in the git repository. We could add CI that checks for the latest versions from the source (Quay.io / GitHub) and creates issues for a human to test the upgrade. We could even partially automate the process in the future.

**Inputs:**
```yaml
currentVersions:
  kube: string
  koor-operator: string
  ksd: string
  ceph: string
```

**Outputs:**
```yaml
latestVersions:
  koor-operator:
    version: string
    repository: string
    chart: string
  ksd:
    version: string
    repository: string
    chart: string
  ceph:
    version: string
    image: string
    hash: string
```
### Risks and Mitigation

- Some users might not like having to connect to an external service to do upgrades
  - This could be mitigated by making our endpoint open source and allow users to install it locally.
- The endpoint needs polishing and flexibility in design
- Results may vary be kubeVersion
  - Test upgrades on multiple k8s versions

## Alternatives

- Keep updating manually
- Skip the version service and directly query quay and github
  - Unsafe because we would like to control and test version upgrades to hopefully upgrade them later.
