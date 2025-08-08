package main

import (
	"context"
	"fmt"
	"time"

	"github.com/perun-engineering/deployment-annotator-for-grafana/internal/grafana"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Helpers to build Grafana annotation payloads using the shared client
func createAnnotation(
	ctx context.Context,
	gc *grafana.Client,
	kind, name, namespace, imageTag, imageRef, eventType string,
) (int64, error) {
	sanitizedName := sanitizeForLog(name)
	sanitizedNamespace := sanitizeForLog(namespace)
	sanitizedTag := sanitizeForLog(imageTag)
	sanitizedRef := sanitizeForLog(imageRef)
	action := map[string]string{"started": "start", "completed": "end", "deleted": "delete"}[eventType]
	what := fmt.Sprintf("deploy-%s:%s", action, sanitizedName)
	if action == "" {
		what = fmt.Sprintf("deploy:%s", sanitizedName)
	}
	title := cases.Title(language.English).String(eventType)
	data := fmt.Sprintf("%s deployment %s", title, sanitizedRef)
	tags := []string{"deploy", sanitizedNamespace, sanitizedName, sanitizedTag, eventType, kind}
	// Add a short context timeout per request
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	return gc.CreateAnnotation(ctx, what, tags, data)
}

func updateAnnotationToRegion(
	ctx context.Context,
	gc *grafana.Client,
	id int64,
	kind, name, namespace, imageTag string,
) error {
	sanitizedName := sanitizeForLog(name)
	sanitizedNamespace := sanitizeForLog(namespace)
	sanitizedTag := sanitizeForLog(imageTag)
	tags := []string{"deploy", sanitizedNamespace, sanitizedName, sanitizedTag, "region", kind}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	return gc.UpdateAnnotationToRegion(ctx, id, tags)
}
