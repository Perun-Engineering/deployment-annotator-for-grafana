package controller

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	StartAnnotation   = "deployment-annotator.io/start-annotation-id"
	EndAnnotation     = "deployment-annotator.io/end-annotation-id"
	VersionAnnotation = "deployment-annotator.io/tracked-version"

	DefaultMaxConcurrentReconciles = 2
)

// AnnotationClient is the seam between the reconciler and the annotation backend.
// grafana.Client satisfies this interface; tests can supply a fake.
type AnnotationClient interface {
	CreateAnnotation(ctx context.Context, what string, tags []string, data string) (int64, error)
	UpdateAnnotationToRegion(ctx context.Context, id int64, tags []string) error
}

// WorkloadReconciler reconciles any workload type via its WorkloadAdapter.
// It handles Kubernetes fetching, namespace checks, version computation, and
// readiness detection. All annotation lifecycle logic is delegated to Lifecycle.
type WorkloadReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Adapter   WorkloadAdapter
	Lifecycle *AnnotationLifecycle
}

func (r *WorkloadReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	kind := r.Adapter.Kind()

	obj := r.Adapter.NewObject()
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return r.handleDeletion(ctx, req, kind)
		}
		logger.Error(err, "Failed to get workload", "kind", kind)
		return ctrl.Result{}, err
	}

	var ns corev1.Namespace
	if err := r.Get(ctx, client.ObjectKey{Name: obj.GetNamespace()}, &ns); err != nil {
		logger.Error(err, "Failed to get namespace", "namespace", obj.GetNamespace())
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
	if ns.Labels["deployment-annotator"] != "enabled" {
		return ctrl.Result{}, nil
	}

	logger.V(1).Info("Processing workload",
		"kind", kind,
		"name", sanitizeForLog(obj.GetName()),
		"namespace", sanitizeForLog(obj.GetNamespace()),
		"generation", obj.GetGeneration())

	return r.handleEvent(ctx, obj, kind)
}

func (r *WorkloadReconciler) handleEvent(ctx context.Context, obj client.Object, kind string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	name := sanitizeForLog(obj.GetName())
	ns := sanitizeForLog(obj.GetNamespace())

	imageRef := r.Adapter.ContainerImage(obj)
	if imageRef == "" {
		logger.Info("No containers found", "kind", kind, "name", name, "namespace", ns)
		return ctrl.Result{}, nil
	}

	imageTag := extractImageTag(imageRef)
	currentVersion := r.Adapter.ComputeVersion(ctx, r.Client, obj, imageTag)
	storedVersion := obj.GetAnnotations()[VersionAnnotation]

	if storedVersion == "" {
		logger.Info("Initializing tracking", "kind", kind, "name", name, "namespace", ns, "version", currentVersion)
		if err := r.Lifecycle.InitializeTracking(ctx, obj, currentVersion); err != nil {
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}
		return ctrl.Result{}, nil
	}

	if storedVersion != currentVersion {
		logger.Info("Version changed", "kind", kind, "name", name, "namespace", ns,
			"oldVersion", storedVersion, "newVersion", currentVersion)
		if err := r.Lifecycle.StartDeployment(ctx, obj, kind, currentVersion, imageRef, imageTag); err != nil {
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}
		return ctrl.Result{}, nil
	}

	logger.V(1).Info("No version change", "kind", kind, "name", name, "namespace", ns, "version", currentVersion)
	if r.Adapter.IsReady(obj) {
		if err := r.Lifecycle.CompleteDeployment(ctx, obj, kind, imageRef, imageTag); err != nil {
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *WorkloadReconciler) handleDeletion(ctx context.Context, req ctrl.Request, kind string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var ns corev1.Namespace
	if err := r.Get(ctx, client.ObjectKey{Name: req.Namespace}, &ns); err != nil {
		logger.Error(err, "Failed to get namespace for deletion check")
		return ctrl.Result{}, nil
	}
	if ns.Labels["deployment-annotator"] != "enabled" {
		logger.V(1).Info("Ignoring deletion in unlabeled namespace", "kind", kind, "name", req.Name)
		return ctrl.Result{}, nil
	}

	if err := r.Lifecycle.RecordDeletion(ctx, kind, req.Name, req.Namespace); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *WorkloadReconciler) mapNamespaceToWorkloads(ctx context.Context, obj client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)
	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return nil
	}

	enabled := ns.Labels["deployment-annotator"] == "enabled"
	list := r.Adapter.NewObjectList()
	if err := r.List(ctx, list, client.InNamespace(ns.Name)); err != nil {
		logger.Error(err, "Failed to list workloads", "kind", r.Adapter.Kind(), "namespace", ns.Name)
		return nil
	}

	items := r.Adapter.ExtractItems(list)

	if !enabled {
		logger.Info("Namespace label removed, cleaning up",
			"kind", r.Adapter.Kind(), "namespace", ns.Name, "count", len(items))
		for _, item := range items {
			_ = r.Lifecycle.CleanupAnnotations(ctx, item)
		}
		return nil
	}

	requests := make([]reconcile.Request, 0, len(items))
	for _, item := range items {
		requests = append(requests, reconcile.Request{
			NamespacedName: client.ObjectKeyFromObject(item),
		})
	}
	logger.Info("Namespace labeled, enqueuing workloads",
		"kind", r.Adapter.Kind(), "namespace", ns.Name, "count", len(requests))
	return requests
}

func mapReplicaSetToDeployment(_ context.Context, obj client.Object) []reconcile.Request {
	rs, ok := obj.(*appsv1.ReplicaSet)
	if !ok {
		return nil
	}
	for _, owner := range rs.OwnerReferences {
		if owner.Kind == "Deployment" && owner.APIVersion == "apps/v1" {
			return []reconcile.Request{{
				NamespacedName: client.ObjectKey{Name: owner.Name, Namespace: rs.Namespace},
			}}
		}
	}
	return nil
}

func (r *WorkloadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	b := ctrl.NewControllerManagedBy(mgr).
		For(r.Adapter.NewObject(), builder.WithPredicates(specChangedPredicate(r.Adapter))).
		Watches(&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(r.mapNamespaceToWorkloads),
			builder.WithPredicates(namespaceLabelChangedPredicate())).
		WithOptions(controller.Options{MaxConcurrentReconciles: DefaultMaxConcurrentReconciles})

	if r.Adapter.Kind() == "deployment" {
		b = b.Watches(&appsv1.ReplicaSet{}, handler.EnqueueRequestsFromMapFunc(mapReplicaSetToDeployment))
	}

	return b.Complete(r)
}

var _ reconcile.Reconciler = (*WorkloadReconciler)(nil)
