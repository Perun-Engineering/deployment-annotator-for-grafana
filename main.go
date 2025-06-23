// Package main implements a Kubernetes Controller
// that automatically creates Grafana annotations for deployment events.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	goruntime "runtime"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// GrafanaIDAnnotation is the annotation key for storing Grafana annotation ID
	GrafanaIDAnnotation = "grafana.io/annotation-id"

	// GrafanaStartAnnotation tracks the start annotation ID
	GrafanaStartAnnotation = "grafana.io/start-annotation-id"

	// GrafanaEndAnnotation tracks the end annotation ID
	GrafanaEndAnnotation = "grafana.io/end-annotation-id"

	// GrafanaVersionAnnotation tracks the deployment version (generation + image tag)
	GrafanaVersionAnnotation = "grafana.io/tracked-version"

	// HTTPTimeoutSeconds is the timeout for HTTP requests in seconds
	HTTPTimeoutSeconds = 30

	// ImageTagSeparatorCount is the minimum number of parts when splitting image by ':'
	ImageTagSeparatorCount = 2

	// ControllerName is the name of this controller
	ControllerName = "grafana-annotation-controller"

	// DefaultMaxConcurrentReconciles is the default number of concurrent reconciles
	DefaultMaxConcurrentReconciles = 2

	// RequeueDelaySeconds is the delay in seconds for requeuing after version changes
	RequeueDelaySeconds = 5
)

var (
	// Build-time variables (set via ldflags)
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

// sanitizeForLog removes characters that could be used for log injection attacks
func sanitizeForLog(input string) string {
	// Remove newline characters that could be used for log injection
	sanitized := strings.ReplaceAll(input, "\n", "")
	sanitized = strings.ReplaceAll(sanitized, "\r", "")
	sanitized = strings.ReplaceAll(sanitized, "\t", "")
	return sanitized
}

// computeDeploymentVersion creates a version string from generation and image tag
func computeDeploymentVersion(deployment *appsv1.Deployment, imageTag string) string {
	return fmt.Sprintf("gen-%d-img-%s", deployment.Generation, imageTag)
}

// GrafanaAnnotation represents a Grafana annotation request
type GrafanaAnnotation struct {
	What string   `json:"what"`
	Tags []string `json:"tags"`
	Data string   `json:"data"`
	When int64    `json:"when"`
}

// GrafanaAnnotationResponse represents the response from creating an annotation
type GrafanaAnnotationResponse struct {
	ID int64 `json:"id"`
}

// GrafanaAnnotationPatch represents an annotation update request
type GrafanaAnnotationPatch struct {
	TimeEnd  int64    `json:"timeEnd"`
	IsRegion bool     `json:"isRegion"`
	Tags     []string `json:"tags"`
}

// DeploymentReconciler reconciles Deployment objects
type DeploymentReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	K8sClient  kubernetes.Interface
	GrafanaURL string
	GrafanaKey string
	HTTPClient *http.Client
}

// Reconcile handles deployment events and creates/updates Grafana annotations
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the deployment
	var deployment appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Deployment was deleted - handle deletion annotation
			logger.Info("Deployment deleted", "deployment", req.Name, "namespace", req.Namespace)
			return r.handleDeploymentDeletion(ctx, req.Name, req.Namespace)
		}
		logger.Error(err, "Failed to get deployment")
		return ctrl.Result{}, err
	}

	// Only process deployments in namespaces labeled for Grafana tracking
	namespace, err := r.K8sClient.CoreV1().Namespaces().Get(ctx, deployment.Namespace, metav1.GetOptions{})
	if err != nil {
		logger.Error(err, "Failed to get namespace", "namespace", deployment.Namespace)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Check if namespace has the deployment-annotator label
	if namespace.Labels == nil || namespace.Labels["deployment-annotator"] != "enabled" {
		// Skip deployments in non-tracked namespaces
		return ctrl.Result{}, nil
	}

	logger.Info("Processing deployment",
		"deployment", sanitizeForLog(deployment.Name),
		"namespace", sanitizeForLog(deployment.Namespace),
		"generation", deployment.Generation,
		"observedGeneration", deployment.Status.ObservedGeneration)

	return r.handleDeploymentEvent(ctx, &deployment)
}

// handleDeploymentEvent processes deployment create/update events
func (r *DeploymentReconciler) handleDeploymentEvent(
	ctx context.Context, deployment *appsv1.Deployment,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Extract container image information
	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		logger.Info("No containers found in deployment",
			"deployment", sanitizeForLog(deployment.Name),
			"namespace", sanitizeForLog(deployment.Namespace))
		return ctrl.Result{}, nil
	}

	imageRef := deployment.Spec.Template.Spec.Containers[0].Image
	imageTag := r.extractImageTag(imageRef)
	currentVersion := computeDeploymentVersion(deployment, imageTag)
	storedVersion := deployment.Annotations[GrafanaVersionAnnotation]
	startAnnotationID := deployment.Annotations[GrafanaStartAnnotation]

	// Handle new deployment version detection
	isNewDeploymentVersion := storedVersion == "" || storedVersion != currentVersion
	if isNewDeploymentVersion {
		return r.handleNewDeploymentVersion(ctx, deployment, currentVersion, imageRef, imageTag, startAnnotationID)
	}

	// Log scaling events
	logger.V(1).Info("Deployment event without version changes (likely scaling)",
		"deployment", sanitizeForLog(deployment.Name),
		"namespace", sanitizeForLog(deployment.Namespace),
		"version", currentVersion)

	// Handle deployment completion
	return r.handleDeploymentCompletion(
		ctx, deployment, currentVersion, imageRef, imageTag, storedVersion, startAnnotationID,
	)
}

// handleNewDeploymentVersion handles the logic for new deployment versions
func (r *DeploymentReconciler) handleNewDeploymentVersion(
	ctx context.Context, deployment *appsv1.Deployment, currentVersion, imageRef, imageTag, startAnnotationID string,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if startAnnotationID == "" {
		return r.createStartAnnotation(ctx, deployment, currentVersion, imageRef, imageTag)
	}

	// Version changed but we already have annotations - clear them
	if err := r.updateDeploymentAnnotations(ctx, deployment, map[string]string{
		GrafanaStartAnnotation:   "",
		GrafanaEndAnnotation:     "",
		GrafanaVersionAnnotation: currentVersion,
	}); err != nil {
		logger.Error(err, "Failed to clear annotations for version change",
			"deployment", sanitizeForLog(deployment.Name),
			"namespace", sanitizeForLog(deployment.Namespace))
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	logger.Info("Deployment version changed, cleared previous annotations",
		"deployment", sanitizeForLog(deployment.Name),
		"namespace", sanitizeForLog(deployment.Namespace),
		"newVersion", currentVersion)

	return ctrl.Result{RequeueAfter: time.Second * RequeueDelaySeconds}, nil
}

// createStartAnnotation creates a new start annotation for a deployment
func (r *DeploymentReconciler) createStartAnnotation(
	ctx context.Context, deployment *appsv1.Deployment, currentVersion, imageRef, imageTag string,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	annotationID, err := r.createGrafanaAnnotation(deployment.Name, deployment.Namespace, imageTag, imageRef, "started")
	if err != nil {
		logger.Error(err, "Failed to create start annotation",
			"deployment", sanitizeForLog(deployment.Name),
			"namespace", sanitizeForLog(deployment.Namespace))
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	if err := r.updateDeploymentAnnotations(ctx, deployment, map[string]string{
		GrafanaStartAnnotation:   strconv.FormatInt(annotationID, 10),
		GrafanaVersionAnnotation: currentVersion,
	}); err != nil {
		logger.Error(err, "Failed to store start annotation ID and version",
			"deployment", sanitizeForLog(deployment.Name),
			"namespace", sanitizeForLog(deployment.Namespace))
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	logger.Info("Created deployment start annotation",
		"deployment", sanitizeForLog(deployment.Name),
		"namespace", sanitizeForLog(deployment.Namespace),
		"annotationID", annotationID,
		"version", currentVersion)

	return ctrl.Result{}, nil
}

// handleDeploymentCompletion handles the logic when a deployment is ready
func (r *DeploymentReconciler) handleDeploymentCompletion(
	ctx context.Context,
	deployment *appsv1.Deployment,
	currentVersion, imageRef, imageTag, storedVersion, startAnnotationID string,
) (ctrl.Result, error) {
	// Check if deployment is ready and we're tracking it
	if !r.isDeploymentReady(deployment) || storedVersion != currentVersion {
		return ctrl.Result{}, nil
	}

	endAnnotationID := deployment.Annotations[GrafanaEndAnnotation]
	if endAnnotationID != "" || startAnnotationID == "" {
		return ctrl.Result{}, nil
	}

	return r.createEndAnnotation(ctx, deployment, imageRef, imageTag, startAnnotationID)
}

// createEndAnnotation creates an end annotation and updates the start annotation to a region
func (r *DeploymentReconciler) createEndAnnotation(
	ctx context.Context, deployment *appsv1.Deployment, imageRef, imageTag, startAnnotationID string,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	annotationID, err := r.createGrafanaAnnotation(
		deployment.Name, deployment.Namespace, imageTag, imageRef, "completed",
	)
	if err != nil {
		logger.Error(err, "Failed to create end annotation",
			"deployment", sanitizeForLog(deployment.Name),
			"namespace", sanitizeForLog(deployment.Namespace))
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	if err := r.updateDeploymentAnnotation(
		ctx, deployment, GrafanaEndAnnotation, strconv.FormatInt(annotationID, 10),
	); err != nil {
		logger.Error(err, "Failed to store end annotation ID",
			"deployment", sanitizeForLog(deployment.Name),
			"namespace", sanitizeForLog(deployment.Namespace))
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	// Update start annotation to create a time region
	if startID, err := strconv.ParseInt(startAnnotationID, 10, 64); err == nil {
		if err := r.updateGrafanaAnnotation(startID, deployment.Name, deployment.Namespace, imageTag); err != nil {
			logger.Error(err, "Failed to update start annotation to region", "startAnnotationID", startID)
		}
	}

	logger.Info("Deployment completed",
		"deployment", sanitizeForLog(deployment.Name),
		"namespace", sanitizeForLog(deployment.Namespace),
		"endAnnotationID", annotationID)

	return ctrl.Result{}, nil
}

// handleDeploymentDeletion processes deployment deletion events
func (r *DeploymentReconciler) handleDeploymentDeletion(
	ctx context.Context, name, namespace string,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Create deletion annotation
	_, err := r.createGrafanaAnnotation(name, namespace, "", "", "deleted")
	if err != nil {
		logger.Error(err, "Failed to create deletion annotation",
			"deployment", sanitizeForLog(name),
			"namespace", sanitizeForLog(namespace))
		return ctrl.Result{}, err
	}

	logger.Info("Created deployment deletion annotation",
		"deployment", sanitizeForLog(name),
		"namespace", sanitizeForLog(namespace))

	return ctrl.Result{}, nil
}

// isDeploymentReady checks if a deployment is ready
func (r *DeploymentReconciler) isDeploymentReady(deployment *appsv1.Deployment) bool {
	// Check if the deployment has the desired number of ready replicas
	return deployment.Status.ReadyReplicas > 0 &&
		deployment.Status.ReadyReplicas == deployment.Status.Replicas &&
		deployment.Status.ObservedGeneration == deployment.Generation
}

// extractImageTag extracts the tag from a container image reference
func (r *DeploymentReconciler) extractImageTag(imageRef string) string {
	parts := strings.Split(imageRef, ":")
	if len(parts) < ImageTagSeparatorCount {
		return "latest"
	}
	return parts[len(parts)-1]
}

// createGrafanaAnnotation creates a new annotation in Grafana
func (r *DeploymentReconciler) createGrafanaAnnotation(
	deploymentName, namespace, imageTag, imageRef, eventType string,
) (int64, error) {
	// Sanitize all user-provided values for Grafana API
	sanitizedName := sanitizeForLog(deploymentName)
	sanitizedNamespace := sanitizeForLog(namespace)
	sanitizedTag := sanitizeForLog(imageTag)
	sanitizedRef := sanitizeForLog(imageRef)

	var what, data string
	switch eventType {
	case "started":
		what = fmt.Sprintf("deploy-start:%s", sanitizedName)
		data = fmt.Sprintf("Started deployment %s", sanitizedRef)
	case "completed":
		what = fmt.Sprintf("deploy-end:%s", sanitizedName)
		data = fmt.Sprintf("Completed deployment %s", sanitizedRef)
	case "deleted":
		what = fmt.Sprintf("deploy-delete:%s", sanitizedName)
		data = fmt.Sprintf("Deleted deployment %s", sanitizedName)
	default:
		what = fmt.Sprintf("deploy:%s", sanitizedName)
		data = fmt.Sprintf("Deployment event: %s", eventType)
	}

	annotation := GrafanaAnnotation{
		What: what,
		Tags: []string{"deploy", sanitizedNamespace, sanitizedName, sanitizedTag, eventType},
		Data: data,
		When: time.Now().Unix(),
	}

	jsonData, err := json.Marshal(annotation)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal annotation: %w", err)
	}

	url := fmt.Sprintf("%s/api/annotations/graphite", r.GrafanaURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.GrafanaKey))

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, fmt.Errorf("grafana API returned status %d and failed to read response body: %w", resp.StatusCode, err)
		}
		return 0, fmt.Errorf("grafana API returned status %d: %s", resp.StatusCode, sanitizeForLog(string(body)))
	}

	var response GrafanaAnnotationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.ID, nil
}

// updateGrafanaAnnotation updates an existing annotation to create a time region
func (r *DeploymentReconciler) updateGrafanaAnnotation(
	annotationID int64, deploymentName, namespace, imageTag string,
) error {
	// Sanitize all user-provided values for Grafana API
	sanitizedName := sanitizeForLog(deploymentName)
	sanitizedNamespace := sanitizeForLog(namespace)
	sanitizedTag := sanitizeForLog(imageTag)

	patch := GrafanaAnnotationPatch{
		TimeEnd:  time.Now().UnixMilli(),
		IsRegion: true,
		Tags:     []string{"deploy", sanitizedNamespace, sanitizedName, sanitizedTag, "region"},
	}

	jsonData, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	url := fmt.Sprintf("%s/api/annotations/%d", r.GrafanaURL, annotationID)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.GrafanaKey))

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("grafana API returned status %d and failed to read response body: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("grafana API returned status %d: %s", resp.StatusCode, sanitizeForLog(string(body)))
	}

	return nil
}

// updateDeploymentAnnotation updates a deployment's annotation
func (r *DeploymentReconciler) updateDeploymentAnnotation(
	ctx context.Context, deployment *appsv1.Deployment, key, value string,
) error {
	return r.updateDeploymentAnnotations(ctx, deployment, map[string]string{key: value})
}

// updateDeploymentAnnotations updates multiple deployment annotations in a single operation
func (r *DeploymentReconciler) updateDeploymentAnnotations(
	ctx context.Context, deployment *appsv1.Deployment, annotations map[string]string,
) error {
	// Get the latest version of the deployment
	var current appsv1.Deployment
	if err := r.Get(ctx, client.ObjectKeyFromObject(deployment), &current); err != nil {
		return fmt.Errorf("failed to get current deployment: %w", err)
	}

	// Update annotations
	if current.Annotations == nil {
		current.Annotations = make(map[string]string)
	}
	for key, value := range annotations {
		current.Annotations[key] = value
	}

	// Update the deployment
	if err := r.Update(ctx, &current); err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	return nil
}

// mapReplicaSetToDeployment maps ReplicaSet events to their parent Deployment
func (r *DeploymentReconciler) mapReplicaSetToDeployment(_ context.Context, obj client.Object) []reconcile.Request {
	replicaSet, ok := obj.(*appsv1.ReplicaSet)
	if !ok {
		return nil
	}

	// Find the parent Deployment from owner references
	for _, owner := range replicaSet.OwnerReferences {
		if owner.Kind == "Deployment" && owner.APIVersion == "apps/v1" {
			return []reconcile.Request{
				{
					NamespacedName: client.ObjectKey{
						Name:      owner.Name,
						Namespace: replicaSet.Namespace,
					},
				},
			}
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Watches(
			&appsv1.ReplicaSet{},
			handler.EnqueueRequestsFromMapFunc(r.mapReplicaSetToDeployment),
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: DefaultMaxConcurrentReconciles,
		}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func main() {
	// Initialize scheme
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		panic(err)
	}

	// Setup logging - use zap directly to avoid stack overflow
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	logger := ctrl.Log.WithName("main")

	// Print version information
	logger.Info("Grafana Annotation Controller")
	logger.Info("Version info", "version", version, "commit", commit, "buildTime", buildTime)
	logger.Info("Runtime info", "goVersion", goruntime.Version(), "os", goruntime.GOOS, "arch", goruntime.GOARCH)

	// Get configuration from environment
	grafanaURL := os.Getenv("GRAFANA_URL")
	if grafanaURL == "" {
		logger.Error(nil, "GRAFANA_URL environment variable is required")
		os.Exit(1)
	}

	grafanaKey := os.Getenv("GRAFANA_API_KEY")
	if grafanaKey == "" {
		logger.Error(nil, "GRAFANA_API_KEY environment variable is required")
		os.Exit(1)
	}

	// Create manager with memory-optimized configuration
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Logger: ctrl.Log.WithName("manager"),
		Metrics: server.Options{
			BindAddress: ":8081",
		},
		HealthProbeBindAddress: ":8080",
		LeaderElection:         false,
	})
	if err != nil {
		logger.Error(err, "Failed to create manager")
		os.Exit(1)
	}

	// Create Kubernetes client
	k8sClient, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		logger.Error(err, "Failed to create Kubernetes client")
		os.Exit(1)
	}

	// Setup reconciler
	reconciler := &DeploymentReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		K8sClient:  k8sClient,
		GrafanaURL: strings.TrimSuffix(grafanaURL, "/"),
		GrafanaKey: grafanaKey,
		HTTPClient: &http.Client{Timeout: HTTPTimeoutSeconds * time.Second},
	}

	if err = reconciler.SetupWithManager(mgr); err != nil {
		logger.Error(err, "Failed to setup controller")
		os.Exit(1)
	}

	// Add health checks
	if err := mgr.AddHealthzCheck("healthz", func(_ *http.Request) error {
		return nil
	}); err != nil {
		logger.Error(err, "Failed to setup health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", func(_ *http.Request) error {
		return nil
	}); err != nil {
		logger.Error(err, "Failed to setup ready check")
		os.Exit(1)
	}

	// Start the manager
	logger.Info("Starting controller", "controllerName", ControllerName)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Error(err, "Failed to start manager")
		os.Exit(1)
	}
}
