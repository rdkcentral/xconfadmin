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

## Process Flow Diagrams

### Admin Create/Update Configuration Flow

```mermaid
sequenceDiagram
    participant Admin as Administrator
    participant UI as XConf UI
    participant AdminSvc as XConf Admin
    participant Cache as Built-in Cache
    participant DB as Cassandra
    participant WebConfig as XConf WebConfig
    
    Note over Admin,WebConfig: Configuration Create/Update Process
    
    Admin->>UI: 1. Create/Update Config
    Note right of Admin: Via Web Interface
    
    UI->>AdminSvc: 2. POST/PUT API Request
    Note right of UI: /xconfAdminService/firmwareconfig
    
    AdminSvc->>AdminSvc: 3. Validate Request
    Note right of AdminSvc: Auth, Schema, Rules
    
    AdminSvc->>DB: 4. Save Configuration
    Note right of AdminSvc: Persist to Database
    
    DB-->>AdminSvc: 5. Confirm Save
    
    AdminSvc->>Cache: 6. Update Local Cache
    Note right of AdminSvc: Refresh built-in cache
    
    AdminSvc->>WebConfig: 7. Invalidate Cache
    Note right of AdminSvc: Notify WebConfig to refresh
    
    WebConfig->>WebConfig: 8. Clear Cache Entry
    Note right of WebConfig: Remove cached config
    
    AdminSvc-->>UI: 9. Success Response
    UI-->>Admin: 10. Confirmation
    
    Note over Admin,WebConfig: Configuration is now active for devices
```

### Device Configuration Request Flow

```mermaid
sequenceDiagram
    participant Device as RDK Device
    participant WebConfig as XConf WebConfig
    participant Cache as Built-in Cache
    participant Rules as Rules Engine
    participant DB as Cassandra
    
    Note over Device,DB: Device Configuration Request Process
    
    Device->>WebConfig: 1. Request Configuration
    Note right of Device: GET /xconf/swu/stb?mac=XX&model=Y
    
    WebConfig->>Cache: 2. Check Cache
    Note right of WebConfig: Look for cached result
    
    alt Cache Hit
        Cache-->>WebConfig: 3a. Return Cached Config
        Note right of Cache: Configuration found
    else Cache Miss
        Cache-->>WebConfig: 3b. Cache Miss
        
        WebConfig->>Rules: 4. Evaluate Rules
        Note right of WebConfig: Apply device-specific rules
        
        Rules->>DB: 5. Query Configurations
        Note right of Rules: Get firmware rules, etc.
        
        DB-->>Rules: 6. Return Config Data
        
        Rules->>Rules: 7. Process Rules
        Note right of Rules: Match device to config
        
        Rules-->>WebConfig: 8. Computed Config
        
        WebConfig->>Cache: 9. Store in Cache
        Note right of WebConfig: Cache for future requests
    end
    
    WebConfig->>WebConfig: 10. Format Response
    WebConfig-->>Device: 11. Return Configuration
    
    Note over Device,DB: Device applies received configuration
```

### Admin Rule Creation Flow

```mermaid
sequenceDiagram
    participant Admin as Administrator
    participant UI as XConf UI
    participant AdminSvc as XConf Admin
    participant Cache as Built-in Cache
    participant DB as Cassandra
    
    Note over Admin,DB: Firmware Rule Creation Process
    
    Admin->>UI: 1. Create Firmware Rule
    Note right of Admin: Define targeting conditions
    
    UI->>AdminSvc: 2. POST /firmwarerule
    Note right of UI: Submit rule definition
    
    AdminSvc->>AdminSvc: 3. Validate Rule
    Note right of AdminSvc: Check syntax, references
    
    AdminSvc->>DB: 4. Check Dependencies
    Note right of AdminSvc: Verify firmware config exists
    
    DB-->>AdminSvc: 5. Dependency Status
    
    alt Dependencies Valid
        AdminSvc->>DB: 6a. Save Rule
        DB-->>AdminSvc: 7a. Confirm Save
        
        AdminSvc->>Cache: 8a. Update Cache
        Note right of AdminSvc: Add rule to cache
        
        AdminSvc-->>UI: 9a. Success Response
    else Dependencies Invalid
        AdminSvc-->>UI: 6b. Error Response
        Note right of AdminSvc: Missing firmware config
    end
    
    UI-->>Admin: 10. Result Notification
```

### Device Rule Evaluation Flow

```mermaid
sequenceDiagram
    participant Device as RDK Device
    participant WebConfig as XConf WebConfig
    participant Cache as Built-in Cache
    participant Rules as Rules Engine
    participant DB as Cassandra
    
    Note over Device,DB: Rule-Based Configuration Delivery
    
    Device->>WebConfig: 1. Configuration Request
    Note right of Device: With device parameters
    
    WebConfig->>Cache: 2. Check Rule Cache
    Note right of WebConfig: Look for matching rules
    
    alt Rules Cached
        Cache-->>WebConfig: 3a. Return Cached Rules
    else Rules Not Cached
        Cache-->>WebConfig: 3b. Cache Miss
        
        WebConfig->>DB: 4. Load All Rules
        Note right of WebConfig: Get firmware rules
        
        DB-->>WebConfig: 5. Return Rules
        
        WebConfig->>Cache: 6. Cache Rules
        Note right of WebConfig: Store for reuse
    end
    
    WebConfig->>Rules: 7. Evaluate Rules
    Note right of Rules: Match device to rules by priority
    
    Rules->>Rules: 8. Apply Conditions
    Note right of Rules: Check model, MAC, env, etc.
    
    alt Rule Matches
        Rules->>DB: 9a. Get Firmware Config
        DB-->>Rules: 10a. Return Config
        Rules-->>WebConfig: 11a. Matched Configuration
    else No Rule Match
        Rules-->>WebConfig: 9b. Default Configuration
    end
    
    WebConfig->>Cache: 12. Cache Result
    Note right of WebConfig: Store computed config
    
    WebConfig-->>Device: 13. Return Configuration
```

### Cache Invalidation Flow

```mermaid
sequenceDiagram
    participant Admin as Administrator
    participant AdminSvc as XConf Admin
    participant AdminCache as Admin Cache
    participant WebConfig as XConf WebConfig
    participant WebCache as WebConfig Cache
    participant DB as Cassandra
    
    Note over Admin,DB: Configuration Update with Cache Invalidation
    
    Admin->>AdminSvc: 1. Update Configuration
    
    AdminSvc->>DB: 2. Update Database
    DB-->>AdminSvc: 3. Confirm Update
    
    AdminSvc->>AdminCache: 4. Invalidate Admin Cache
    Note right of AdminSvc: Clear related cache entries
    
    AdminSvc->>WebConfig: 5. Send Cache Invalidation
    Note right of AdminSvc: REST call or message
    
    WebConfig->>WebCache: 6. Clear Cache Entries
    Note right of WebConfig: Remove affected configs
    
    WebConfig-->>AdminSvc: 7. Confirm Invalidation
    AdminSvc-->>Admin: 8. Update Complete
    
    Note over Admin,DB: Next device request will get fresh config
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