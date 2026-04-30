package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// sanitizeForLog removes characters that could be used for log injection attacks.
func sanitizeForLog(input string) string {
	s := strings.ReplaceAll(input, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\t", "")
	return s
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

// createAnnotation builds and sends a Grafana annotation for a workload event.
func createAnnotation(
	ctx context.Context, gc AnnotationClient,
	kind, name, namespace, imageTag, imageRef, eventType string,
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
	return gc.CreateAnnotation(ctx, what, tags, data)
}

// updateAnnotationToRegion patches a start annotation into a time-region.
func updateAnnotationToRegion(
	ctx context.Context, gc AnnotationClient,
	id int64, kind, name, namespace, imageTag string,
) error {
	tags := []string{
		"deploy",
		sanitizeForLog(namespace),
		sanitizeForLog(name),
		sanitizeForLog(imageTag),
		"region", kind,
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	return gc.UpdateAnnotationToRegion(ctx, id, tags)
}

// patchAnnotations applies a strategic-merge patch to obj's annotations.
func patchAnnotations(ctx context.Context, c client.Client, obj client.Object, annotations map[string]string) error {
	patch := map[string]interface{}{
		"metadata": map[string]interface{}{"annotations": annotations},
	}
	b, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("marshal patch: %w", err)
	}
	if err := c.Patch(ctx, obj, client.RawPatch(client.Merge.Type(), b)); err != nil {
		return fmt.Errorf("patch annotations: %w", err)
	}
	return nil
}

// specChangedPredicate triggers on spec changes. When includeStatus is true
// it also triggers on status changes (used by StatefulSet/DaemonSet which
// detect completion via their own status, not via a secondary watch).
func specChangedPredicate(includeStatus bool) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(event.CreateEvent) bool { return true },
		DeleteFunc:  func(event.DeleteEvent) bool { return true },
		GenericFunc: func(event.GenericEvent) bool { return true },
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldSpec, _ := json.Marshal(specOf(e.ObjectOld))
			newSpec, _ := json.Marshal(specOf(e.ObjectNew))
			if !bytes.Equal(oldSpec, newSpec) {
				return true
			}
			if !includeStatus {
				return false
			}
			oldStatus, _ := json.Marshal(statusOf(e.ObjectOld))
			newStatus, _ := json.Marshal(statusOf(e.ObjectNew))
			return !bytes.Equal(oldStatus, newStatus)
		},
	}
}

// namespaceLabelChangedPredicate triggers when the deployment-annotator label toggles.
func namespaceLabelChangedPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			ns, ok := e.Object.(*corev1.Namespace)
			return ok && ns.Labels["deployment-annotator"] == "enabled"
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldNs, ok1 := e.ObjectOld.(*corev1.Namespace)
			newNs, ok2 := e.ObjectNew.(*corev1.Namespace)
			if !ok1 || !ok2 {
				return false
			}
			was := oldNs.Labels["deployment-annotator"] == "enabled"
			now := newNs.Labels["deployment-annotator"] == "enabled"
			return was != now
		},
		DeleteFunc:  func(event.DeleteEvent) bool { return false },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}
}

func specOf(obj client.Object) interface{} {
	switch w := obj.(type) {
	case *appsv1.Deployment:
		return w.Spec
	case *appsv1.StatefulSet:
		return w.Spec
	case *appsv1.DaemonSet:
		return w.Spec
	}
	return nil
}

func statusOf(obj client.Object) interface{} {
	switch w := obj.(type) {
	case *appsv1.Deployment:
		return w.Status
	case *appsv1.StatefulSet:
		return w.Status
	case *appsv1.DaemonSet:
		return w.Status
	}
	return nil
}
