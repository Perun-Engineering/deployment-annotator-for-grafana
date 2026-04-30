package controller

import (
	"context"
	"strconv"
	"time"

	"github.com/perun-engineering/deployment-annotator-for-grafana/internal/grafana"
	apputil "github.com/perun-engineering/deployment-annotator-for-grafana/internal/util"
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

// WorkloadReconciler reconciles any workload type via its WorkloadAdapter.
type WorkloadReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	GClient *grafana.Client
	Adapter WorkloadAdapter
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

	// Check namespace label
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
		"name", apputil.SanitizeForLog(obj.GetName()),
		"namespace", apputil.SanitizeForLog(obj.GetNamespace()),
		"generation", obj.GetGeneration())

	return r.handleEvent(ctx, obj, kind)
}

func (r *WorkloadReconciler) handleEvent(ctx context.Context, obj client.Object, kind string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	name := apputil.SanitizeForLog(obj.GetName())
	ns := apputil.SanitizeForLog(obj.GetNamespace())

	imageRef := r.Adapter.ContainerImage(obj)
	if imageRef == "" {
		logger.Info("No containers found", "kind", kind, "name", name, "namespace", ns)
		return ctrl.Result{}, nil
	}

	imageTag := apputil.ExtractImageTag(imageRef)
	currentVersion := r.Adapter.ComputeVersion(ctx, r.Client, obj, imageTag)
	annotations := obj.GetAnnotations()
	storedVersion := annotations[VersionAnnotation]
	startID := annotations[StartAnnotation]

	if storedVersion == "" {
		logger.Info("Initializing tracking", "kind", kind, "name", name, "namespace", ns, "version", currentVersion)
		return r.initializeTracking(ctx, obj, currentVersion)
	}

	if storedVersion != currentVersion {
		logger.Info("Version changed", "kind", kind, "name", name, "namespace", ns,
			"oldVersion", storedVersion, "newVersion", currentVersion)
		return r.handleNewVersion(ctx, obj, kind, currentVersion, imageRef, imageTag, startID)
	}

	logger.V(1).Info("No version change", "kind", kind, "name", name, "namespace", ns, "version", currentVersion)
	return r.handleCompletion(ctx, obj, kind, currentVersion, imageRef, imageTag, storedVersion, startID)
}

func (r *WorkloadReconciler) initializeTracking(
	ctx context.Context, obj client.Object, version string,
) (ctrl.Result, error) {
	if err := patchAnnotations(ctx, r.Client, obj, map[string]string{
		VersionAnnotation: version,
	}); err != nil {
		log.FromContext(ctx).Error(err, "Failed to initialize tracking")
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}
	return ctrl.Result{}, nil
}

func (r *WorkloadReconciler) handleNewVersion(
	ctx context.Context, obj client.Object,
	kind, version, imageRef, imageTag, startID string,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if startID == "" {
		return r.createStartAnnotation(ctx, obj, kind, version, imageRef, imageTag)
	}

	newID, err := createAnnotation(ctx, r.GClient, kind, obj.GetName(), obj.GetNamespace(), imageTag, imageRef, "started")
	if err != nil {
		logger.Error(err, "Failed to create start annotation for version change")
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}
	if err := patchAnnotations(ctx, r.Client, obj, map[string]string{
		StartAnnotation:   strconv.FormatInt(newID, 10),
		EndAnnotation:     "",
		VersionAnnotation: version,
	}); err != nil {
		logger.Error(err, "Failed to update annotations for version change")
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}
	logger.Info("Updated annotations for version change", "kind", kind, "newStartAnnotationID", newID)
	return ctrl.Result{}, nil
}

func (r *WorkloadReconciler) createStartAnnotation(
	ctx context.Context, obj client.Object,
	kind, version, imageRef, imageTag string,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	id, err := createAnnotation(ctx, r.GClient, kind, obj.GetName(), obj.GetNamespace(), imageTag, imageRef, "started")
	if err != nil {
		logger.Error(err, "Failed to create start annotation")
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}
	if err := patchAnnotations(ctx, r.Client, obj, map[string]string{
		StartAnnotation:   strconv.FormatInt(id, 10),
		VersionAnnotation: version,
	}); err != nil {
		logger.Error(err, "Failed to store start annotation")
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}
	logger.Info("Created start annotation", "kind", kind, "annotationID", id, "version", version)
	return ctrl.Result{}, nil
}

func (r *WorkloadReconciler) handleCompletion(
	ctx context.Context, obj client.Object,
	kind, currentVersion, imageRef, imageTag, storedVersion, startID string,
) (ctrl.Result, error) {
	if !r.Adapter.IsReady(obj) || storedVersion != currentVersion {
		return ctrl.Result{}, nil
	}
	annotations := obj.GetAnnotations()
	if annotations[EndAnnotation] != "" || startID == "" {
		return ctrl.Result{}, nil
	}
	return r.createEndAnnotation(ctx, obj, kind, imageRef, imageTag, startID)
}

func (r *WorkloadReconciler) createEndAnnotation(
	ctx context.Context, obj client.Object,
	kind, imageRef, imageTag, startID string,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	id, err := createAnnotation(ctx, r.GClient, kind, obj.GetName(), obj.GetNamespace(), imageTag, imageRef, "completed")
	if err != nil {
		logger.Error(err, "Failed to create end annotation")
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}
	if err := patchAnnotations(ctx, r.Client, obj, map[string]string{
		EndAnnotation: strconv.FormatInt(id, 10),
	}); err != nil {
		logger.Error(err, "Failed to store end annotation")
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}
	if sid, err := strconv.ParseInt(startID, 10, 64); err == nil {
		err := updateAnnotationToRegion(
			ctx, r.GClient, sid, kind, obj.GetName(), obj.GetNamespace(), imageTag,
		)
		if err != nil {
			logger.Error(err, "Failed to update start annotation to region", "startAnnotationID", sid)
		}
	}
	logger.Info("Workload completed", "kind", kind, "endAnnotationID", id)
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

	if _, err := createAnnotation(ctx, r.GClient, kind, req.Name, req.Namespace, "", "", "deleted"); err != nil {
		logger.Error(err, "Failed to create deletion annotation")
		return ctrl.Result{}, err
	}
	logger.Info("Created deletion annotation", "kind", kind, "name", req.Name, "namespace", req.Namespace)
	return ctrl.Result{}, nil
}

// mapNamespaceToWorkloads enqueues all workloads in a namespace when its label changes,
// or cleans up annotations when the label is removed.
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

	items := extractItems(list)

	if !enabled {
		logger.Info("Namespace label removed, cleaning up",
			"kind", r.Adapter.Kind(), "namespace", ns.Name, "count", len(items))
		for _, item := range items {
			_ = r.cleanupAnnotations(ctx, item)
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

func (r *WorkloadReconciler) cleanupAnnotations(ctx context.Context, obj client.Object) error {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return nil
	}
	has := false
	for _, k := range []string{StartAnnotation, EndAnnotation, VersionAnnotation} {
		if _, ok := annotations[k]; ok {
			has = true
			break
		}
	}
	if !has {
		return nil
	}
	if err := patchAnnotations(ctx, r.Client, obj, map[string]string{
		StartAnnotation: "", EndAnnotation: "", VersionAnnotation: "",
	}); err != nil {
		return err
	}
	log.FromContext(ctx).Info("Cleaned up annotations",
		"kind", r.Adapter.Kind(),
		"name", apputil.SanitizeForLog(obj.GetName()),
		"namespace", apputil.SanitizeForLog(obj.GetNamespace()))
	return nil
}

// mapReplicaSetToDeployment maps ReplicaSet events to their parent Deployment.
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

// SetupWithManager registers the controller with the manager.
func (r *WorkloadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	b := ctrl.NewControllerManagedBy(mgr).
		For(r.Adapter.NewObject(), builder.WithPredicates(specChangedPredicate(r.Adapter.WatchesStatus()))).
		Watches(&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(r.mapNamespaceToWorkloads),
			builder.WithPredicates(namespaceLabelChangedPredicate())).
		WithOptions(controller.Options{MaxConcurrentReconciles: DefaultMaxConcurrentReconciles})

	// Deployments detect completion via ReplicaSet events
	if r.Adapter.Kind() == "deployment" {
		b = b.Watches(&appsv1.ReplicaSet{}, handler.EnqueueRequestsFromMapFunc(mapReplicaSetToDeployment))
	}

	return b.Complete(r)
}

// extractItems pulls individual objects from a typed ObjectList.
func extractItems(list client.ObjectList) []client.Object {
	switch l := list.(type) {
	case *appsv1.DeploymentList:
		out := make([]client.Object, len(l.Items))
		for i := range l.Items {
			out[i] = &l.Items[i]
		}
		return out
	case *appsv1.StatefulSetList:
		out := make([]client.Object, len(l.Items))
		for i := range l.Items {
			out[i] = &l.Items[i]
		}
		return out
	case *appsv1.DaemonSetList:
		out := make([]client.Object, len(l.Items))
		for i := range l.Items {
			out[i] = &l.Items[i]
		}
		return out
	}
	return nil
}

var _ reconcile.Reconciler = (*WorkloadReconciler)(nil)
