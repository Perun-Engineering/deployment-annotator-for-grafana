# Deployment Annotator for Grafana

## Badges

[![CI](https://github.com/perun-engineering/deployment-annotator-for-grafana/workflows/CI/badge.svg)](https://github.com/perun-engineering/deployment-annotator-for-grafana/actions/workflows/ci.yml)
[![Security](https://github.com/perun-engineering/deployment-annotator-for-grafana/workflows/Security/badge.svg)](https://github.com/perun-engineering/deployment-annotator-for-grafana/actions/workflows/security.yml)
[![Release](https://github.com/perun-engineering/deployment-annotator-for-grafana/workflows/Release/badge.svg)](https://github.com/perun-engineering/deployment-annotator-for-grafana/actions/workflows/release.yml)

[![Go Report Card](https://goreportcard.com/badge/github.com/perun-engineering/deployment-annotator-for-grafana)](https://goreportcard.com/report/github.com/perun-engineering/deployment-annotator-for-grafana)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker Image](https://img.shields.io/badge/docker-ghcr.io-blue?logo=docker)](https://github.com/perun-engineering/deployment-annotator-for-grafana/pkgs/container/deployment-annotator-for-grafana)

A production-ready Kubernetes Controller that automatically creates per-component Grafana annotations for deployment events, providing granular timing visibility that traditional CI/CD pipeline annotations cannot offer.

## Why This Tool?

### Traditional CI/CD Annotation Approach

In traditional CI/CD pipelines, deployment annotations are typically created manually during the deployment step:

```yaml
# Traditional approach - single annotation for entire pipeline
- name: Create deployment annotation
  run: |
    curl -X POST $GRAFANA_URL/api/annotations \
      -d '{"what":"deploy-start:my-app","when":'$(date +%s)'}'

- name: Deploy with Helm
  run: helm upgrade my-app ./chart --wait

- name: Create completion annotation
  run: |
    curl -X POST $GRAFANA_URL/api/annotations \
      -d '{"what":"deploy-end:my-app","when":'$(date +%s)'}'
```

### Problems with Traditional Approach

#### 1. **Coarse-Grained Timing**
- **Single timeline** for entire Helm chart deployment
- **Cannot distinguish** between individual component deployment times
- **Hides deployment bottlenecks** within multi-component applications

#### 2. **Multi-Component Helm Charts**
Consider a typical microservices Helm chart with multiple deployments:

```yaml
# helm-chart/templates/
├── frontend-deployment.yaml      # Takes 30 seconds to be ready
├── backend-deployment.yaml       # Takes 2 minutes to be ready
├── database-deployment.yaml      # Takes 5 minutes to be ready
├── cache-deployment.yaml         # Takes 10 seconds to be ready
└── worker-deployment.yaml        # Takes 1 minute to be ready
```

**Traditional annotation timeline:**
```
CI/CD Start ────────────────────────────────────── CI/CD End
     │                                                  │
     └── Single annotation covering entire 5+ minutes ──┘
```

**Problem**: You can't see that the database was the bottleneck while cache deployed quickly.

#### 3. **Independent Component Lifecycles**
- Components have **different resource requirements**
- Components have **different startup times**
- Components may **fail independently**
- Pipeline annotations **mask individual component behavior**

### This Controller's Approach

#### **Per-Component Granular Tracking**

```
Database:   ├────────────────────────────────────────┤ (5 min)
Backend:    ├─────────────────────┤                     (2 min)
Frontend:   ├──────────┤                                 (30 sec)
Worker:     ├────────────────┤                           (1 min)
Cache:      ├─────┤                                      (10 sec)
```

Each deployment gets its own annotation timeline showing:
- **Exact start time** when deployment spec changes
- **Individual rollout duration** for each component
- **Precise completion time** when all replicas are ready
- **Component-specific issues** that don't affect others

#### **Better Observability**

**Traditional View:**
```
MyApp Deployment: Started 10:00 → Finished 10:05 (5 minutes)
```

**Controller View:**
```
Database:  Started 10:00 → Finished 10:05 (5 minutes)
Backend:   Started 10:00 → Finished 10:02 (2 minutes)
Frontend:  Started 10:00 → Finished 10:00 (30 seconds)
Worker:    Started 10:00 → Finished 10:01 (1 minute)
Cache:     Started 10:00 → Finished 10:00 (10 seconds)
```

**Insights gained:**
- Database is the bottleneck (5x slower than next component)
- Frontend and cache deploy very quickly
- Backend and worker have reasonable deployment times
- You can optimize database deployment strategy

#### **Real-World Benefits**

1. **Performance Debugging**: Identify which components slow down deployments
2. **Resource Planning**: Understand individual component resource needs
3. **Rollback Decisions**: See which specific component failed during deployment
4. **SLA Tracking**: Monitor component-level deployment SLAs
5. **Team Accountability**: Different teams can track their component's deployment performance

### **Example: Microservices Platform**

A typical e-commerce platform with this controller shows:

```grafana
Cart Service:     ├──┤ (15s) - Lightweight, fast startup
Product Service:  ├─────┤ (45s) - Medium complexity
Payment Service:  ├────────────┤ (2m) - Heavy validation startup
Search Service:   ├──────────────────┤ (3m) - Elasticsearch indexing
Analytics:        ├─┤ (8s) - Stateless, very fast
```

**Without this controller**: Single 3-minute annotation hiding that analytics and cart deploy quickly while search is the bottleneck.

**With this controller**: Clear visibility that search service needs optimization while cart and analytics are performing well.

## Overview

This controller watches `apps/v1` **Deployments** in namespaces labeled with `deployment-annotator=enabled`. When a deployment specification actually changes (not just scaling), it:

1. Uses Kubernetes `generation` and image tag to create a version identifier
2. Compares current version with previously stored version to detect actual changes
3. Creates Grafana annotations only for real deployment changes, not scaling events
4. Stores annotation IDs and version in deployment annotations for persistence
5. Monitors the deployment until it becomes available
6. Updates the Grafana annotation to mark the deployment completion

## Features

- **Per-component granular tracking**: Individual deployment timelines instead of single CI/CD pipeline annotation
- **Multi-component Helm chart support**: See exact timing for each deployment within complex charts
- **Smart change detection**: Uses Kubernetes generation + image tag to detect actual deployment changes vs scaling events
- **Event-driven completion**: Watches ReplicaSet events for instant deployment completion detection (no polling)
- **KEDA/HPA compatible**: Ignores replica count changes, only tracks real application updates
- **Namespace filtering**: Only processes deployments in namespaces with `deployment-annotator=enabled` label
- **Stateless operation**: Survives pod restarts by storing annotation IDs and versions in deployment annotations
- **Production-ready**: Includes proper RBAC, security contexts, and health checks
- **Helm deployment**: Complete Helm chart with configurable values

## Prerequisites

- Kubernetes cluster (1.19+)
- Grafana instance with API access
- Grafana API key with annotation permissions
- Helm 3.x

## Quick Start

### 1. Install with Helm

```bash
# Install the controller
helm install deployment-annotator-controller ./helm/deployment-annotator-controller \
  --set grafana.url=https://your-grafana-instance.com \
  --set-string grafana.apiKey=your-grafana-api-key
```

### 2. Create Tracked Namespaces

```bash
# Apply test namespaces
kubectl apply -f examples/test-namespace.yaml
```

### 3. Test with Sample Deployment

```bash
# Deploy a test application
kubectl apply -f examples/sample-deployment.yaml

# Monitor the deployment
kubectl rollout status deployment/cart-service -n test-grafana-tracking

# Check controller logs
kubectl logs -l app.kubernetes.io/name=grafana-annotation-controller -f
```

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `GRAFANA_URL` | Grafana instance URL | Yes |
| `GRAFANA_API_KEY` | Grafana API key with annotation permissions | Yes |

### Helm Values

Key configuration options in `values.yaml`:

```yaml
# Grafana configuration
grafana:
  url: "https://your-grafana-instance.com"
  apiKey: "your-api-key"

# Controller image
image:
  repository: ghcr.io/perun-engineering/deployment-annotator-for-grafana
  tag: latest

# Resource limits
resources:
  limits:
    cpu: 500m
    memory: 128Mi
  requests:
    cpu: 100m
    memory: 64Mi
```

## Grafana Configuration

### Setting Up Annotation Queries

To display deployment annotations in your Grafana dashboards, you need to configure annotation queries. This allows you to overlay deployment events on your metrics for better correlation analysis.

#### 1. Built-in Annotations Query

Navigate to your dashboard and add annotation queries:

**Dashboard Settings** → **Annotations** → **Add Annotation Query**

**Basic Configuration:**
```json
{
  "name": "Deployments",
  "datasource": "--Grafana--",
  "enable": true,
  "hide": false,
  "iconColor": "blue",
  "query": {
    "datasource": {
      "type": "grafana",
      "uid": "-- Grafana --"
    },
    "filter": {
      "tags": ["deploy"]
    },
    "limit": 100
  }
}
```

#### 2. Namespace-Specific Annotations

Filter annotations by specific namespaces:

```json
{
  "name": "Production Deployments",
  "datasource": "--Grafana--",
  "enable": true,
  "hide": false,
  "iconColor": "red",
  "query": {
    "datasource": {
      "type": "grafana",
      "uid": "-- Grafana --"
    },
    "filter": {
      "tags": ["deploy", "production"]
    },
    "limit": 100
  }
}
```

#### 3. Application-Specific Annotations

Track specific applications:

```json
{
  "name": "Cart Service Deployments",
  "datasource": "--Grafana--",
  "enable": true,
  "hide": false,
  "iconColor": "green",
  "query": {
    "datasource": {
      "type": "grafana",
      "uid": "-- Grafana --"
    },
    "filter": {
      "tags": ["deploy", "cart-service"]
    },
    "limit": 50
  }
}
```

## How It Works

### 1. Controller Registration

The controller uses controller-runtime to watch Deployment resources in namespaces with the `deployment-annotator=enabled` label.

### 2. Smart Change Detection

The controller uses Kubernetes native fields to detect actual deployment changes:

1. **Generation Tracking**: Uses deployment `generation` which increments only on spec changes
2. **Image Tag Tracking**: Combines generation with container image tag for version identification
3. **Version Comparison**: Compares current version (gen-X-img-Y) with stored version in annotations
4. **Scaling Ignored**: Replica count changes don't increment generation, so no annotations created

### 3. Deployment Processing

When a deployment specification actually changes:

```mermaid
sequenceDiagram
    participant K8s as Kubernetes API
    participant C as Controller
    participant G as Grafana

    K8s->>C: Deployment event (Create/Update)
    C->>C: Create version from generation + image tag
    C->>C: Compare with stored version
    alt Version differs (real change)
        C->>G: POST /api/annotations/graphite (start)
        G->>C: Return annotation ID
        C->>K8s: Update deployment with annotation ID + version
    end
    Note over C: Event-driven completion detection
    K8s->>C: ReplicaSet ready event
    C->>C: Check deployment readiness
    C->>G: POST /api/annotations/graphite (end)
        G->>C: Return end annotation ID
        C->>G: PATCH /api/annotations/{start-id} (create time region)
    else Same version (scaling/status)
        C->>C: Skip annotation creation
    end
```

### 4. Grafana Annotations

**Start Annotation:**
```json
{
  "what": "deploy-start:cart-service",
  "tags": ["deploy", "production", "cart-service", "1.21", "started"],
  "data": "Started deployment nginx:1.21",
  "when": 1640995200
}
```

**End Annotation:**
```json
{
  "what": "deploy-end:cart-service",
  "tags": ["deploy", "production", "cart-service", "1.21", "completed"],
  "data": "Completed deployment nginx:1.21",
  "when": 1640995800
}
```

**Time Region Update (Start Annotation):**
```json
{
  "timeEnd": 1640995800000,
  "isRegion": true,
  "tags": ["deploy", "production", "cart-service", "1.21", "region"]
}
```

**Deletion Annotation:**
```json
{
  "what": "deploy-delete:cart-service",
  "tags": ["deploy", "production", "cart-service", "", "deleted"],
  "data": "Deleted deployment cart-service",
  "when": 1640996000
}
```

## Security Considerations

- **RBAC**: Minimal permissions (get/update deployments only)
- **Network policies**: Consider restricting controller network access
- **API key rotation**: Regularly rotate Grafana API keys
- **TLS**: All communication uses TLS encryption
- **Non-root**: Container runs as non-root user (UID 1001)

## Monitoring

The controller exposes metrics and logs for monitoring:

- **Health endpoints**:
  - `/healthz` for liveness probes (port 8080)
  - `/readyz` for readiness probes (port 8080)
- **Metrics endpoint**: `/metrics` for Prometheus scraping (port 8081)
- **Structured logging**: JSON logs with appropriate log levels
- **Controller-runtime metrics**: Built-in metrics for reconciliation performance

## Contributing

We use [Conventional Commits](https://www.conventionalcommits.org/) and [Semantic Versioning](https://semver.org/) for this project.

## Licence

This project is licensed under the MIT Licence - see the LICENCE file for details.
