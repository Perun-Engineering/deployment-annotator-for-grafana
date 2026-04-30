# CONTEXT — Deployment Annotator for Grafana

## Domain vocabulary

- **Workload** — a Kubernetes `apps/v1` resource that runs pods: Deployment, StatefulSet, or DaemonSet. The controller treats all three uniformly through a `WorkloadAdapter`.
- **Tracked namespace** — a namespace carrying the label `deployment-annotator=enabled`. The controller only processes workloads in tracked namespaces.
- **Version** — an opaque string that identifies a workload's current spec. Built from Kubernetes generation + container image tag (or pod-template-hash for Deployments). Two reconcile events with the same version are treated as no-ops (scaling, rescheduling).
- **Annotation lifecycle** — the three-phase Grafana annotation sequence for a workload change: **start** (spec changed) → **end** (rollout complete) → **region** (start annotation patched into a time-region spanning start→end).
- **Adapter** — a small interface (`WorkloadAdapter`) that captures the differences between workload kinds: version computation, readiness check, and whether completion is detected via status changes or a secondary watch.
- **AnnotationClient** — the seam between the reconciler and the annotation backend. Defined in `internal/controller` (consumer-side). `grafana.Client` satisfies it; tests supply a fake. Two methods: `CreateAnnotation` and `UpdateAnnotationToRegion`.
- **Completion detection** — how the controller learns a rollout finished. Deployments use ReplicaSet events (secondary watch). StatefulSets and DaemonSets use their own status-change predicates.

## Package layout

| Package | Responsibility |
|---|---|
| `main` | Wiring only: config, clients, manager, adapter registration |
| `internal/controller` | `WorkloadReconciler` + `WorkloadAdapter` interface + three adapter implementations + shared helpers (including log sanitization and image-tag extraction) |
| `internal/grafana` | HTTP client for Grafana annotation API |
