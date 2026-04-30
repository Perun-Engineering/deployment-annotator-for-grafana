package controller

import (
	"bytes"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// specChangedPredicate triggers on spec changes. When the adapter watches
// status it also triggers on status changes (used by StatefulSet/DaemonSet
// which detect completion via their own status, not via a secondary watch).
func specChangedPredicate(adapter WorkloadAdapter) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(event.CreateEvent) bool { return true },
		DeleteFunc:  func(event.DeleteEvent) bool { return true },
		GenericFunc: func(event.GenericEvent) bool { return true },
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldSpec, _ := json.Marshal(adapter.Spec(e.ObjectOld))
			newSpec, _ := json.Marshal(adapter.Spec(e.ObjectNew))
			if !bytes.Equal(oldSpec, newSpec) {
				return true
			}
			if !adapter.WatchesStatus() {
				return false
			}
			oldStatus, _ := json.Marshal(adapter.Status(e.ObjectOld))
			newStatus, _ := json.Marshal(adapter.Status(e.ObjectNew))
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
