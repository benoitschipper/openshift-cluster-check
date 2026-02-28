# OpenShift Cluster Health Checker

A lightweight, in-cluster Go binary that periodically polls OpenShift/Kubernetes APIs and exposes five binary Prometheus gauges for platform component health.

## Purpose

SREs and operators need a simple, reliable way to detect degraded control-plane state without deploying heavy monitoring stacks. This health-checker exposes binary signals (`0` = healthy, `1` = unhealthy) that can be scraped by any Prometheus-compatible monitoring stack and used to trigger alerts.

---

## Metrics Reference

| Metric | Source | Unhealthy Condition |
|---|---|---|
| `openshift_cluster_operators_degraded` | `ClusterOperator` (all except `etcd`) | Any operator has `Degraded=True` or `Available=False` |
| `openshift_nodes_not_ready` | `Node` | Any node has condition `Ready != True` |
| `openshift_system_pods_failing` | `Pod` (system namespaces only) | Any pod has `phase=Failed` or a container in `CrashLoopBackOff`, `OOMKilled`, or `Error` |
| `openshift_clusterversion_degraded` | `ClusterVersion` named `version` | `Degraded=True` or `Available=False` |
| `openshift_etcd_degraded` | `ClusterOperator` named `etcd` | `Degraded=True` or `Available=False` |

All metrics are Prometheus `Gauge` type with values `0` (healthy) or `1` (unhealthy).

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `CHECK_INTERVAL` | `30` | How often to run health checks, in seconds. Must be a positive integer. |
| `METRICS_PORT` | `8080` | HTTP port for the `/metrics` endpoint. Must be 1–65535. |
| `SYSTEM_NAMESPACE_PREFIXES` | `openshift-,kube-` | Comma-separated list of namespace prefixes considered system namespaces for pod checks. |
| `SYSTEM_NAMESPACES` | _(empty)_ | Comma-separated list of exact namespace names considered system namespaces for pod checks. Empty by default — the `kube-` prefix covers all `kube-*` namespaces. |

### Extending the Namespace Filter

The pod health check only inspects pods in system/platform namespaces. By default, these are:
- All namespaces starting with `openshift-`
- All namespaces starting with `kube-` (covers `kube-system`, `kube-public`, `kube-node-lease`, and any future `kube-*` namespaces added by Kubernetes)

The `kube-` prefix is used instead of an exact-match list to ensure forward-compatibility: any new `kube-*` namespace introduced by future Kubernetes versions is automatically included without configuration changes.

To add additional namespaces, set the environment variables in `deploy/deployment.yaml`:

**Add a namespace prefix** (e.g., to include `my-infra-logging`, `my-infra-metrics`):
```yaml
- name: SYSTEM_NAMESPACE_PREFIXES
  value: "openshift-,kube-,my-infra-"
```

**Add an exact namespace name** (e.g., to include `monitoring`):
```yaml
- name: SYSTEM_NAMESPACES
  value: "monitoring"
```

Both variables can be combined. A namespace is included if it matches **any** configured prefix **or** any configured exact name.

---

## RBAC Requirements

The health-checker requires a `ClusterRole` with the following read-only permissions:

| Resource | API Group | Verbs |
|---|---|---|
| `nodes` | `""` (core) | `get`, `list` |
| `pods` | `""` (core) | `get`, `list` |
| `namespaces` | `""` (core) | `get`, `list` |
| `clusteroperators` | `config.openshift.io` | `get`, `list` |
| `clusterversions` | `config.openshift.io` | `get`, `list` |

No `watch`, write, patch, update, delete, or mutate permissions are granted. `watch` is intentionally omitted because the health-checker uses a polling model (ticker-based), not an informer/watch-stream model. Granting `watch` would open a persistent streaming connection that is never used.

---

## SCC Requirements

The health-checker is compatible with the OpenShift **`restricted`** SCC (or **`restricted-v2`** on OpenShift 4.11+). **No privileged or custom SCC is required.**

**Why no custom SCC is needed:**
- `runAsNonRoot: true` is enforced at both pod and container level
- No `runAsUser` is set — OpenShift's `restricted` SCC assigns an arbitrary UID from the namespace's allocated UID range at pod start. The container image is built to be UID-agnostic (binary owned by group 0 with group-executable permissions), so it runs correctly under any assigned UID.
- Read-only root filesystem (`readOnlyRootFilesystem: true`)
- No privilege escalation (`allowPrivilegeEscalation: false`)
- All Linux capabilities dropped (`capabilities.drop: [ALL]`)
- No `hostPath` mounts, no host networking, no host PID, no host IPC

**Arbitrary UID behavior:** OpenShift injects a random non-root UID (e.g., `1000650000`) from the namespace's UID range. This is intentional and required for `restricted` SCC compatibility. The health-checker binary and container image are designed to work correctly under any such UID.

---

## Manifest Apply Order

Apply the manifests in this order to avoid dependency errors:

```bash
# 1. Create the target namespace first
kubectl create namespace openshift-health-checker

# 2. Apply manifests in order
kubectl apply -f deploy/serviceaccount.yaml
kubectl apply -f deploy/clusterrole.yaml
kubectl apply -f deploy/clusterrolebinding.yaml
kubectl apply -f deploy/deployment.yaml
kubectl apply -f deploy/service.yaml
```

Or apply all at once (Kubernetes handles ordering for non-dependent resources):
```bash
kubectl apply -f deploy/
```

---

## Building the Container Image

The container image is built to support **arbitrary UIDs** as required by OpenShift's `restricted` SCC. The binary is owned by group 0 (root group) with group-executable permissions, so it runs correctly under any non-root UID assigned by OpenShift at pod start.

> **Important:** Images built before this change may have a fixed `USER` instruction or incorrect file ownership. You must rebuild the image after this change to ensure arbitrary UID support.

```bash
# Build the image
docker build -t quay.io/your-org/health-checker:latest .

# Verify arbitrary UID support (optional local test)
docker run --rm --user 123456:0 quay.io/your-org/health-checker:latest --help 2>&1 || true

# Push to your registry
docker push quay.io/your-org/health-checker:latest
```

Update the `image:` field in `deploy/deployment.yaml` with your actual image reference before applying.

---

## Example Prometheus Alert Rules

```yaml
groups:
  - name: openshift-cluster-health
    rules:
      - alert: ClusterOperatorDegraded
        expr: openshift_cluster_operators_degraded == 1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "One or more OpenShift cluster operators are degraded"
          description: "At least one ClusterOperator (excluding etcd) is Degraded=True or Available=False for more than 5 minutes."

      - alert: EtcdDegraded
        expr: openshift_etcd_degraded == 1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "etcd ClusterOperator is degraded"
          description: "The etcd ClusterOperator is Degraded=True or Available=False."

      - alert: NodeNotReady
        expr: openshift_nodes_not_ready == 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "One or more cluster nodes are not ready"
          description: "At least one Node has condition Ready != True for more than 5 minutes."

      - alert: SystemPodFailing
        expr: openshift_system_pods_failing == 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "One or more system pods are failing"
          description: "At least one pod in a system namespace is in Failed phase or has a container in CrashLoopBackOff/OOMKilled/Error state."

      - alert: ClusterVersionDegraded
        expr: openshift_clusterversion_degraded == 1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "ClusterVersion is degraded"
          description: "The ClusterVersion 'version' is Degraded=True or Available=False."

      - alert: HealthCheckerMissing
        expr: absent(openshift_cluster_operators_degraded)
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Health checker metrics are missing"
          description: "The openshift-cluster-health-checker is not reporting metrics. The pod may be down."
```

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  OpenShift Cluster                  │
│                                                     │
│  ┌──────────────────────────────────────────────┐   │
│  │         health-checker Pod                   │   │
│  │                                              │   │
│  │  ┌────────────┐    ┌──────────────────────┐  │   │
│  │  │  Checker   │    │   Prometheus HTTP    │  │   │
│  │  │  Loop      │───▶│   /metrics :8080     │  │   │
│  │  │  (30s)     │    └──────────────────────┘  │   │
│  │  └─────┬──────┘                              │   │
│  │        │                                     │   │
│  │        ▼                                     │   │
│  │  ┌─────────────────────────────────────┐     │   │
│  │  │  client-go (in-cluster config)      │     │   │
│  │  │  - ClusterOperators                 │     │   │
│  │  │  - ClusterVersion                   │     │   │
│  │  │  - Nodes                            │     │   │
│  │  │  - Pods (system namespaces only)    │     │   │
│  │  └─────────────────────────────────────┘     │   │
│  └──────────────────────────────────────────────┘   │
│                                                     │
│  ┌──────────────────────────────────────────────┐   │
│  │  Service :8080 → /metrics                    │   │
│  └──────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```
