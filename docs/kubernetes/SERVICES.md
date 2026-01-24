# Kubernetes Services

> A comprehensive guide to understanding Kubernetes Services

## Table of Contents

- [What is a Service?](#what-is-a-service)
- [The Problem Services Solve](#the-problem-services-solve)
- [How Services Work](#how-services-work)
- [Service Types](#service-types)
- [Creating a Service](#creating-a-service)
- [YAML Breakdown](#yaml-breakdown)
- [DNS and Service Discovery](#dns-and-service-discovery)
- [Common Commands](#common-commands)
- [Best Practices](#best-practices)

---

## What is a Service?

A **Service** is a stable network endpoint that provides access to a group of Pods. It acts as an internal load balancer and gives Pods a consistent way to communicate, regardless of their changing IP addresses.

```mermaid
graph LR
    subgraph "The Problem"
        C1[Client] -->|10.1.0.29?| P1[Pod<br/>IP: 10.1.0.29]
        C1 -.->|Pod restarts...| P1
        P1 -.->|New IP!| P2[Pod<br/>IP: 10.1.0.45]
        C1 -->|Connection Failed!| P2
    end
```

```mermaid
graph LR
    subgraph "The Solution"
        C2[Client] -->|service-name:80| S[Service<br/>Stable IP]
        S -->|Load Balance| P3[Pod 1]
        S -->|Load Balance| P4[Pod 2]
        S -->|Load Balance| P5[Pod 3]
    end
    
    style S fill:#e3f2fd
```

**Key Insight**: Clients connect to the Service (stable), not directly to Pods (ephemeral).

---

## The Problem Services Solve

### Pod IPs Are Ephemeral

```mermaid
sequenceDiagram
    participant Client
    participant Pod as Pod (10.1.0.29)
    participant NewPod as New Pod (10.1.0.45)
    
    Client->>Pod: Connect to 10.1.0.29
    Pod-->>Client: Response ✓
    
    Note over Pod: Pod crashes or restarts
    
    Client->>Pod: Connect to 10.1.0.29
    Pod--xClient: Connection Refused! ❌
    
    Note over NewPod: New pod has different IP
    Note over Client: Client doesn't know new IP!
```

### With a Service

```mermaid
sequenceDiagram
    participant Client
    participant Service as Service (10.96.0.100)
    participant Pod1 as Pod 1
    participant Pod2 as Pod 2
    
    Client->>Service: Connect to service:80
    Service->>Pod1: Forward to Pod 1
    Pod1-->>Service: Response
    Service-->>Client: Response ✓
    
    Note over Pod1: Pod 1 crashes
    
    Client->>Service: Connect to service:80
    Service->>Pod2: Forward to Pod 2
    Pod2-->>Service: Response
    Service-->>Client: Response ✓
    
    Note over Client: Client doesn't notice! ✓
```

---

## How Services Work

### Label-Based Selection

Services find their target Pods using **label selectors**, just like Deployments.

```mermaid
graph TB
    subgraph "Service"
        S[Service<br/>selector: app=nginx]
    end
    
    subgraph "Pods"
        P1[Pod 1<br/>labels: app=nginx ✓]
        P2[Pod 2<br/>labels: app=nginx ✓]
        P3[Pod 3<br/>labels: app=redis ✗]
    end
    
    S -->|matches| P1
    S -->|matches| P2
    S -.->|no match| P3
    
    style P1 fill:#c8e6c9
    style P2 fill:#c8e6c9
    style P3 fill:#ffcdd2
```

### Endpoints

The Service maintains a list of **Endpoints** - the actual Pod IPs it routes to.

```mermaid
graph LR
    subgraph "Service: nginx-service"
        S[ClusterIP: 10.96.0.100<br/>Port: 80]
    end
    
    subgraph "Endpoints (Auto-Updated)"
        E[10.1.0.29:80<br/>10.1.0.30:80<br/>10.1.0.31:80]
    end
    
    subgraph "Pods"
        P1[Pod 1: 10.1.0.29]
        P2[Pod 2: 10.1.0.30]
        P3[Pod 3: 10.1.0.31]
    end
    
    S --> E
    E --> P1
    E --> P2
    E --> P3
```

When Pods are added/removed, Kubernetes automatically updates the Endpoints list.

---

## Service Types

### Overview

```mermaid
graph TB
    subgraph "Service Types"
        CI[ClusterIP<br/>Internal only]
        NP[NodePort<br/>External via node port]
        LB[LoadBalancer<br/>External via cloud LB]
    end
    
    subgraph "Accessibility"
        INT[Inside Cluster]
        EXT[Outside Cluster]
    end
    
    CI --> INT
    NP --> INT
    NP --> EXT
    LB --> INT
    LB --> EXT
    
    style CI fill:#e3f2fd
    style NP fill:#fff3e0
    style LB fill:#e8f5e9
```

### 1. ClusterIP (Default)

Only accessible from **inside** the cluster.

```mermaid
graph LR
    subgraph "Kubernetes Cluster"
        subgraph "Internal Traffic"
            FE[Frontend Pod] -->|http://backend-svc| SVC[ClusterIP Service<br/>10.96.0.100]
            SVC --> BE1[Backend Pod 1]
            SVC --> BE2[Backend Pod 2]
        end
    end
    
    EXT[External Client] -.->|Cannot Access| SVC
    
    style EXT fill:#ffcdd2
```

**Use cases:**
- Internal APIs
- Databases
- Caches (Redis, Memcached)
- Service-to-service communication

### 2. NodePort

Accessible from **outside** via `<NodeIP>:<NodePort>`.

```mermaid
graph LR
    subgraph "External"
        EXT[Browser/Client]
    end
    
    subgraph "Node"
        NP[NodePort<br/>:30080]
    end
    
    subgraph "Cluster"
        SVC[Service]
        P1[Pod 1]
        P2[Pod 2]
    end
    
    EXT -->|localhost:30080| NP
    NP --> SVC
    SVC --> P1
    SVC --> P2
    
    style NP fill:#fff3e0
```

**Port range:** 30000-32767

**Use cases:**
- Development/testing
- Simple external access
- When you don't have a cloud load balancer

### 3. LoadBalancer

Gets an **external IP** from cloud provider (AWS, GCP, Azure).

```mermaid
graph LR
    subgraph "Internet"
        USER[Users]
    end
    
    subgraph "Cloud Provider"
        LB[Load Balancer<br/>External IP: 52.1.2.3]
    end
    
    subgraph "Kubernetes Cluster"
        SVC[Service]
        P1[Pod 1]
        P2[Pod 2]
        P3[Pod 3]
    end
    
    USER -->|http://52.1.2.3| LB
    LB --> SVC
    SVC --> P1
    SVC --> P2
    SVC --> P3
    
    style LB fill:#e8f5e9
```

**Use cases:**
- Production external access
- Public-facing APIs
- Web applications

> **Note:** In Docker Desktop, LoadBalancer behaves like NodePort (no real external IP).

### Comparison Table

| Type | Internal Access | External Access | Use Case |
|------|----------------|-----------------|----------|
| ClusterIP | ✅ Yes | ❌ No | Internal services |
| NodePort | ✅ Yes | ✅ Via NodeIP:Port | Development, simple access |
| LoadBalancer | ✅ Yes | ✅ Via External IP | Production in cloud |

---

## Creating a Service

### ClusterIP Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service

spec:
  type: ClusterIP        # Default, can be omitted
  
  selector:
    app: my-app          # Must match Pod labels
  
  ports:
    - name: http
      port: 80           # Service port
      targetPort: 8080   # Pod/container port
```

### NodePort Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-nodeport-service

spec:
  type: NodePort
  
  selector:
    app: my-app
  
  ports:
    - name: http
      port: 80           # Service port (internal)
      targetPort: 8080   # Pod port
      nodePort: 30080    # External port (30000-32767)
```

### LoadBalancer Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-loadbalancer-service

spec:
  type: LoadBalancer
  
  selector:
    app: my-app
  
  ports:
    - name: http
      port: 80
      targetPort: 8080
```

---

## YAML Breakdown

### Port Mapping Explained

```mermaid
graph LR
    subgraph "External (NodePort only)"
        NP[nodePort: 30080]
    end
    
    subgraph "Service"
        SP[port: 80]
    end
    
    subgraph "Pod"
        TP[targetPort: 8080]
    end
    
    NP -->|"localhost:30080"| SP
    SP -->|"service:80"| TP
    
    style NP fill:#fff3e0
    style SP fill:#e3f2fd
    style TP fill:#e8f5e9
```

| Field | Description | Example |
|-------|-------------|---------|
| `port` | Port the Service listens on | Clients call `service:80` |
| `targetPort` | Port on the Pod/container | App runs on `8080` |
| `nodePort` | Port on the Node (NodePort only) | External calls `localhost:30080` |

### The Selector Connection

```mermaid
graph LR
    subgraph "Service YAML"
        SS[selector:<br/>  app: my-app]
    end
    
    subgraph "Deployment's Pod Template"
        PL[labels:<br/>  app: my-app]
    end
    
    SS <-->|MUST MATCH| PL
```

---

## DNS and Service Discovery

### Automatic DNS

Kubernetes automatically creates DNS entries for Services.

```mermaid
graph TB
    subgraph "DNS Resolution"
        A[my-service] -->|resolves to| B[10.96.0.100]
        C[my-service.default] -->|resolves to| B
        D[my-service.default.svc.cluster.local] -->|resolves to| B
    end
```

### DNS Formats

| Format | Description | Example |
|--------|-------------|---------|
| `<service>` | Short name (same namespace) | `redis` |
| `<service>.<namespace>` | Cross-namespace | `redis.cache` |
| `<service>.<namespace>.svc.cluster.local` | Fully qualified | `redis.cache.svc.cluster.local` |

### Using DNS in Your Code

```go
// Instead of hardcoding IPs
// ❌ Don't do this
client.Connect("10.96.0.100:6379")

// ✅ Do this - use service name
client.Connect("redis:6379")

// ✅ Or for cross-namespace
client.Connect("redis.cache:6379")
```

---

## Common Commands

### Create/Update

```bash
# Create or update from YAML
kubectl apply -f service.yaml

# Create imperatively
kubectl expose deployment my-deployment --port=80 --target-port=8080
```

### View

```bash
# List services
kubectl get services
kubectl get svc  # Short form

# Detailed view
kubectl get services -o wide

# Describe (shows endpoints, events)
kubectl describe service my-service

# View endpoints
kubectl get endpoints my-service
```

### Debug

```bash
# Test from inside cluster
kubectl run test --image=busybox --rm -it --restart=Never -- wget -qO- http://my-service:80

# Check if service resolves
kubectl run test --image=busybox --rm -it --restart=Never -- nslookup my-service
```

### Delete

```bash
# Delete service
kubectl delete service my-service

# Delete from YAML
kubectl delete -f service.yaml
```

---

## Best Practices

### 1. Name Your Ports

```yaml
ports:
  - name: http      # ✅ Named port
    port: 80
    targetPort: 8080
  - name: metrics   # ✅ Named port
    port: 9090
    targetPort: 9090
```

### 2. Use Consistent Labels

```yaml
# Deployment
spec:
  selector:
    matchLabels:
      app: my-api
      tier: backend
  template:
    metadata:
      labels:
        app: my-api
        tier: backend

---
# Service
spec:
  selector:
    app: my-api
    tier: backend    # Match all relevant labels
```

### 3. Use ClusterIP for Internal Services

```mermaid
graph LR
    subgraph "Good Architecture"
        ING[Ingress] --> API[API Service<br/>ClusterIP]
        API --> DB[Database Service<br/>ClusterIP]
        API --> CACHE[Cache Service<br/>ClusterIP]
    end
```

Only expose what needs external access.

### 4. Don't Expose Databases Externally

```yaml
# ❌ Don't do this in production
spec:
  type: NodePort  # Database exposed!
  selector:
    app: postgres

# ✅ Do this
spec:
  type: ClusterIP  # Internal only
  selector:
    app: postgres
```

### 5. Use Health Checks with Services

Services only route to **Ready** pods. Combine with readiness probes:

```yaml
# In your Deployment
containers:
- name: my-app
  readinessProbe:
    httpGet:
      path: /health
      port: 8080
    initialDelaySeconds: 5
    periodSeconds: 3
```

---

## Service vs Deployment Relationship

```mermaid
graph TB
    subgraph "What You Create"
        D[Deployment YAML]
        S[Service YAML]
    end
    
    subgraph "What Kubernetes Creates"
        DEP[Deployment]
        RS[ReplicaSet]
        P1[Pod 1<br/>labels: app=nginx]
        P2[Pod 2<br/>labels: app=nginx]
        SVC[Service<br/>selector: app=nginx]
    end
    
    D --> DEP
    DEP --> RS
    RS --> P1
    RS --> P2
    
    S --> SVC
    SVC -.->|selects| P1
    SVC -.->|selects| P2
    
    style SVC fill:#e3f2fd
    style P1 fill:#e8f5e9
    style P2 fill:#e8f5e9
```

**Important:** Deployment and Service are **independent** resources. They're connected only through **labels**.

---

## Quick Reference

### Traffic Flow

```mermaid
graph LR
    subgraph "External"
        EXT[Client]
    end
    
    subgraph "NodePort: 30080"
        NP[Node Port]
    end
    
    subgraph "Service: 80"
        SVC[ClusterIP]
    end
    
    subgraph "Pod: 8080"
        POD[Container]
    end
    
    EXT -->|1. localhost:30080| NP
    NP -->|2. Forward| SVC
    SVC -->|3. Load Balance| POD
```

### Essential Commands Cheat Sheet

| Task | Command |
|------|---------|
| Create | `kubectl apply -f service.yaml` |
| List | `kubectl get services` |
| Details | `kubectl describe service <name>` |
| Endpoints | `kubectl get endpoints <name>` |
| Test | `kubectl run test --image=busybox --rm -it -- wget -qO- http://<service>` |
| Delete | `kubectl delete service <name>` |

### Service Types Decision Tree

```mermaid
graph TD
    A[Need external access?] -->|No| B[ClusterIP]
    A -->|Yes| C[Running in cloud?]
    C -->|Yes| D[LoadBalancer]
    C -->|No| E[NodePort]
    
    style B fill:#e3f2fd
    style D fill:#e8f5e9
    style E fill:#fff3e0
```

---

## Related Resources

- [Deployments](./DEPLOYMENTS.md) - Managing Pod replicas
- [ConfigMaps & Secrets](./CONFIGMAPS_SECRETS.md) - Configuration management
- [Ingress](./INGRESS.md) - HTTP routing and TLS
- [Kubernetes Mental Model](./KUBERNETES_MENTAL_MODEL.md) - Overall architecture

