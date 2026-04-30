# CONTEXT — Deployment Annotator for Grafana

## Domain vocabulary

- **Workload** — a Kubernetes `apps/v1` resource that runs pods: Deployment, StatefulSet, or DaemonSet. The controller treats all three uniformly through a `WorkloadAdapter`.
- **Tracked namespace** — a namespace carrying the label `deployment-annotator=enabled`. The controller only processes workloads in tracked namespaces.
- **Version** — an opaque string that identifies a workload's current spec. Built from Kubernetes generation + container image tag (or pod-template-hash for Deployments). Two reconcile events with the same version are treated as no-ops (scaling, rescheduling).
- **Annotation lifecycle** — the three-phase Grafana annotation sequence for a workload change: **start** (spec changed) → **end** (rollout complete) → **region** (start annotation patched into a time-region spanning start→end). Owned by the concrete `AnnotationLifecycle` struct, which persists annotation IDs and tracked version as Kubernetes annotations on the workload. The reconciler delegates all Grafana interaction and annotation-state bookkeeping to this struct.
- **Adapter** — a small interface (`WorkloadAdapter`) that captures all differences between workload kinds: version computation, readiness check, spec/status extraction, list unpacking, and whether completion is detected via status changes or a secondary watch. No code outside the adapter type-switches on concrete workload types.
- **AnnotationClient** — the seam between the reconciler and the annotation backend. Defined in `internal/controller` (consumer-side). `grafana.Client` satisfies it; tests supply a fake. Two methods: `CreateAnnotation` and `UpdateAnnotationToRegion`.
- **Completion detection** — how the controller learns a rollout finished. Deployments use ReplicaSet events (secondary watch). StatefulSets and DaemonSets use their own status-change predicates.

## Package layout

| Package | Responsibility |
|---|---|
| `main` | Wiring only: config, clients, manager, adapter registration |
| `internal/controller` | `WorkloadReconciler` (orchestration: fetch, namespace check, version, readiness) + `AnnotationLifecycle` (annotation state machine: start/complete/delete/initialize + persistence) + `WorkloadAdapter` interface + three adapter implementations + predicates and pure utilities |
| `internal/grafana` | HTTP client for Grafana annotation API |
