package controller

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// --- fake AnnotationClient ---

type annotationCall struct {
	method string
	what   string
	tags   []string
	data   string
	id     int64
}

type fakeAnnotationClient struct {
	calls  []annotationCall
	nextID int64
}

func (f *fakeAnnotationClient) CreateAnnotation(
	_ context.Context, what string, tags []string, data string,
) (int64, error) {
	f.nextID++
	f.calls = append(f.calls, annotationCall{method: "create", what: what, tags: tags, data: data, id: f.nextID})
	return f.nextID, nil
}

func (f *fakeAnnotationClient) UpdateAnnotationToRegion(_ context.Context, id int64, tags []string) error {
	f.calls = append(f.calls, annotationCall{method: "region", id: id, tags: tags})
	return nil
}

func (f *fakeAnnotationClient) createCalls() []annotationCall {
	var out []annotationCall
	for _, c := range f.calls {
		if c.method == "create" {
			out = append(out, c)
		}
	}
	return out
}

func (f *fakeAnnotationClient) regionCalls() []annotationCall {
	var out []annotationCall
	for _, c := range f.calls {
		if c.method == "region" {
			out = append(out, c)
		}
	}
	return out
}

// --- test helpers ---

func testScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = appsv1.AddToScheme(s)
	return s
}

func trackedNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"deployment-annotator": "enabled"},
		},
	}
}

func untrackedNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

func deployment(name, namespace, image string, generation int64) *appsv1.Deployment {
	replicas := int32(1)
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: namespace,
			Generation: generation,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "main", Image: image}}},
			},
		},
	}
	return d
}

func readyDeployment(name, namespace, image string, generation int64) *appsv1.Deployment {
	d := deployment(name, namespace, image, generation)
	d.Status = appsv1.DeploymentStatus{
		UpdatedReplicas:    1,
		AvailableReplicas:  1,
		ObservedGeneration: generation,
	}
	return d
}

func newReconciler(objs []client.Object, gc *fakeAnnotationClient) (*WorkloadReconciler, client.Client) {
	scheme := testScheme()
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		WithStatusSubresource(&appsv1.Deployment{}).
		Build()
	lc := &AnnotationLifecycle{Client: c, GClient: gc}
	r := &WorkloadReconciler{
		Client:    c,
		Scheme:    scheme,
		Adapter:   DeploymentAdapter{},
		Lifecycle: lc,
	}
	return r, c
}

func reconcileReq(name, namespace string) ctrl.Request {
	return ctrl.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: namespace}}
}

func getDeployment(t *testing.T, c client.Client, name, namespace string) *appsv1.Deployment {
	t.Helper()
	d := &appsv1.Deployment{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, d); err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	return d
}

// --- tests ---

func TestReconcile_NewWorkload_InitializesTracking(t *testing.T) {
	gc := &fakeAnnotationClient{}
	d := deployment("app", "ns", "nginx:1.21", 1)
	r, c := newReconciler([]client.Object{trackedNamespace("ns"), d}, gc)

	_, err := r.Reconcile(context.Background(), reconcileReq("app", "ns"))
	if err != nil {
		t.Fatal(err)
	}

	got := getDeployment(t, c, "app", "ns")
	if v := got.Annotations[VersionAnnotation]; v == "" {
		t.Fatal("expected tracked-version annotation to be set")
	}
	if len(gc.calls) != 0 {
		t.Fatalf("expected no Grafana calls for initialization, got %d", len(gc.calls))
	}
}

func TestReconcile_VersionChange_CreatesStartAnnotation(t *testing.T) {
	gc := &fakeAnnotationClient{}
	d := deployment("app", "ns", "nginx:1.21", 1)
	d.Annotations = map[string]string{VersionAnnotation: "gen-0-img-old"}
	r, c := newReconciler([]client.Object{trackedNamespace("ns"), d}, gc)

	_, err := r.Reconcile(context.Background(), reconcileReq("app", "ns"))
	if err != nil {
		t.Fatal(err)
	}

	creates := gc.createCalls()
	if len(creates) != 1 {
		t.Fatalf("expected 1 create call, got %d", len(creates))
	}
	if creates[0].what != "deploy-start:app" {
		t.Fatalf("expected deploy-start:app, got %s", creates[0].what)
	}

	got := getDeployment(t, c, "app", "ns")
	if got.Annotations[StartAnnotation] == "" {
		t.Fatal("expected start-annotation-id to be set")
	}
	if got.Annotations[EndAnnotation] != "" {
		t.Fatal("expected end-annotation-id to be cleared")
	}
}

func TestReconcile_ReadyWithStartID_CompletesDeployment(t *testing.T) {
	gc := &fakeAnnotationClient{}
	d := readyDeployment("app", "ns", "nginx:1.21", 1)
	version := fmt.Sprintf("gen-%d-img-%s", d.Generation, "1.21")
	d.Annotations = map[string]string{
		VersionAnnotation: version,
		StartAnnotation:   "100",
	}
	r, c := newReconciler([]client.Object{trackedNamespace("ns"), d}, gc)

	_, err := r.Reconcile(context.Background(), reconcileReq("app", "ns"))
	if err != nil {
		t.Fatal(err)
	}

	creates := gc.createCalls()
	if len(creates) != 1 {
		t.Fatalf("expected 1 create call (end), got %d", len(creates))
	}
	if creates[0].what != "deploy-end:app" {
		t.Fatalf("expected deploy-end:app, got %s", creates[0].what)
	}

	regions := gc.regionCalls()
	if len(regions) != 1 {
		t.Fatalf("expected 1 region call, got %d", len(regions))
	}
	if regions[0].id != 100 {
		t.Fatalf("expected region on start ID 100, got %d", regions[0].id)
	}

	got := getDeployment(t, c, "app", "ns")
	if got.Annotations[EndAnnotation] == "" {
		t.Fatal("expected end-annotation-id to be set")
	}
}

func TestReconcile_AlreadyCompleted_IsIdempotent(t *testing.T) {
	gc := &fakeAnnotationClient{}
	d := readyDeployment("app", "ns", "nginx:1.21", 1)
	version := fmt.Sprintf("gen-%d-img-%s", d.Generation, "1.21")
	d.Annotations = map[string]string{
		VersionAnnotation: version,
		StartAnnotation:   "100",
		EndAnnotation:     "101",
	}
	r, _ := newReconciler([]client.Object{trackedNamespace("ns"), d}, gc)

	_, err := r.Reconcile(context.Background(), reconcileReq("app", "ns"))
	if err != nil {
		t.Fatal(err)
	}
	if len(gc.calls) != 0 {
		t.Fatalf("expected no Grafana calls, got %d", len(gc.calls))
	}
}

func TestReconcile_UntrackedNamespace_Skips(t *testing.T) {
	gc := &fakeAnnotationClient{}
	d := deployment("app", "ns", "nginx:1.21", 1)
	r, _ := newReconciler([]client.Object{untrackedNamespace("ns"), d}, gc)

	_, err := r.Reconcile(context.Background(), reconcileReq("app", "ns"))
	if err != nil {
		t.Fatal(err)
	}
	if len(gc.calls) != 0 {
		t.Fatalf("expected no Grafana calls, got %d", len(gc.calls))
	}
}

func TestReconcile_DeletedWorkload_CreatesDeleteAnnotation(t *testing.T) {
	gc := &fakeAnnotationClient{}
	// No deployment — only the namespace exists
	r, _ := newReconciler([]client.Object{trackedNamespace("ns")}, gc)

	_, err := r.Reconcile(context.Background(), reconcileReq("app", "ns"))
	if err != nil {
		t.Fatal(err)
	}

	creates := gc.createCalls()
	if len(creates) != 1 {
		t.Fatalf("expected 1 create call, got %d", len(creates))
	}
	if creates[0].what != "deploy-delete:app" {
		t.Fatalf("expected deploy-delete:app, got %s", creates[0].what)
	}
}

func TestReconcile_DeletedWorkload_UntrackedNamespace_Skips(t *testing.T) {
	gc := &fakeAnnotationClient{}
	r, _ := newReconciler([]client.Object{untrackedNamespace("ns")}, gc)

	_, err := r.Reconcile(context.Background(), reconcileReq("app", "ns"))
	if err != nil {
		t.Fatal(err)
	}
	if len(gc.calls) != 0 {
		t.Fatalf("expected no Grafana calls, got %d", len(gc.calls))
	}
}

func TestReconcile_VersionChange_ClearsEndAnnotation(t *testing.T) {
	gc := &fakeAnnotationClient{}
	d := deployment("app", "ns", "nginx:1.22", 2)
	d.Annotations = map[string]string{
		VersionAnnotation: "gen-1-img-1.21",
		StartAnnotation:   "100",
		EndAnnotation:     "101",
	}
	r, c := newReconciler([]client.Object{trackedNamespace("ns"), d}, gc)

	_, err := r.Reconcile(context.Background(), reconcileReq("app", "ns"))
	if err != nil {
		t.Fatal(err)
	}

	got := getDeployment(t, c, "app", "ns")
	if got.Annotations[EndAnnotation] != "" {
		t.Fatalf("expected end-annotation-id to be cleared, got %q", got.Annotations[EndAnnotation])
	}
	newStartID := got.Annotations[StartAnnotation]
	if newStartID == "" || newStartID == "100" {
		t.Fatalf("expected new start annotation ID, got %q", newStartID)
	}
}

func TestReconcile_NotReady_DoesNotComplete(t *testing.T) {
	gc := &fakeAnnotationClient{}
	d := deployment("app", "ns", "nginx:1.21", 1)
	version := fmt.Sprintf("gen-%d-img-%s", d.Generation, "1.21")
	d.Annotations = map[string]string{
		VersionAnnotation: version,
		StartAnnotation:   "100",
	}
	// Status defaults to zero — not ready
	r, _ := newReconciler([]client.Object{trackedNamespace("ns"), d}, gc)

	_, err := r.Reconcile(context.Background(), reconcileReq("app", "ns"))
	if err != nil {
		t.Fatal(err)
	}
	if len(gc.calls) != 0 {
		t.Fatalf("expected no Grafana calls while not ready, got %d", len(gc.calls))
	}
}

func TestReconcile_FullLifecycle(t *testing.T) {
	gc := &fakeAnnotationClient{}
	d := deployment("app", "ns", "nginx:1.21", 1)
	r, c := newReconciler([]client.Object{trackedNamespace("ns"), d}, gc)
	req := reconcileReq("app", "ns")

	// Step 1: initialize tracking
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	got := getDeployment(t, c, "app", "ns")
	if got.Annotations[VersionAnnotation] == "" {
		t.Fatal("step 1: expected version to be set")
	}
	if len(gc.calls) != 0 {
		t.Fatal("step 1: expected no Grafana calls")
	}

	// Step 2: simulate version change by updating the stored version to something old
	got.Annotations[VersionAnnotation] = "old-version"
	if err := c.Update(context.Background(), got); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	got = getDeployment(t, c, "app", "ns")
	if got.Annotations[StartAnnotation] == "" {
		t.Fatal("step 2: expected start annotation ID")
	}
	startID := got.Annotations[StartAnnotation]

	// Step 3: simulate readiness
	got.Status = appsv1.DeploymentStatus{
		UpdatedReplicas: 1, AvailableReplicas: 1, ObservedGeneration: got.Generation,
	}
	if err := c.Status().Update(context.Background(), got); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	got = getDeployment(t, c, "app", "ns")
	if got.Annotations[EndAnnotation] == "" {
		t.Fatal("step 3: expected end annotation ID")
	}

	// Verify: 2 creates (start + end), 1 region
	creates := gc.createCalls()
	if len(creates) != 2 {
		t.Fatalf("expected 2 creates, got %d", len(creates))
	}
	regions := gc.regionCalls()
	if len(regions) != 1 {
		t.Fatalf("expected 1 region, got %d", len(regions))
	}
	sid, _ := strconv.ParseInt(startID, 10, 64)
	if regions[0].id != sid {
		t.Fatalf("region should reference start ID %d, got %d", sid, regions[0].id)
	}
}
