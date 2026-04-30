package grafana

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var fixedTime = time.Date(2025, 6, 15, 12, 30, 45, 0, time.UTC)

func fixedNow() time.Time { return fixedTime }

func TestCreateAnnotation_UsesInjectedTime(t *testing.T) {
	var got Annotation
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(AnnotationResponse{ID: 42})
	}))
	defer srv.Close()

	c := &Client{URL: srv.URL, APIKey: "test", HTTPClient: srv.Client(), Now: fixedNow}
	id, err := c.CreateAnnotation(context.Background(), "deploy-start:app", []string{"deploy"}, "data")
	if err != nil {
		t.Fatal(err)
	}
	if id != 42 {
		t.Fatalf("got id %d, want 42", id)
	}
	if got.When != fixedTime.Unix() {
		t.Fatalf("got when %d, want %d (epoch seconds)", got.When, fixedTime.Unix())
	}
}

func TestUpdateAnnotationToRegion_UsesInjectedTime(t *testing.T) {
	var got AnnotationPatch
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{URL: srv.URL, APIKey: "test", HTTPClient: srv.Client(), Now: fixedNow}
	err := c.UpdateAnnotationToRegion(context.Background(), 1, []string{"deploy", "region"})
	if err != nil {
		t.Fatal(err)
	}
	if got.TimeEnd != fixedTime.UnixMilli() {
		t.Fatalf("got timeEnd %d, want %d (epoch millis)", got.TimeEnd, fixedTime.UnixMilli())
	}
	if !got.IsRegion {
		t.Fatal("expected isRegion=true")
	}
}

func TestCreateAnnotation_DefaultsToTimeNow(t *testing.T) {
	var got Annotation
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(AnnotationResponse{ID: 1})
	}))
	defer srv.Close()

	before := time.Now().Unix()
	c := &Client{URL: srv.URL, APIKey: "test", HTTPClient: srv.Client()}
	_, err := c.CreateAnnotation(context.Background(), "w", []string{}, "d")
	if err != nil {
		t.Fatal(err)
	}
	after := time.Now().Unix()
	if got.When < before || got.When > after {
		t.Fatalf("when %d not in [%d, %d]", got.When, before, after)
	}
}
