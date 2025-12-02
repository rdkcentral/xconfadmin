# XConf System Architecture

## System Architecture Diagram

```mermaid
graph TB
    subgraph "Client Layer"
        RDK[RDK Devices]
        UI[Web UI]
        API[External APIs]
    end
    
    subgraph "Service Layer"
        subgraph "XConf Admin - Port: 9001"
            AdminAPI[Admin API Handler]
            AdminCache[Built-in Cache<br/>Redis/In-Memory]
            AdminLogic[Business Logic]
        end
        
        subgraph "XConf WebConfig - Port: 9000"
            WebAPI[WebConfig API]
            WebCache[Built-in Cache<br/>Redis/In-Memory]
            RulesEngine[Rules Engine]
        end
        
        subgraph "XConf UI - Port: 8081"
            UIServer[UI Server & Proxy]
        end
    end
    
    subgraph "Data Layer"
        Cassandra[(Cassandra Database)]
    end
    
    subgraph "External Services"
        SAT[SAT Service<br/>Authentication]
        IDP[IDP Service<br/>Identity Provider]
        Metrics[Prometheus<br/>Metrics]
        Tracing[OTEL Collector<br/>Tracing]
    end
    
    RDK -->|Configuration Requests| WebAPI
    UI -->|User Interface| UIServer
    API -->|Admin Operations| AdminAPI
    
    UIServer -->|Proxy Requests| AdminAPI
    
    WebAPI --> WebCache
    WebCache -->|Cache Miss| RulesEngine
    RulesEngine --> Cassandra
    
    AdminAPI --> AdminCache
    AdminCache -->|Cache Miss| AdminLogic
    AdminLogic --> Cassandra
    
    AdminLogic -->|Config Updates| WebCache
    
    AdminAPI --> SAT
    AdminAPI --> IDP
    WebAPI --> SAT
    
    AdminAPI --> Metrics
    WebAPI --> Metrics
    UIServer --> Metrics
    
    AdminAPI --> Tracing
    WebAPI --> Tracing
    
    style AdminCache fill:#e1f5fe
    style WebCache fill:#e1f5fe
    style Cassandra fill:#f3e5f5
    style SAT fill:#fff3e0
    style IDP fill:#fff3e0
```

## Deployment Architecture Diagram

```mermaid
graph TB
    subgraph "Load Balancer Layer"
        LB[Load Balancer<br/>NGINX/HAProxy]
    end
    
    subgraph "Service Instances"
        subgraph "XConf Admin Cluster"
            XA1[Admin Instance 1<br/>Built-in Cache]
            XA2[Admin Instance 2<br/>Built-in Cache]
        end
        
        subgraph "XConf WebConfig Cluster"
            WC1[WebConfig Instance 1<br/>Built-in Cache]
            WC2[WebConfig Instance 2<br/>Built-in Cache]
        end
        
        subgraph "XConf UI Cluster"
            UI1[UI Instance 1]
            UI2[UI Instance 2]
        end
    end
    
    subgraph "Data Layer"
        C1[(Cassandra<br/>Node 1)]
        C2[(Cassandra<br/>Node 2)]
        C3[(Cassandra<br/>Node 3)]
    end
    
    subgraph "External Services"
        SAT[SAT Service]
        IDP[Identity Provider]
        Prometheus[Prometheus]
        OTEL[OTEL Collector]
    end
    
    LB --> XA1
    LB --> XA2
    LB --> WC1
    LB --> WC2
    LB --> UI1
    LB --> UI2
    
    XA1 --> C1
    XA2 --> C2
    WC1 --> C1
    WC2 --> C3
    
    XA1 -.->|Cache Sync| XA2
    WC1 -.->|Cache Sync| WC2
    
    XA1 --> SAT
    XA2 --> SAT
    WC1 --> SAT
    WC2 --> SAT
    
    style XA1 fill:#e3f2fd
    style XA2 fill:#e3f2fd
    style WC1 fill:#e8f5e8
    style WC2 fill:#e8f5e8
```

## How to Generate PNG

To convert this Mermaid diagram to a PNG file, you can use several methods:

### Method 1: Mermaid CLI
```bash
# Install Mermaid CLI
npm install -g @mermaid-js/mermaid-cli

# Generate PNG from this markdown file
mmdc -i architecture.md -o architecture.png -t neutral -b white
```

### Method 2: Online Mermaid Live Editor
1. Visit: https://mermaid.live/
2. Copy the mermaid code from above
3. Paste into the editor
4. Download as PNG

### Method 3: VS Code Extension
1. Install "Mermaid Markdown Syntax Highlighting" extension
2. Open this file in VS Code
3. Right-click on the mermaid diagram
4. Select "Export Diagram" → PNG

### Method 4: GitHub Integration
GitHub automatically renders Mermaid diagrams in markdown files, and you can screenshot or use browser tools to save as PNG.

## Architecture Features

### Built-in Cache Integration
- **XConf Admin**: Contains its own built-in cache (Redis/In-Memory) integrated within the service
- **XConf WebConfig**: Has its own built-in cache layer for high-performance device requests
- **No Separate Cache Layer**: Eliminated the standalone cache infrastructure

### Enhanced Service Structure
- Each service shows internal components (API Handler, Cache, Business Logic)
- Cache operations are handled internally within each service
- Direct cache-to-database flow on cache misses

### Simplified Data Flow
- **WebConfig**: API → Built-in Cache → Rules Engine → Cassandra
- **Admin**: API → Built-in Cache → Business Logic → Cassandra
- **Cross-Service**: Admin can invalidate WebConfig cache on configuration updates

### Deployment Benefits
- **Reduced Infrastructure**: No need for separate Redis cluster management
- **Simplified Operations**: Cache management is part of service lifecycle
- **Better Performance**: In-process caching reduces network overhead
- **Cache Synchronization**: Services can sync cache state when needed