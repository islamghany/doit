# Kubernetes Mental Model Guide

> **Goal:** Build a complete mental model of Kubernetes before writing any YAML
>
> **Approach:** Understand WHY before HOW, use analogies, build incrementally

---

## Table of Contents

1. [The Big Picture: What Problem Does Kubernetes Solve?](#1-the-big-picture-what-problem-does-kubernetes-solve)
2. [The Restaurant Analogy](#2-the-restaurant-analogy)
3. [Kubernetes Architecture](#3-kubernetes-architecture)
4. [Core Concepts (The Building Blocks)](#4-core-concepts-the-building-blocks)
5. [How Objects Relate to Each Other](#5-how-objects-relate-to-each-other)
6. [The Declarative Model](#6-the-declarative-model)
7. [Networking in Kubernetes](#7-networking-in-kubernetes)
8. [Storage in Kubernetes](#8-storage-in-kubernetes)
9. [Configuration Management](#9-configuration-management)
10. [From Docker Compose to Kubernetes](#10-from-docker-compose-to-kubernetes)

---

## 1. The Big Picture: What Problem Does Kubernetes Solve?

### The Problem: Managing Containers at Scale

```
WITHOUT KUBERNETES (Manual Container Management):
═══════════════════════════════════════════════════

You have 3 servers and need to run 10 copies of your API...

Server 1          Server 2          Server 3
┌─────────┐      ┌─────────┐      ┌─────────┐
│ API ?   │      │ API ?   │      │ API ?   │
│ API ?   │      │ API ?   │      │ API ?   │
│ API ?   │      │ API ?   │      │ API ?   │
│ ???     │      │ ???     │      │ ???     │
└─────────┘      └─────────┘      └─────────┘

Questions you must answer MANUALLY:
• Which server has capacity for each container?
• What if Server 2 crashes? Who restarts containers?
• How do users reach any of these containers?
• How do I update all 10 without downtime?
• How do I scale to 20 when traffic spikes?
• How do containers find each other (API → Database)?
```

```
WITH KUBERNETES (Automated Container Orchestration):
════════════════════════════════════════════════════

You tell Kubernetes: "I want 10 copies of my API running"
Kubernetes handles EVERYTHING else.

┌─────────────────────────────────────────────────────────┐
│                    KUBERNETES CLUSTER                    │
│                                                          │
│  "I want 10 API pods"  ──→  Kubernetes figures out:     │
│                              • Where to place them       │
│                              • How to reach them         │
│                              • How to restart if crash   │
│                              • How to update them        │
│                              • How to scale them         │
│                                                          │
│  Server 1          Server 2          Server 3           │
│  ┌─────────┐      ┌─────────┐      ┌─────────┐         │
│  │ API API │      │ API API │      │ API API │         │
│  │ API API │      │ API     │      │ API API │         │
│  └─────────┘      └─────────┘      └─────────┘         │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### Kubernetes in One Sentence

> **Kubernetes is a system that automates deploying, scaling, and managing containerized applications.**

Or even simpler:

> **You describe WHAT you want, Kubernetes figures out HOW to make it happen.**

---

## 2. The Restaurant Analogy

Think of Kubernetes as a **restaurant operation**:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    KUBERNETES = RESTAURANT OPERATION                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  RESTAURANT CONCEPT          KUBERNETES EQUIVALENT                          │
│  ══════════════════          ════════════════════                           │
│                                                                              │
│  Restaurant Chain      →     Cluster                                        │
│  (the whole business)        (all your infrastructure)                      │
│                                                                              │
│  Head Office           →     Control Plane                                  │
│  (makes decisions)           (API Server, Scheduler, etc.)                  │
│                                                                              │
│  Restaurant Locations  →     Nodes (Worker Machines)                        │
│  (where work happens)        (servers that run containers)                  │
│                                                                              │
│  Kitchen Stations      →     Pods                                           │
│  (where dishes made)         (where containers run)                         │
│                                                                              │
│  Recipe Book           →     Deployment                                     │
│  (how to make dishes)        (how to run your app)                          │
│                                                                              │
│  Menu                  →     Service                                        │
│  (what customers see)        (how to access your app)                       │
│                                                                              │
│  Waiters               →     Ingress                                        │
│  (route customers)           (route external traffic)                       │
│                                                                              │
│  Ingredient Storage    →     ConfigMaps & Secrets                           │
│  (supplies)                  (configuration & sensitive data)               │
│                                                                              │
│  Pantry/Freezer        →     Persistent Volumes                             │
│  (long-term storage)         (data that survives restarts)                  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### How the Restaurant Works

```
Customer Order Flow (= HTTP Request Flow):
══════════════════════════════════════════

1. Customer arrives at restaurant
   └── User sends HTTP request

2. Waiter greets customer, takes to table
   └── Ingress receives request, routes to Service

3. Customer orders from menu
   └── Service load-balances to available Pod

4. Kitchen prepares dish
   └── Pod processes request (your container runs)

5. Dish served to customer
   └── Response returned to user


Behind the Scenes (= Kubernetes Control Loop):
══════════════════════════════════════════════

Head Office constantly monitors:
• "We need 5 chefs on duty" (Deployment: replicas: 5)
• "Chef called in sick? Hire replacement!" (Pod crashes → restart)
• "Lunch rush? Add more chefs!" (HPA: scale up)
• "Slow night? Send some home" (HPA: scale down)
• "New recipe? Train all chefs gradually" (Rolling update)
```

---

## 3. Kubernetes Architecture

### The Two Planes

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         KUBERNETES CLUSTER                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                        CONTROL PLANE                                    │ │
│  │                    (The Brain - Makes Decisions)                        │ │
│  │                                                                         │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │ │
│  │  │ API Server  │  │  Scheduler  │  │ Controller  │  │    etcd     │   │ │
│  │  │             │  │             │  │  Manager    │  │             │   │ │
│  │  │ Front door  │  │ Assigns     │  │ Maintains   │  │ Database    │   │ │
│  │  │ for all     │  │ pods to     │  │ desired     │  │ stores all  │   │ │
│  │  │ requests    │  │ nodes       │  │ state       │  │ cluster     │   │ │
│  │  │             │  │             │  │             │  │ state       │   │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │ │
│  │                                                                         │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                    │                                         │
│                                    │ kubectl, API calls                      │
│                                    ▼                                         │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                         DATA PLANE                                      │ │
│  │                   (The Workers - Do the Work)                           │ │
│  │                                                                         │ │
│  │  Node 1                    Node 2                    Node 3             │ │
│  │  ┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐│ │
│  │  │ ┌──────┐ ┌──────┐│     │ ┌──────┐ ┌──────┐│     │ ┌──────┐ ┌──────┐││ │
│  │  │ │ Pod  │ │ Pod  ││     │ │ Pod  │ │ Pod  ││     │ │ Pod  │ │ Pod  │││ │
│  │  │ └──────┘ └──────┘│     │ └──────┘ └──────┘│     │ └──────┘ └──────┘││ │
│  │  │                  │     │                  │     │                  ││ │
│  │  │ ┌──────────────┐ │     │ ┌──────────────┐ │     │ ┌──────────────┐ ││ │
│  │  │ │   kubelet    │ │     │ │   kubelet    │ │     │ │   kubelet    │ ││ │
│  │  │ │ (node agent) │ │     │ │ (node agent) │ │     │ │ (node agent) │ ││ │
│  │  │ └──────────────┘ │     │ └──────────────┘ │     │ └──────────────┘ ││ │
│  │  │ ┌──────────────┐ │     │ ┌──────────────┐ │     │ ┌──────────────┐ ││ │
│  │  │ │  kube-proxy  │ │     │ │  kube-proxy  │ │     │ │  kube-proxy  │ ││ │
│  │  │ │ (networking) │ │     │ │ (networking) │ │     │ │ (networking) │ ││ │
│  │  │ └──────────────┘ │     │ └──────────────┘ │     │ └──────────────┘ ││ │
│  │  └──────────────────┘     └──────────────────┘     └──────────────────┘│ │
│  │                                                                         │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Control Plane Components Explained

| Component              | Role                                                         | Restaurant Analogy                            |
| ---------------------- | ------------------------------------------------------------ | --------------------------------------------- |
| **API Server**         | Front door for all operations. Everything goes through here. | Reception desk - all requests come here first |
| **Scheduler**          | Decides which node runs new pods                             | Manager who assigns chefs to stations         |
| **Controller Manager** | Ensures desired state matches actual state                   | Quality inspector who checks everything       |
| **etcd**               | Stores all cluster data (the source of truth)                | The filing cabinet with all records           |

### Worker Node Components Explained

| Component             | Role                                          | Restaurant Analogy                   |
| --------------------- | --------------------------------------------- | ------------------------------------ |
| **kubelet**           | Agent on each node, manages pods              | Station manager at each location     |
| **kube-proxy**        | Handles networking for pods                   | Phone system connecting all stations |
| **Container Runtime** | Actually runs containers (Docker, containerd) | The actual kitchen equipment         |

---

## 4. Core Concepts (The Building Blocks)

### 4.1 Pod - The Smallest Deployable Unit

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              POD                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  A Pod is like a "logical host" - containers that MUST run together         │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                           POD                                        │    │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐              │    │
│  │  │ Container 1 │    │ Container 2 │    │ Container 3 │              │    │
│  │  │  (your app) │    │  (sidecar)  │    │  (optional) │              │    │
│  │  └─────────────┘    └─────────────┘    └─────────────┘              │    │
│  │         │                  │                  │                      │    │
│  │         └──────────────────┴──────────────────┘                      │    │
│  │                           │                                          │    │
│  │              Shared: Network (localhost)                             │    │
│  │              Shared: Storage (volumes)                               │    │
│  │              Shared: Lifecycle (start/stop together)                 │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  KEY INSIGHT: Most pods have just ONE container!                            │
│  Multi-container pods are for tightly coupled helpers (sidecars)            │
│                                                                              │
│  Examples:                                                                   │
│  • Single container: Your API (most common)                                 │
│  • Multi-container: API + log shipper sidecar                               │
│  • Multi-container: API + service mesh proxy (Istio)                        │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Pod Characteristics:**

- Pods are **ephemeral** (temporary) - they can be killed and recreated anytime
- Pods get a **unique IP address** within the cluster
- Containers in a pod share `localhost` - they can talk via `localhost:port`
- **Never create pods directly** - use Deployments instead!

---

### 4.2 Deployment - How to Run Your App

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           DEPLOYMENT                                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  A Deployment manages a set of identical Pods                               │
│  It's the "recipe" for running your application                             │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                        DEPLOYMENT                                    │    │
│  │                    "doit-api-deployment"                             │    │
│  │                                                                      │    │
│  │  Spec:                                                               │    │
│  │  ├── replicas: 3        (I want 3 copies)                           │    │
│  │  ├── selector:          (which pods I manage)                       │    │
│  │  └── template:          (how to create pods)                        │    │
│  │       ├── image: doit-api:v1.0                                      │    │
│  │       ├── ports: [8080]                                             │    │
│  │       ├── resources:                                                │    │
│  │       │   ├── requests: cpu: 100m, memory: 128Mi                    │    │
│  │       │   └── limits: cpu: 500m, memory: 512Mi                      │    │
│  │       └── env: [...]                                                │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                              │                                               │
│                              │ Creates & Manages                             │
│                              ▼                                               │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                        REPLICASET                                    │    │
│  │              (automatically created by Deployment)                   │    │
│  │                                                                      │    │
│  │  Ensures exactly 3 pods are running at all times                    │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                              │                                               │
│                              │ Creates & Manages                             │
│                              ▼                                               │
│        ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                    │
│        │    Pod 1    │  │    Pod 2    │  │    Pod 3    │                    │
│        │  doit-api   │  │  doit-api   │  │  doit-api   │                    │
│        │ 10.1.0.15   │  │ 10.1.0.16   │  │ 10.1.0.17   │                    │
│        └─────────────┘  └─────────────┘  └─────────────┘                    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**What Deployment Gives You:**

- **Desired State**: "I want 3 replicas" - K8s maintains this
- **Self-Healing**: Pod dies? Deployment creates a new one
- **Rolling Updates**: Update image without downtime
- **Rollback**: Something wrong? Roll back to previous version
- **Scaling**: Change replicas from 3 to 10 instantly

---

### 4.3 Service - How to Access Your App

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            SERVICE                                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  THE PROBLEM:                                                                │
│  ═══════════                                                                 │
│  Pods are ephemeral - they come and go, IPs change constantly               │
│  How do other services find your pods?                                      │
│                                                                              │
│  ┌─────────────┐                                                            │
│  │ Other Pod   │──→ 10.1.0.15 (Pod 1) ← Pod dies, IP gone!                 │
│  │ wants to    │──→ 10.1.0.16 (Pod 2) ← New pod, new IP!                   │
│  │ reach API   │──→ ???                                                     │
│  └─────────────┘                                                            │
│                                                                              │
│  THE SOLUTION: SERVICE                                                       │
│  ════════════════════                                                        │
│  A Service provides a STABLE endpoint that never changes                    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         SERVICE                                      │    │
│  │                    "doit-api-service"                                │    │
│  │                                                                      │    │
│  │  Stable DNS: doit-api-service.default.svc.cluster.local             │    │
│  │  Stable IP:  10.96.0.100 (ClusterIP - never changes)                │    │
│  │                                                                      │    │
│  │  selector:                                                           │    │
│  │    app: doit-api    ← "Route to pods with this label"               │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                              │                                               │
│                              │ Load balances to                              │
│                              ▼                                               │
│        ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                    │
│        │    Pod 1    │  │    Pod 2    │  │    Pod 3    │                    │
│        │ app:doit-api│  │ app:doit-api│  │ app:doit-api│                    │
│        │ 10.1.0.15   │  │ 10.1.0.16   │  │ 10.1.0.17   │                    │
│        └─────────────┘  └─────────────┘  └─────────────┘                    │
│                                                                              │
│  Now other pods just call: http://doit-api-service:8080                     │
│  Service handles finding healthy pods automatically!                         │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Service Types:**

| Type                    | Use Case                   | Accessibility             |
| ----------------------- | -------------------------- | ------------------------- |
| **ClusterIP** (default) | Internal communication     | Only inside cluster       |
| **NodePort**            | Development/testing        | External via node IP:port |
| **LoadBalancer**        | Production (cloud)         | External via cloud LB     |
| **ExternalName**        | Reference external service | DNS alias                 |

```
Service Types Visualized:
═════════════════════════

ClusterIP (Internal Only):
┌──────────────────────────────────────┐
│           CLUSTER                     │
│  ┌──────────┐     ┌──────────┐       │
│  │ Frontend │────→│ Service  │──→Pods│
│  │   Pod    │     │ ClusterIP│       │
│  └──────────┘     └──────────┘       │
│                                       │
│  ✓ Internal communication            │
│  ✗ Not accessible from outside       │
└──────────────────────────────────────┘

NodePort (Exposes on Node):
┌──────────────────────────────────────┐
│           CLUSTER                     │
│  ┌──────────┐     ┌──────────┐       │
│  │  Node    │     │ Service  │──→Pods│
│  │ :30080   │────→│ NodePort │       │
│  └──────────┘     └──────────┘       │
│       ↑                               │
└───────│──────────────────────────────┘
        │
   External: http://node-ip:30080

LoadBalancer (Cloud):
┌──────────────────────────────────────┐
│           CLUSTER                     │
│  ┌──────────┐     ┌──────────┐       │
│  │  Cloud   │     │ Service  │──→Pods│
│  │   LB     │────→│ LoadBal. │       │
│  └──────────┘     └──────────┘       │
│       ↑                               │
└───────│──────────────────────────────┘
        │
   External: http://api.example.com
```

---

### 4.4 Labels and Selectors - The Glue

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      LABELS AND SELECTORS                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Labels are key-value pairs attached to objects                             │
│  Selectors find objects by their labels                                     │
│                                                                              │
│  Think of it like:                                                          │
│  • Labels = Name tags on people at a conference                             │
│  • Selectors = "Find everyone with 'Engineering' on their tag"              │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Pod 1                     Pod 2                     Pod 3          │    │
│  │  ┌─────────────────┐      ┌─────────────────┐      ┌─────────────┐ │    │
│  │  │ labels:         │      │ labels:         │      │ labels:     │ │    │
│  │  │   app: doit-api │      │   app: doit-api │      │   app: redis│ │    │
│  │  │   env: prod     │      │   env: prod     │      │   env: prod │ │    │
│  │  │   tier: backend │      │   tier: backend │      │   tier: cache│ │    │
│  │  └─────────────────┘      └─────────────────┘      └─────────────┘ │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  Service with selector: { app: doit-api }                                   │
│  ════════════════════════════════════════                                   │
│  Matches: Pod 1 ✓, Pod 2 ✓, Pod 3 ✗                                        │
│                                                                              │
│  Service with selector: { env: prod, tier: backend }                        │
│  ════════════════════════════════════════════════════                       │
│  Matches: Pod 1 ✓, Pod 2 ✓, Pod 3 ✗                                        │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Common Label Patterns:**

```yaml
labels:
  app: doit-api # Application name
  version: v1.2.3 # Version
  environment: production # Environment
  tier: backend # Architecture tier
  team: platform # Owning team
  release: stable # Release track
```

---

### 4.5 Namespace - Virtual Clusters

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          NAMESPACES                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Namespaces divide a cluster into virtual sub-clusters                      │
│  Like folders on your computer - organize and isolate resources             │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                      KUBERNETES CLUSTER                              │    │
│  │                                                                      │    │
│  │  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  │    │
│  │  │ namespace: dev   │  │ namespace: staging│  │ namespace: prod  │  │    │
│  │  │                  │  │                  │  │                  │  │    │
│  │  │ • doit-api       │  │ • doit-api       │  │ • doit-api       │  │    │
│  │  │ • postgres       │  │ • postgres       │  │ • postgres       │  │    │
│  │  │ • redis          │  │ • redis          │  │ • redis          │  │    │
│  │  │                  │  │                  │  │                  │  │    │
│  │  │ Resources: Low   │  │ Resources: Med   │  │ Resources: High  │  │    │
│  │  └──────────────────┘  └──────────────────┘  └──────────────────┘  │    │
│  │                                                                      │    │
│  │  ┌──────────────────┐  ┌──────────────────┐                        │    │
│  │  │ namespace:       │  │ namespace:       │                        │    │
│  │  │ kube-system      │  │ monitoring       │                        │    │
│  │  │                  │  │                  │                        │    │
│  │  │ • coredns        │  │ • prometheus     │                        │    │
│  │  │ • kube-proxy     │  │ • grafana        │                        │    │
│  │  │ • metrics-server │  │ • jaeger         │                        │    │
│  │  └──────────────────┘  └──────────────────┘                        │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  Benefits:                                                                   │
│  • Isolation: dev can't accidentally affect prod                            │
│  • Resource Quotas: limit CPU/memory per namespace                          │
│  • Access Control: different permissions per namespace                      │
│  • Organization: logical grouping of related resources                      │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Default Namespaces:**

- `default` - Where resources go if you don't specify
- `kube-system` - Kubernetes system components
- `kube-public` - Publicly accessible data
- `kube-node-lease` - Node heartbeat data

---

## 5. How Objects Relate to Each Other

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    KUBERNETES OBJECT RELATIONSHIPS                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│                         External Traffic                                     │
│                              │                                               │
│                              ▼                                               │
│                      ┌──────────────┐                                       │
│                      │   Ingress    │  Routes external traffic              │
│                      │  (optional)  │  by hostname/path                     │
│                      └──────┬───────┘                                       │
│                             │                                                │
│                             ▼                                                │
│                      ┌──────────────┐                                       │
│                      │   Service    │  Stable endpoint                      │
│                      │              │  Load balances to pods                │
│                      └──────┬───────┘                                       │
│                             │ selector: app=doit-api                        │
│                             ▼                                                │
│                      ┌──────────────┐                                       │
│                      │  Deployment  │  Manages desired state                │
│                      │              │  Handles updates/rollbacks            │
│                      └──────┬───────┘                                       │
│                             │ creates                                        │
│                             ▼                                                │
│                      ┌──────────────┐                                       │
│                      │  ReplicaSet  │  Maintains pod count                  │
│                      │ (automatic)  │  (usually don't touch)                │
│                      └──────┬───────┘                                       │
│                             │ creates                                        │
│                             ▼                                                │
│        ┌─────────────┬─────────────┬─────────────┐                          │
│        │    Pod      │    Pod      │    Pod      │  Actual containers       │
│        │             │             │             │                          │
│        └──────┬──────┴──────┬──────┴──────┬──────┘                          │
│               │             │             │                                  │
│               ▼             ▼             ▼                                  │
│        ┌─────────────────────────────────────────┐                          │
│        │           ConfigMap / Secret            │  Configuration           │
│        └─────────────────────────────────────────┘                          │
│               │                                                              │
│               ▼                                                              │
│        ┌─────────────────────────────────────────┐                          │
│        │        PersistentVolumeClaim            │  Storage                 │
│        └─────────────────────────────────────────┘                          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 6. The Declarative Model

### Imperative vs Declarative

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    IMPERATIVE VS DECLARATIVE                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  IMPERATIVE (How to do it - step by step):                                  │
│  ══════════════════════════════════════════                                  │
│                                                                              │
│  "Create a pod, then create another pod, then create a service..."          │
│                                                                              │
│  kubectl run nginx --image=nginx                                            │
│  kubectl expose pod nginx --port=80                                         │
│  kubectl scale deployment nginx --replicas=3                                │
│                                                                              │
│  Problems:                                                                   │
│  • Hard to reproduce                                                        │
│  • No version control                                                       │
│  • Can't review changes                                                     │
│  • Easy to make mistakes                                                    │
│                                                                              │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  DECLARATIVE (What you want - desired state):                               │
│  ════════════════════════════════════════════                                │
│                                                                              │
│  "Here's a YAML file describing what I want. Make it happen."               │
│                                                                              │
│  kubectl apply -f deployment.yaml                                           │
│                                                                              │
│  Benefits:                                                                   │
│  ✓ Version controlled (Git)                                                 │
│  ✓ Reproducible                                                             │
│  ✓ Code review for infrastructure                                           │
│  ✓ Self-documenting                                                         │
│  ✓ GitOps ready                                                             │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### The Control Loop

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    KUBERNETES CONTROL LOOP                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Kubernetes constantly runs a loop:                                         │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                                                                      │    │
│  │    ┌──────────────┐         ┌──────────────┐                        │    │
│  │    │   DESIRED    │         │   ACTUAL     │                        │    │
│  │    │    STATE     │         │    STATE     │                        │    │
│  │    │              │         │              │                        │    │
│  │    │ "3 replicas" │ ──?──── │ "2 running"  │                        │    │
│  │    │              │ MATCH?  │              │                        │    │
│  │    └──────────────┘         └──────────────┘                        │    │
│  │           │                        │                                │    │
│  │           │      ┌─────────────────┘                                │    │
│  │           │      │                                                  │    │
│  │           ▼      ▼                                                  │    │
│  │    ┌──────────────────┐                                             │    │
│  │    │    COMPARE       │                                             │    │
│  │    │                  │                                             │    │
│  │    │  Desired: 3      │                                             │    │
│  │    │  Actual:  2      │                                             │    │
│  │    │  Diff:    1      │                                             │    │
│  │    └────────┬─────────┘                                             │    │
│  │             │                                                       │    │
│  │             ▼                                                       │    │
│  │    ┌──────────────────┐                                             │    │
│  │    │     ACTION       │                                             │    │
│  │    │                  │                                             │    │
│  │    │ "Create 1 pod"   │                                             │    │
│  │    │                  │                                             │    │
│  │    └──────────────────┘                                             │    │
│  │             │                                                       │    │
│  │             └──────────────────────────────────────┐                │    │
│  │                                                    │                │    │
│  │    ┌──────────────┐         ┌──────────────┐      │                │    │
│  │    │   DESIRED    │         │   ACTUAL     │      │                │    │
│  │    │    STATE     │         │    STATE     │◀─────┘                │    │
│  │    │              │         │              │                        │    │
│  │    │ "3 replicas" │ ══✓══── │ "3 running"  │  MATCH!               │    │
│  │    │              │         │              │                        │    │
│  │    └──────────────┘         └──────────────┘                        │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  This loop runs CONTINUOUSLY. If a pod crashes, K8s notices and fixes it.  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 7. Networking in Kubernetes

### The Networking Model

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    KUBERNETES NETWORKING                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  RULE 1: Every Pod gets its own IP address                                  │
│  RULE 2: Pods can communicate with any other pod (no NAT)                   │
│  RULE 3: Nodes can communicate with any pod (no NAT)                        │
│  RULE 4: Pod's IP is same inside and outside the pod                        │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         CLUSTER NETWORK                              │    │
│  │                                                                      │    │
│  │  Node 1 (192.168.1.10)              Node 2 (192.168.1.11)           │    │
│  │  ┌────────────────────────┐        ┌────────────────────────┐       │    │
│  │  │                        │        │                        │       │    │
│  │  │  Pod A (10.1.1.5)      │        │  Pod C (10.1.2.7)      │       │    │
│  │  │  ┌──────────────────┐  │        │  ┌──────────────────┐  │       │    │
│  │  │  │ Container        │  │        │  │ Container        │  │       │    │
│  │  │  │ :8080            │  │◀──────▶│  │ :8080            │  │       │    │
│  │  │  └──────────────────┘  │  Pod   │  └──────────────────┘  │       │    │
│  │  │                        │  to    │                        │       │    │
│  │  │  Pod B (10.1.1.6)      │  Pod   │  Pod D (10.1.2.8)      │       │    │
│  │  │  ┌──────────────────┐  │        │  ┌──────────────────┐  │       │    │
│  │  │  │ Container        │  │        │  │ Container        │  │       │    │
│  │  │  │ :5432            │  │        │  │ :6379            │  │       │    │
│  │  │  └──────────────────┘  │        │  └──────────────────┘  │       │    │
│  │  │                        │        │                        │       │    │
│  │  └────────────────────────┘        └────────────────────────┘       │    │
│  │                                                                      │    │
│  │  Pod A can reach Pod D directly: curl http://10.1.2.8:6379          │    │
│  │  (But don't do this! Use Services for stable endpoints)             │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### DNS in Kubernetes

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    KUBERNETES DNS                                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Every Service gets a DNS name automatically!                               │
│                                                                              │
│  Format: <service-name>.<namespace>.svc.cluster.local                       │
│                                                                              │
│  Examples:                                                                   │
│  ─────────                                                                   │
│  • doit-api.default.svc.cluster.local                                       │
│  • postgres.default.svc.cluster.local                                       │
│  • prometheus.monitoring.svc.cluster.local                                  │
│                                                                              │
│  Shortcuts (within same namespace):                                         │
│  ──────────────────────────────────                                         │
│  • Just use service name: "doit-api"                                        │
│  • Or with namespace: "doit-api.default"                                    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                                                                      │    │
│  │  Pod (in default namespace)                                         │    │
│  │  ┌──────────────────────────────────────────────────────────────┐   │    │
│  │  │                                                               │   │    │
│  │  │  # All of these work:                                        │   │    │
│  │  │  curl http://postgres:5432                                   │   │    │
│  │  │  curl http://postgres.default:5432                           │   │    │
│  │  │  curl http://postgres.default.svc:5432                       │   │    │
│  │  │  curl http://postgres.default.svc.cluster.local:5432         │   │    │
│  │  │                                                               │   │    │
│  │  │  # Cross-namespace (must include namespace):                 │   │    │
│  │  │  curl http://prometheus.monitoring:9090                      │   │    │
│  │  │                                                               │   │    │
│  │  └──────────────────────────────────────────────────────────────┘   │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 8. Storage in Kubernetes

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    KUBERNETES STORAGE                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  THE PROBLEM:                                                                │
│  Pods are ephemeral - when they die, their data dies with them              │
│                                                                              │
│  THE SOLUTION: Persistent Volumes                                           │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                                                                      │    │
│  │  PersistentVolume (PV)         PersistentVolumeClaim (PVC)          │    │
│  │  ════════════════════          ══════════════════════════           │    │
│  │                                                                      │    │
│  │  "Here's some storage"    →    "I need some storage"                │    │
│  │  (Admin creates)               (Developer requests)                  │    │
│  │                                                                      │    │
│  │  ┌──────────────────┐          ┌──────────────────┐                 │    │
│  │  │ PersistentVolume │◀─────────│ PersistentVolume │                 │    │
│  │  │                  │  binds   │     Claim        │                 │    │
│  │  │ capacity: 10Gi   │          │                  │                 │    │
│  │  │ accessModes:     │          │ request: 5Gi     │                 │    │
│  │  │   - ReadWriteOnce│          │ accessModes:     │                 │    │
│  │  │ storageClass:    │          │   - ReadWriteOnce│                 │    │
│  │  │   standard       │          │                  │                 │    │
│  │  └──────────────────┘          └────────┬─────────┘                 │    │
│  │                                         │                            │    │
│  │                                         │ mounted by                 │    │
│  │                                         ▼                            │    │
│  │                                ┌──────────────────┐                  │    │
│  │                                │       Pod        │                  │    │
│  │                                │                  │                  │    │
│  │                                │ volumes:         │                  │    │
│  │                                │   - name: data   │                  │    │
│  │                                │     pvc: my-pvc  │                  │    │
│  │                                │                  │                  │    │
│  │                                │ containers:      │                  │    │
│  │                                │   volumeMounts:  │                  │    │
│  │                                │     - /data      │                  │    │
│  │                                └──────────────────┘                  │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  Access Modes:                                                               │
│  • ReadWriteOnce (RWO) - Single node can mount read/write                   │
│  • ReadOnlyMany (ROX) - Multiple nodes can mount read-only                  │
│  • ReadWriteMany (RWX) - Multiple nodes can mount read/write                │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 9. Configuration Management

### ConfigMaps and Secrets

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    CONFIGMAPS AND SECRETS                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ConfigMap: Non-sensitive configuration                                      │
│  Secret: Sensitive data (passwords, tokens, keys)                           │
│                                                                              │
│  ┌──────────────────────────────┐  ┌──────────────────────────────┐         │
│  │         ConfigMap            │  │          Secret              │         │
│  │                              │  │                              │         │
│  │  data:                       │  │  data:                       │         │
│  │    LOG_LEVEL: "info"         │  │    DB_PASSWORD: "base64..."  │         │
│  │    DB_HOST: "postgres"       │  │    JWT_SECRET: "base64..."   │         │
│  │    DB_PORT: "5432"           │  │    API_KEY: "base64..."      │         │
│  │    ENVIRONMENT: "production" │  │                              │         │
│  │                              │  │  type: Opaque                │         │
│  └──────────────────────────────┘  └──────────────────────────────┘         │
│                                                                              │
│  Usage in Pod:                                                               │
│  ══════════════                                                              │
│                                                                              │
│  Option 1: Environment Variables                                            │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │  env:                                                                 │   │
│  │    - name: LOG_LEVEL                                                  │   │
│  │      valueFrom:                                                       │   │
│  │        configMapKeyRef:                                               │   │
│  │          name: doit-config                                            │   │
│  │          key: LOG_LEVEL                                               │   │
│  │    - name: DB_PASSWORD                                                │   │
│  │      valueFrom:                                                       │   │
│  │        secretKeyRef:                                                  │   │
│  │          name: doit-secrets                                           │   │
│  │          key: DB_PASSWORD                                             │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  Option 2: All at once (envFrom)                                            │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │  envFrom:                                                             │   │
│  │    - configMapRef:                                                    │   │
│  │        name: doit-config                                              │   │
│  │    - secretRef:                                                       │   │
│  │        name: doit-secrets                                             │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  Option 3: Mount as files                                                   │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │  volumes:                                                             │   │
│  │    - name: config-volume                                              │   │
│  │      configMap:                                                       │   │
│  │        name: doit-config                                              │   │
│  │  containers:                                                          │   │
│  │    - volumeMounts:                                                    │   │
│  │        - name: config-volume                                          │   │
│  │          mountPath: /etc/config                                       │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 10. From Docker Compose to Kubernetes

### Mapping Concepts

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    DOCKER COMPOSE → KUBERNETES                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Docker Compose Concept       Kubernetes Equivalent                         │
│  ══════════════════════       ═════════════════════                         │
│                                                                              │
│  services:                →   Deployment + Service                          │
│    api:                       (one for each service)                        │
│                                                                              │
│  image: doit-api:v1       →   spec.containers[].image                       │
│                                                                              │
│  ports:                   →   Service (type: ClusterIP/NodePort/LB)         │
│    - "8080:8080"              + containerPort in Deployment                 │
│                                                                              │
│  environment:             →   ConfigMap + Secret                            │
│    DB_HOST: postgres          + env/envFrom in Deployment                   │
│                                                                              │
│  volumes:                 →   PersistentVolumeClaim                         │
│    - data:/var/lib/data       + volumeMounts in Deployment                  │
│                                                                              │
│  networks:                →   Not needed! All pods can talk                 │
│    - app-network              (use NetworkPolicy to restrict)               │
│                                                                              │
│  depends_on:              →   Not direct equivalent                         │
│    - postgres                 Use init containers or readiness probes       │
│                                                                              │
│  healthcheck:             →   livenessProbe + readinessProbe                │
│    test: curl...              (more granular in K8s)                        │
│                                                                              │
│  restart: always          →   Deployment handles this automatically         │
│                                                                              │
│  deploy:                  →   Deployment spec                               │
│    replicas: 3                spec.replicas: 3                              │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Side-by-Side Example

```yaml
# DOCKER COMPOSE                    # KUBERNETES
# ══════════════                    # ══════════

version: "3.8" # Multiple YAML files or one with ---
services:
  api: # --- Deployment ---
    image: doit-api:v1.0 # apiVersion: apps/v1
    ports: # kind: Deployment
      - "8080:8080" # metadata:
    environment: #   name: doit-api
      DB_HOST: postgres # spec:
      DB_PORT: "5432" #   replicas: 3
      LOG_LEVEL: info #   selector:
    depends_on: #     matchLabels:
      postgres: #       app: doit-api
        condition: healthy #   template:
    healthcheck: #     metadata:
      test: ["CMD", "curl", ...] #       labels:
      interval: 30s #         app: doit-api
    deploy: #     spec:
      replicas:
        3 #       containers:
        #       - name: api
        #         image: doit-api:v1.0
        #         ports:
        #         - containerPort: 8080
        #         envFrom:
        #         - configMapRef:
        #             name: doit-config
        #         livenessProbe:
        #           httpGet:
        #             path: /health
        #             port: 8080
        #         readinessProbe:
        #           httpGet:
        #             path: /ready
        #             port: 8080
        #
        # --- Service ---
        # apiVersion: v1
        # kind: Service
        # metadata:
        #   name: doit-api
        # spec:
        #   selector:
        #     app: doit-api
        #   ports:
        #   - port: 80
        #     targetPort: 8080
        #
        # --- ConfigMap ---
        # apiVersion: v1
        # kind: ConfigMap
        # metadata:
        #   name: doit-config
        # data:
        #   DB_HOST: postgres
        #   DB_PORT: "5432"
        #   LOG_LEVEL: info
```

---

## 11. Docker Compose vs Kubernetes: Deep Comparison

### When to Use What

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    WHEN TO USE DOCKER COMPOSE VS KUBERNETES                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  USE DOCKER COMPOSE WHEN:                USE KUBERNETES WHEN:               │
│  ════════════════════════                ════════════════════               │
│                                                                              │
│  ✓ Local development                    ✓ Production deployment             │
│  ✓ Single machine deployment            ✓ Multi-machine clusters            │
│  ✓ Simple applications                  ✓ Complex microservices             │
│  ✓ Quick prototyping                    ✓ High availability required        │
│  ✓ CI/CD testing                        ✓ Auto-scaling needed               │
│  ✓ Small team projects                  ✓ Large team/enterprise             │
│  ✓ Learning containers                  ✓ Cloud-native applications         │
│                                                                              │
│  Complexity: Low                        Complexity: High                    │
│  Learning Curve: Days                   Learning Curve: Weeks/Months        │
│  Overhead: Minimal                      Overhead: Significant               │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Feature Comparison Table

| Feature                  | Docker Compose                        | Kubernetes                         |
| ------------------------ | ------------------------------------- | ---------------------------------- |
| **Primary Use**          | Local development, simple deployments | Production, enterprise, cloud      |
| **Scale**                | Single host                           | Multiple hosts (cluster)           |
| **Self-Healing**         | ❌ Manual restart                     | ✅ Automatic restart & replacement |
| **Auto-Scaling**         | ❌ Not built-in                       | ✅ HPA, VPA, Cluster Autoscaler    |
| **Rolling Updates**      | ⚠️ Basic                              | ✅ Advanced with rollback          |
| **Load Balancing**       | ❌ External needed                    | ✅ Built-in (Service)              |
| **Service Discovery**    | ✅ DNS by service name                | ✅ DNS + more options              |
| **Secret Management**    | ⚠️ Basic (env files)                  | ✅ Secrets, external vaults        |
| **Health Checks**        | ✅ Basic                              | ✅ Liveness + Readiness probes     |
| **Resource Limits**      | ✅ Basic                              | ✅ Requests + Limits + Quotas      |
| **Network Policies**     | ❌ Not built-in                       | ✅ Fine-grained control            |
| **Storage**              | ✅ Volumes                            | ✅ PV, PVC, StorageClasses         |
| **Configuration**        | ✅ .env files                         | ✅ ConfigMaps, Secrets             |
| **Declarative**          | ✅ YAML                               | ✅ YAML                            |
| **Learning Curve**       | 🟢 Easy                               | 🔴 Steep                           |
| **Operational Overhead** | 🟢 Low                                | 🔴 High                            |

### Architecture Differences

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    ARCHITECTURE COMPARISON                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  DOCKER COMPOSE                                                              │
│  ══════════════                                                              │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                      SINGLE HOST                                     │    │
│  │                                                                      │    │
│  │    docker-compose.yml                                               │    │
│  │           │                                                          │    │
│  │           ▼                                                          │    │
│  │    ┌─────────────┐                                                  │    │
│  │    │   Docker    │                                                  │    │
│  │    │   Engine    │                                                  │    │
│  │    └──────┬──────┘                                                  │    │
│  │           │                                                          │    │
│  │    ┌──────┴──────┬──────────────┬──────────────┐                   │    │
│  │    │             │              │              │                    │    │
│  │    ▼             ▼              ▼              ▼                    │    │
│  │  ┌─────┐     ┌─────┐      ┌─────────┐    ┌─────────┐              │    │
│  │  │ API │     │ API │      │Postgres │    │  Redis  │              │    │
│  │  └─────┘     └─────┘      └─────────┘    └─────────┘              │    │
│  │                                                                      │    │
│  │  • All containers on ONE machine                                    │    │
│  │  • Docker manages lifecycle                                         │    │
│  │  • Simple networking (bridge)                                       │    │
│  │  • If host dies, everything dies                                    │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  KUBERNETES                                                                  │
│  ══════════                                                                  │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                      CLUSTER (Multiple Hosts)                        │    │
│  │                                                                      │    │
│  │    kubectl / YAML                                                   │    │
│  │           │                                                          │    │
│  │           ▼                                                          │    │
│  │    ┌─────────────────────────────────────────┐                      │    │
│  │    │           CONTROL PLANE                  │                      │    │
│  │    │  (API Server, Scheduler, Controllers)   │                      │    │
│  │    └─────────────────┬───────────────────────┘                      │    │
│  │                      │                                               │    │
│  │         ┌────────────┼────────────┐                                 │    │
│  │         │            │            │                                 │    │
│  │         ▼            ▼            ▼                                 │    │
│  │    ┌─────────┐  ┌─────────┐  ┌─────────┐                           │    │
│  │    │ Node 1  │  │ Node 2  │  │ Node 3  │                           │    │
│  │    │┌───┐┌───┐│  │┌───┐┌───┐│  │┌───┐┌───┐│                           │    │
│  │    ││API││API││  ││API││DB ││  ││API││Red││                           │    │
│  │    │└───┘└───┘│  │└───┘└───┘│  │└───┘└───┘│                           │    │
│  │    └─────────┘  └─────────┘  └─────────┘                           │    │
│  │                                                                      │    │
│  │  • Containers distributed across MULTIPLE machines                  │    │
│  │  • Control plane manages everything                                 │    │
│  │  • Complex networking (overlay, CNI)                                │    │
│  │  • If node dies, pods reschedule to other nodes                    │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Self-Healing Comparison

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    SELF-HEALING BEHAVIOR                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  SCENARIO: Container crashes                                                │
│                                                                              │
│  DOCKER COMPOSE:                                                            │
│  ────────────────                                                           │
│                                                                              │
│  restart: always  ──→  Docker restarts container on SAME host              │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Host                                                                │   │
│  │  ┌─────────┐                    ┌─────────┐                         │   │
│  │  │   API   │  ──crashes──→      │   API   │  ← restarted            │   │
│  │  │  (dead) │                    │ (alive) │    on same host         │   │
│  │  └─────────┘                    └─────────┘                         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ⚠️  If HOST dies, everything dies. No recovery.                           │
│                                                                              │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  KUBERNETES:                                                                │
│  ───────────                                                                │
│                                                                              │
│  Deployment with replicas: 3                                                │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Cluster                                                             │   │
│  │                                                                      │   │
│  │  Node 1           Node 2           Node 3                           │   │
│  │  ┌─────────┐     ┌─────────┐     ┌─────────┐                       │   │
│  │  │   API   │     │   API   │     │   API   │                       │   │
│  │  │ (alive) │     │  (dead) │     │ (alive) │                       │   │
│  │  └─────────┘     └────┬────┘     └─────────┘                       │   │
│  │                       │                                             │   │
│  │                       │ crashes                                     │   │
│  │                       ▼                                             │   │
│  │  Node 1           Node 2           Node 3                           │   │
│  │  ┌─────────┐     ┌─────────┐     ┌─────────┐                       │   │
│  │  │   API   │     │   API   │     │   API   │  ← new pod            │   │
│  │  │ (alive) │     │ (alive) │     │ (alive) │    scheduled          │   │
│  │  └─────────┘     └─────────┘     └─────────┘    (maybe on          │   │
│  │                                                  different node)   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ✅  If NODE dies, pods are rescheduled to healthy nodes automatically     │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Scaling Comparison

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    SCALING COMPARISON                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  DOCKER COMPOSE:                                                            │
│  ────────────────                                                           │
│                                                                              │
│  # Manual scaling (limited to one host)                                     │
│  docker-compose up --scale api=5                                            │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Single Host (limited resources)                                     │   │
│  │                                                                      │   │
│  │  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐                           │   │
│  │  │ API │ │ API │ │ API │ │ API │ │ API │  ← All on same machine   │   │
│  │  └─────┘ └─────┘ └─────┘ └─────┘ └─────┘                           │   │
│  │                                                                      │   │
│  │  ⚠️  Can't scale beyond host capacity                               │   │
│  │  ⚠️  No automatic scaling                                           │   │
│  │  ⚠️  Need external load balancer                                    │   │
│  │                                                                      │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  KUBERNETES:                                                                │
│  ───────────                                                                │
│                                                                              │
│  # Manual scaling                                                           │
│  kubectl scale deployment api --replicas=10                                 │
│                                                                              │
│  # OR Automatic scaling (HPA)                                               │
│  apiVersion: autoscaling/v2                                                 │
│  kind: HorizontalPodAutoscaler                                              │
│  spec:                                                                      │
│    minReplicas: 3                                                           │
│    maxReplicas: 20                                                          │
│    metrics:                                                                 │
│    - type: Resource                                                         │
│      resource:                                                              │
│        name: cpu                                                            │
│        target:                                                              │
│          averageUtilization: 70                                             │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Cluster (multiple nodes, auto-scaling)                              │   │
│  │                                                                      │   │
│  │  Node 1       Node 2       Node 3       Node 4 (auto-added)         │   │
│  │  ┌───┐┌───┐  ┌───┐┌───┐  ┌───┐┌───┐  ┌───┐┌───┐                   │   │
│  │  │API││API│  │API││API│  │API││API│  │API││API│                   │   │
│  │  └───┘└───┘  └───┘└───┘  └───┘└───┘  └───┘└───┘                   │   │
│  │                                                                      │   │
│  │  ✅  Scale across multiple nodes                                    │   │
│  │  ✅  Automatic pod scaling (HPA)                                    │   │
│  │  ✅  Automatic node scaling (Cluster Autoscaler)                    │   │
│  │  ✅  Built-in load balancing (Service)                              │   │
│  │                                                                      │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Update/Deployment Strategy Comparison

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    DEPLOYMENT STRATEGIES                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  DOCKER COMPOSE:                                                            │
│  ────────────────                                                           │
│                                                                              │
│  # Update process                                                           │
│  docker-compose pull                                                        │
│  docker-compose up -d                                                       │
│                                                                              │
│  Timeline:                                                                  │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │  t=0     │ Old containers running                                    │  │
│  │  t=1     │ Stop old containers ← DOWNTIME STARTS                    │  │
│  │  t=2     │ Start new containers                                      │  │
│  │  t=3     │ New containers ready ← DOWNTIME ENDS                     │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ⚠️  Potential downtime during updates                                     │
│  ⚠️  No automatic rollback                                                 │
│  ⚠️  Manual intervention needed if update fails                            │
│                                                                              │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  KUBERNETES (Rolling Update - Default):                                     │
│  ──────────────────────────────────────                                     │
│                                                                              │
│  # Update process                                                           │
│  kubectl set image deployment/api api=doit-api:v2                          │
│                                                                              │
│  Timeline (zero downtime):                                                  │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │  t=0     │ [v1] [v1] [v1]     │ 3 old pods running                  │  │
│  │  t=1     │ [v1] [v1] [v1] [v2]│ New pod starting                    │  │
│  │  t=2     │ [v1] [v1] [v2] [v2]│ New pod ready, old terminating      │  │
│  │  t=3     │ [v1] [v2] [v2] [v2]│ Continuing rollout                  │  │
│  │  t=4     │ [v2] [v2] [v2]     │ Rollout complete                    │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ✅  Zero downtime                                                         │
│  ✅  Automatic rollback on failure                                         │
│  ✅  Configurable rollout speed                                            │
│                                                                              │
│  # Rollback if something goes wrong                                        │
│  kubectl rollout undo deployment/api                                       │
│                                                                              │
│  # Check rollout status                                                    │
│  kubectl rollout status deployment/api                                     │
│                                                                              │
│  # View rollout history                                                    │
│  kubectl rollout history deployment/api                                    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Configuration & Secrets Comparison

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    CONFIGURATION MANAGEMENT                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  DOCKER COMPOSE:                                                            │
│  ────────────────                                                           │
│                                                                              │
│  # .env file (plain text!)                                                  │
│  DB_PASSWORD=supersecret                                                    │
│  JWT_SECRET=mysecretkey                                                     │
│                                                                              │
│  # docker-compose.yml                                                       │
│  services:                                                                  │
│    api:                                                                     │
│      env_file:                                                              │
│        - .env                                                               │
│      environment:                                                           │
│        - DB_HOST=postgres                                                   │
│                                                                              │
│  ⚠️  Secrets in plain text files                                           │
│  ⚠️  .env files often committed to git by mistake                          │
│  ⚠️  No secret rotation                                                    │
│  ⚠️  No access control                                                     │
│                                                                              │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  KUBERNETES:                                                                │
│  ───────────                                                                │
│                                                                              │
│  # ConfigMap (non-sensitive)                                                │
│  apiVersion: v1                                                             │
│  kind: ConfigMap                                                            │
│  metadata:                                                                  │
│    name: app-config                                                         │
│  data:                                                                      │
│    DB_HOST: postgres                                                        │
│    LOG_LEVEL: info                                                          │
│                                                                              │
│  # Secret (sensitive - base64 encoded, can be encrypted)                   │
│  apiVersion: v1                                                             │
│  kind: Secret                                                               │
│  metadata:                                                                  │
│    name: app-secrets                                                        │
│  type: Opaque                                                               │
│  data:                                                                      │
│    DB_PASSWORD: c3VwZXJzZWNyZXQ=  # base64                                 │
│    JWT_SECRET: bXlzZWNyZXRrZXk=   # base64                                 │
│                                                                              │
│  ✅  Secrets separate from config                                          │
│  ✅  Can encrypt secrets at rest (EncryptionConfig)                        │
│  ✅  RBAC controls who can access secrets                                  │
│  ✅  Can integrate with external vaults (AWS Secrets Manager, HashiCorp)   │
│  ✅  Secrets can be rotated without redeploying                            │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Networking Comparison

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    NETWORKING COMPARISON                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  DOCKER COMPOSE:                                                            │
│  ────────────────                                                           │
│                                                                              │
│  networks:                                                                  │
│    app-network:                                                             │
│      driver: bridge                                                         │
│                                                                              │
│  • Simple bridge network                                                    │
│  • All services can reach each other by name                               │
│  • No network policies (all or nothing)                                    │
│  • Port mapping for external access                                        │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Bridge Network                                                      │   │
│  │                                                                      │   │
│  │  ┌─────┐     ┌─────────┐     ┌───────┐                             │   │
│  │  │ API │◄───►│ Postgres│◄───►│ Redis │                             │   │
│  │  └─────┘     └─────────┘     └───────┘                             │   │
│  │                                                                      │   │
│  │  Everyone can talk to everyone (no restrictions)                    │   │
│  │                                                                      │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  KUBERNETES:                                                                │
│  ───────────                                                                │
│                                                                              │
│  • Flat network (all pods can reach all pods by default)                   │
│  • Services for stable endpoints                                           │
│  • NetworkPolicies for fine-grained control                                │
│  • Ingress for external HTTP(S) routing                                    │
│  • Multiple ingress controllers available                                  │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Cluster Network with NetworkPolicy                                  │   │
│  │                                                                      │   │
│  │  ┌─────────────────────────────────────────────────────────────┐    │   │
│  │  │  NetworkPolicy: Only API can access Postgres                │    │   │
│  │  │                                                              │    │   │
│  │  │  ┌─────┐     ┌─────────┐                                    │    │   │
│  │  │  │ API │────►│ Postgres│                                    │    │   │
│  │  │  └─────┘     └─────────┘                                    │    │   │
│  │  │     │              ▲                                        │    │   │
│  │  │     │              │ BLOCKED                                │    │   │
│  │  │     ▼              │                                        │    │   │
│  │  │  ┌───────┐    ┌────┴────┐                                   │    │   │
│  │  │  │ Redis │    │ Hacker  │                                   │    │   │
│  │  │  └───────┘    │   Pod   │                                   │    │   │
│  │  │               └─────────┘                                   │    │   │
│  │  └─────────────────────────────────────────────────────────────┘    │   │
│  │                                                                      │   │
│  │  ✅  Fine-grained network security                                  │   │
│  │                                                                      │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Health Checks Comparison

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    HEALTH CHECKS COMPARISON                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  DOCKER COMPOSE:                                                            │
│  ────────────────                                                           │
│                                                                              │
│  healthcheck:                                                               │
│    test: ["CMD", "curl", "-f", "http://localhost:8080/health"]             │
│    interval: 30s                                                            │
│    timeout: 10s                                                             │
│    retries: 3                                                               │
│    start_period: 40s                                                        │
│                                                                              │
│  • Single health check                                                      │
│  • Used for: container restart, depends_on conditions                      │
│  • Binary: healthy or unhealthy                                            │
│                                                                              │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  KUBERNETES:                                                                │
│  ───────────                                                                │
│                                                                              │
│  THREE types of probes:                                                     │
│                                                                              │
│  1. LIVENESS PROBE: "Is the container alive?"                              │
│     ─────────────────────────────────────────                              │
│     • If fails: Container is KILLED and restarted                          │
│     • Use for: Detecting deadlocks, hung processes                         │
│                                                                              │
│     livenessProbe:                                                          │
│       httpGet:                                                              │
│         path: /health/live                                                  │
│         port: 8080                                                          │
│       initialDelaySeconds: 10                                               │
│       periodSeconds: 30                                                     │
│                                                                              │
│  2. READINESS PROBE: "Is the container ready for traffic?"                 │
│     ─────────────────────────────────────────────────────                  │
│     • If fails: Pod removed from Service (no traffic)                      │
│     • Container keeps running (not killed)                                 │
│     • Use for: Waiting for dependencies, warming up                        │
│                                                                              │
│     readinessProbe:                                                         │
│       httpGet:                                                              │
│         path: /health/ready                                                 │
│         port: 8080                                                          │
│       initialDelaySeconds: 5                                                │
│       periodSeconds: 10                                                     │
│                                                                              │
│  3. STARTUP PROBE: "Has the container started successfully?"               │
│     ─────────────────────────────────────────────────────                  │
│     • Disables liveness/readiness until startup succeeds                   │
│     • Use for: Slow-starting containers                                    │
│                                                                              │
│     startupProbe:                                                           │
│       httpGet:                                                              │
│         path: /health/startup                                               │
│         port: 8080                                                          │
│       failureThreshold: 30                                                  │
│       periodSeconds: 10                                                     │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                                                                      │   │
│  │  Probe Types:                                                       │   │
│  │  • httpGet: HTTP request (most common)                              │   │
│  │  • tcpSocket: TCP connection                                        │   │
│  │  • exec: Run command in container                                   │   │
│  │  • grpc: gRPC health check                                          │   │
│  │                                                                      │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Summary: Migration Path

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    MIGRATION PATH: COMPOSE → KUBERNETES                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Your Journey:                                                              │
│  ═════════════                                                              │
│                                                                              │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                   │
│  │   Docker    │     │   Docker    │     │ Kubernetes  │                   │
│  │  (single    │────►│  Compose    │────►│  (cluster   │                   │
│  │ container)  │     │ (multi-     │     │ orchestr.)  │                   │
│  │             │     │ container)  │     │             │                   │
│  └─────────────┘     └─────────────┘     └─────────────┘                   │
│                                                                              │
│  You are here: ─────────────────────►                                       │
│                                                                              │
│  Skills that transfer:                                                      │
│  ═════════════════════                                                      │
│  ✅ Container concepts (images, volumes, networks)                         │
│  ✅ YAML configuration                                                     │
│  ✅ Environment variables                                                  │
│  ✅ Health checks                                                          │
│  ✅ Service discovery (DNS)                                                │
│  ✅ Declarative infrastructure                                             │
│                                                                              │
│  New concepts to learn:                                                     │
│  ══════════════════════                                                     │
│  📚 Pods, Deployments, ReplicaSets                                         │
│  📚 Services (ClusterIP, NodePort, LoadBalancer)                           │
│  📚 ConfigMaps and Secrets                                                 │
│  📚 Ingress                                                                │
│  📚 Namespaces                                                             │
│  📚 RBAC (Role-Based Access Control)                                       │
│  📚 Horizontal Pod Autoscaler                                              │
│  📚 Helm (package manager)                                                 │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Summary: Key Mental Models

### 1. Desired State vs Actual State

> You declare what you want. Kubernetes makes it happen and keeps it that way.

### 2. Everything is an Object

> Pods, Services, Deployments - all are API objects with a spec (desired) and status (actual).

### 3. Labels are the Glue

> Services find Pods via labels. Deployments manage Pods via labels. Everything connects through labels.

### 4. Pods are Ephemeral

> Never depend on a specific Pod. Use Services for stable endpoints.

### 5. Layers of Abstraction

> Deployment → ReplicaSet → Pod → Container. Each layer handles different concerns.

### 6. Networking is Flat

> Every Pod can reach every other Pod. Use Services for discovery, NetworkPolicies for security.

---

## Next Steps

Ready to get hands-on? Let's:

1. **Set up a local Kubernetes cluster** (Docker Desktop, minikube, or kind) ✅
2. **Deploy your first Pod** manually ✅
3. **Create a Deployment** for your DoIt API
4. **Expose it with a Service**
5. **Add ConfigMaps and Secrets**
6. **Set up Ingress** for external access

---

## 12. Hands-On: Essential kubectl Commands

### Basic Commands You'll Use Daily

```bash
# ═══════════════════════════════════════════════════════════════════════════
# CLUSTER INFORMATION
# ═══════════════════════════════════════════════════════════════════════════

kubectl cluster-info              # Show cluster endpoint
kubectl get nodes                 # List all nodes
kubectl get nodes -o wide         # More details (IP, OS, runtime)

# ═══════════════════════════════════════════════════════════════════════════
# WORKING WITH PODS
# ═══════════════════════════════════════════════════════════════════════════

kubectl get pods                  # List pods in default namespace
kubectl get pods -A               # List pods in ALL namespaces
kubectl get pods -o wide          # Show more info (IP, Node)
kubectl get pods -w               # Watch for changes (live updates)

kubectl describe pod <name>       # Detailed info about a pod
kubectl logs <pod-name>           # View pod logs
kubectl logs <pod-name> -f        # Follow logs (like tail -f)
kubectl logs <pod-name> -c <container>  # Logs from specific container

kubectl exec <pod-name> -- <cmd>  # Run command in pod
kubectl exec -it <pod-name> -- sh # Interactive shell in pod

kubectl delete pod <name>         # Delete a pod

# ═══════════════════════════════════════════════════════════════════════════
# WORKING WITH DEPLOYMENTS
# ═══════════════════════════════════════════════════════════════════════════

kubectl get deployments           # List deployments
kubectl describe deployment <name> # Deployment details
kubectl scale deployment <name> --replicas=5  # Scale up/down
kubectl rollout status deployment <name>      # Check rollout progress
kubectl rollout history deployment <name>     # View rollout history
kubectl rollout undo deployment <name>        # Rollback to previous

# ═══════════════════════════════════════════════════════════════════════════
# WORKING WITH SERVICES
# ═══════════════════════════════════════════════════════════════════════════

kubectl get services              # List services (or: kubectl get svc)
kubectl describe service <name>   # Service details
kubectl get endpoints             # Show service endpoints (pod IPs)

# ═══════════════════════════════════════════════════════════════════════════
# APPLYING MANIFESTS
# ═══════════════════════════════════════════════════════════════════════════

kubectl apply -f <file.yaml>      # Create/update resources from file
kubectl apply -f <directory>/     # Apply all files in directory
kubectl delete -f <file.yaml>     # Delete resources defined in file

# ═══════════════════════════════════════════════════════════════════════════
# DEBUGGING
# ═══════════════════════════════════════════════════════════════════════════

kubectl get events                # Cluster events (useful for debugging)
kubectl get events --sort-by='.lastTimestamp'  # Sorted by time
kubectl top pods                  # Resource usage (needs metrics-server)
kubectl top nodes                 # Node resource usage
```

### Output Formats

```bash
kubectl get pods -o yaml          # Full YAML output
kubectl get pods -o json          # JSON output
kubectl get pods -o wide          # Extra columns
kubectl get pods -o name          # Just names

# Custom columns
kubectl get pods -o custom-columns=NAME:.metadata.name,STATUS:.status.phase
```

### Namespace Operations

```bash
kubectl get namespaces            # List namespaces
kubectl create namespace dev      # Create namespace
kubectl get pods -n kube-system   # Pods in specific namespace
kubectl config set-context --current --namespace=dev  # Set default namespace
```
