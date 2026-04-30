package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// AnnotationLifecycle owns the three-phase Grafana annotation sequence
// (start → end → region) and persists annotation IDs + tracked version
// as Kubernetes annotations on the workload.
type AnnotationLifecycle struct {
	Client  client.Client
	GClient AnnotationClient
}

// InitializeTracking stores the version without creating a Grafana annotation,
// because we don't know when the deployment actually happened.
func (l *AnnotationLifecycle) InitializeTracking(ctx context.Context, obj client.Object, version string) error {
	if err := l.patchAnnotations(ctx, obj, map[string]string{
		VersionAnnotation: version,
	}); err != nil {
		log.FromContext(ctx).Error(err, "Failed to initialize tracking")
		return err
	}
	return nil
}

// StartDeployment creates a start annotation and stores the annotation ID + new version.
func (l *AnnotationLifecycle) StartDeployment(
	ctx context.Context, obj client.Object, kind, version, imageRef, imageTag string,
) error {
	logger := log.FromContext(ctx)
	id, err := l.createAnnotation(ctx, kind, obj.GetName(), obj.GetNamespace(), imageTag, imageRef, "started")
	if err != nil {
		logger.Error(err, "Failed to create start annotation")
		return err
	}
	if err := l.patchAnnotations(ctx, obj, map[string]string{
		StartAnnotation:   strconv.FormatInt(id, 10),
		EndAnnotation:     "",
		VersionAnnotation: version,
	}); err != nil {
		logger.Error(err, "Failed to store start annotation")
		return err
	}
	logger.Info("Created start annotation", "kind", kind, "annotationID", id, "version", version)
	return nil
}

// CompleteDeployment creates an end annotation and patches the start annotation
// into a time-region. Idempotent — returns nil if already completed or if
// there is no start annotation to complete.
func (l *AnnotationLifecycle) CompleteDeployment(
	ctx context.Context, obj client.Object, kind, imageRef, imageTag string,
) error {
	annotations := obj.GetAnnotations()
	startID := annotations[StartAnnotation]
	if startID == "" || annotations[EndAnnotation] != "" {
		return nil
	}

	logger := log.FromContext(ctx)
	id, err := l.createAnnotation(ctx, kind, obj.GetName(), obj.GetNamespace(), imageTag, imageRef, "completed")
	if err != nil {
		logger.Error(err, "Failed to create end annotation")
		return err
	}
	if err := l.patchAnnotations(ctx, obj, map[string]string{
		EndAnnotation: strconv.FormatInt(id, 10),
	}); err != nil {
		logger.Error(err, "Failed to store end annotation")
		return err
	}
	if sid, err := strconv.ParseInt(startID, 10, 64); err == nil {
		tags := []string{
			"deploy",
			sanitizeForLog(obj.GetNamespace()),
			sanitizeForLog(obj.GetName()),
			sanitizeForLog(imageTag),
			"region", kind,
		}
		rctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		if err := l.GClient.UpdateAnnotationToRegion(rctx, sid, tags); err != nil {
			logger.Error(err, "Failed to update start annotation to region", "startAnnotationID", sid)
		}
	}
	logger.Info("Workload completed", "kind", kind, "endAnnotationID", id)
	return nil
}

// RecordDeletion creates a deletion annotation. No workload object is needed
// because the workload has already been deleted.
func (l *AnnotationLifecycle) RecordDeletion(ctx context.Context, kind, name, namespace string) error {
	if _, err := l.createAnnotation(ctx, kind, name, namespace, "", "", "deleted"); err != nil {
		log.FromContext(ctx).Error(err, "Failed to create deletion annotation")
		return err
	}
	log.FromContext(ctx).Info("Created deletion annotation", "kind", kind, "name", name, "namespace", namespace)
	return nil
}

// CleanupAnnotations removes all deployment-annotator annotations from a workload.
func (l *AnnotationLifecycle) CleanupAnnotations(ctx context.Context, obj client.Object) error {
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
	return l.patchAnnotations(ctx, obj, map[string]string{
		StartAnnotation: "", EndAnnotation: "", VersionAnnotation: "",
	})
}

// --- internal helpers (absorbed from helpers.go) ---

func (l *AnnotationLifecycle) createAnnotation(
	ctx context.Context, kind, name, namespace, imageTag, imageRef, eventType string,
) (int64, error) {
	sName := sanitizeForLog(name)
	sNS := sanitizeForLog(namespace)
	sTag := sanitizeForLog(imageTag)
	sRef := sanitizeForLog(imageRef)
	action := map[string]string{"started": "start", "completed": "end", "deleted": "delete"}[eventType]
	what := fmt.Sprintf("deploy-%s:%s", action, sName)
	if action == "" {
		what = fmt.Sprintf("deploy:%s", sName)
	}
	data := fmt.Sprintf("%s deployment %s", cases.Title(language.English).String(eventType), sRef)
	tags := []string{"deploy", sNS, sName, sTag, eventType, kind}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	return l.GClient.CreateAnnotation(ctx, what, tags, data)
}

func (l *AnnotationLifecycle) patchAnnotations(
	ctx context.Context, obj client.Object, annotations map[string]string,
) error {
	patch := map[string]interface{}{
		"metadata": map[string]interface{}{"annotations": annotations},
	}
	b, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("marshal patch: %w", err)
	}
	if err := l.Client.Patch(ctx, obj, client.RawPatch(client.Merge.Type(), b)); err != nil {
		return fmt.Errorf("patch annotations: %w", err)
	}
	return nil
}

// extractImageTag returns a human-friendly version tag from an image reference.
func extractImageTag(imageRef string) string {
	if at := strings.LastIndex(imageRef, "@"); at != -1 {
		digest := imageRef[at+1:]
		if colon := strings.Index(digest, ":"); colon != -1 && len(digest) > colon+7 {
			return digest[colon+1 : colon+8]
		}
		return digest
	}
	lastSlash := strings.LastIndex(imageRef, "/")
	lastColon := strings.LastIndex(imageRef, ":")
	if lastColon != -1 && lastColon > lastSlash {
		return imageRef[lastColon+1:]
	}
	return "latest"
}

// sanitizeForLog removes characters that could be used for log injection attacks.
func sanitizeForLog(input string) string {
	s := strings.ReplaceAll(input, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\t", "")
	return s
}
