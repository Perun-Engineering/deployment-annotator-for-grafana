# 1. WorkloadAdapter and AnnotationClient seams

Date: 2026-05-31

## Status

Accepted

## Context

The controller annotates Grafana on changes to three Kubernetes workload kinds — Deployment, StatefulSet, and DaemonSet. These kinds differ in how a version is computed, how readiness is checked, how spec/status is extracted, how lists are unpacked, and how rollout completion is detected (Deployments observe ReplicaSet events via a secondary watch; StatefulSets and DaemonSets use their own status-change predicates).

Left unmanaged, those differences leak into the reconciler as type switches on concrete workload types, and the reconciler also ends up owning direct HTTP calls to Grafana plus the bookkeeping of annotation IDs. Both couplings make the orchestration logic hard to read and hard to test.

## Decision

Two seams isolate the variation:

- **`WorkloadAdapter`** — an interface capturing every per-kind difference (version computation, readiness, spec/status extraction, list unpacking, completion-detection strategy). No code outside the adapter implementations type-switches on concrete workload types. Adapters are registered in `main`.
- **`AnnotationClient`** — a two-method consumer-side interface (`CreateAnnotation`, `UpdateAnnotationToRegion`) defined in `internal/controller`. The concrete `grafana.Client` satisfies it; tests supply a fake.

The `WorkloadReconciler` keeps orchestration only (fetch, namespace check, version, readiness) and delegates all Grafana interaction and annotation-state bookkeeping to the concrete `AnnotationLifecycle` struct, which persists annotation IDs and the tracked version as Kubernetes annotations on the workload.

## Consequences

- Adding a new workload kind means writing one adapter, not editing the reconciler.
- The reconciler is unit-testable against fakes with no Grafana or live cluster.
- The consumer-side interface keeps `internal/controller` free of a hard dependency on `internal/grafana`.
- Adapter implementations concentrate the per-kind complexity; they are the place to look when a kind behaves unexpectedly.

See `CONTEXT.md` for the domain vocabulary (Workload, Adapter, AnnotationClient, Annotation lifecycle, Completion detection).
