package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkloadAdapter captures the differences between Deployment, StatefulSet,
// and DaemonSet so a single WorkloadReconciler can handle all three.
type WorkloadAdapter interface {
	Kind() string
	NewObject() client.Object
	NewObjectList() client.ObjectList
	ComputeVersion(ctx context.Context, c client.Client, obj client.Object, imageTag string) string
	IsReady(obj client.Object) bool
	WatchesStatus() bool
}

// --- Deployment adapter ---

type DeploymentAdapter struct{}

func (DeploymentAdapter) Kind() string                     { return "deployment" }
func (DeploymentAdapter) NewObject() client.Object         { return &appsv1.Deployment{} }
func (DeploymentAdapter) NewObjectList() client.ObjectList { return &appsv1.DeploymentList{} }
func (DeploymentAdapter) WatchesStatus() bool              { return false }

func (DeploymentAdapter) ComputeVersion(
	ctx context.Context, c client.Client, obj client.Object, imageTag string,
) string {
	d := obj.(*appsv1.Deployment)
	rsList := &appsv1.ReplicaSetList{}
	if err := c.List(ctx, rsList,
		client.InNamespace(d.Namespace),
		client.MatchingLabels(d.Spec.Selector.MatchLabels),
	); err == nil {
		var current *appsv1.ReplicaSet
		for i := range rsList.Items {
			rs := &rsList.Items[i]
			if metav1.IsControlledBy(rs, d) {
				if current == nil || rs.CreationTimestamp.After(current.CreationTimestamp.Time) {
					current = rs
				}
			}
		}
		if current != nil {
			if h, ok := current.Labels["pod-template-hash"]; ok {
				return fmt.Sprintf("hash-%s-img-%s", h, imageTag)
			}
		}
	}
	return fmt.Sprintf("gen-%d-img-%s", d.Generation, imageTag)
}

func (DeploymentAdapter) IsReady(obj client.Object) bool {
	d := obj.(*appsv1.Deployment)
	desired := int32(0)
	if d.Spec.Replicas != nil {
		desired = *d.Spec.Replicas
	}
	return d.Status.UpdatedReplicas == desired &&
		d.Status.AvailableReplicas == desired &&
		d.Status.ObservedGeneration == d.Generation
}

// --- StatefulSet adapter ---

type StatefulSetAdapter struct{}

func (StatefulSetAdapter) Kind() string                     { return "statefulset" }
func (StatefulSetAdapter) NewObject() client.Object         { return &appsv1.StatefulSet{} }
func (StatefulSetAdapter) NewObjectList() client.ObjectList { return &appsv1.StatefulSetList{} }
func (StatefulSetAdapter) WatchesStatus() bool              { return true }

func (StatefulSetAdapter) ComputeVersion(
	_ context.Context, _ client.Client, obj client.Object, imageTag string,
) string {
	return fmt.Sprintf("gen-%d-img-%s", obj.(*appsv1.StatefulSet).Generation, imageTag)
}

func (StatefulSetAdapter) IsReady(obj client.Object) bool {
	s := obj.(*appsv1.StatefulSet)
	desired := int32(0)
	if s.Spec.Replicas != nil {
		desired = *s.Spec.Replicas
	}
	return s.Status.ReadyReplicas == desired &&
		s.Status.ObservedGeneration == s.Generation
}

// --- DaemonSet adapter ---

type DaemonSetAdapter struct{}

func (DaemonSetAdapter) Kind() string                     { return "daemonset" }
func (DaemonSetAdapter) NewObject() client.Object         { return &appsv1.DaemonSet{} }
func (DaemonSetAdapter) NewObjectList() client.ObjectList { return &appsv1.DaemonSetList{} }
func (DaemonSetAdapter) WatchesStatus() bool              { return true }

func (DaemonSetAdapter) ComputeVersion(_ context.Context, _ client.Client, obj client.Object, imageTag string) string {
	return fmt.Sprintf("gen-%d-img-%s", obj.(*appsv1.DaemonSet).Generation, imageTag)
}

func (DaemonSetAdapter) IsReady(obj client.Object) bool {
	d := obj.(*appsv1.DaemonSet)
	return d.Status.NumberAvailable > 0 &&
		d.Status.NumberAvailable == d.Status.DesiredNumberScheduled &&
		d.Status.ObservedGeneration == d.Generation
}
