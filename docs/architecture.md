# XConf System Architecture

## System Architecture Diagram

```mermaid
graph TB
    %% Clients
    RDK[RDK Devices] 
    Operators[Operators/Admins]
    
    %% XConf Services
    UI[XConf UI<br/>Port: 8081<br/>Web Interface]
    Admin[XConf Admin<br/>Port: 9001<br/>API + Cache]
    WebConfig[XConf WebConfig<br/>Port: 9000<br/>Data Service + Cache]
    
    %% Database
    DB[(Cassandra<br/>Database)]
    
    %% Flow
    RDK --> WebConfig
    Operators --> UI
    UI --> Admin
    Admin --> DB
    WebConfig --> DB
    
    %% Styling
    style UI fill:#e3f2fd
    style Admin fill:#fff3e0
    style WebConfig fill:#e8f5e8
    style DB fill:#f3e5f5
```

## Deployment Architecture Diagram

```mermaid
graph TB
    %% Load Balancer
    LB[Load Balancer]
    
    %% Service Instances
    UI1[XConf UI]
    UI2[XConf UI]
    
    Admin1[XConf Admin]
    Admin2[XConf Admin] 
    
    Web1[XConf WebConfig]
    Web2[XConf WebConfig]
    
    %% Database Cluster
    DB1[(Cassandra)]
    DB2[(Cassandra)]
    DB3[(Cassandra)]
    
    %% Connections
    LB --> UI1
    LB --> UI2
    LB --> Admin1
    LB --> Admin2
    LB --> Web1
    LB --> Web2
    
    Admin1 --> DB1
    Admin2 --> DB2
    Web1 --> DB1
    Web2 --> DB3
    
    %% Styling
    style UI1 fill:#e3f2fd
    style UI2 fill:#e3f2fd
    style Admin1 fill:#fff3e0
    style Admin2 fill:#fff3e0
    style Web1 fill:#e8f5e8
    style Web2 fill:#e8f5e8
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