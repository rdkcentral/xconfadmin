# XConf Tagging Service API Documentation

**Version:** 1.0  
**Last Updated:** January 2026  
**Base URL:** `/taggingService`  
**Default Port:** 9001

---

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Core Concepts](#core-concepts)
4. [API Endpoints](#api-endpoints)
5. [Data Models](#data-models)
6. [Request/Response Examples](#requestresponse-examples)
7. [Error Handling](#error-handling)
8. [Usage Patterns](#usage-patterns)
9. [Best Practices](#best-practices)

---

## Overview

The XConf Tagging Service API provides comprehensive tag management capabilities for device grouping and configuration targeting. It enables dynamic device categorization through tag-based systems, supporting both individual member management and percentage-based distribution for advanced deployment strategies.

### Key Capabilities

- **Tag Management**: Create, read, update, and delete device tags
- **Member Management**: Add/remove individual or batch members to/from tags
- **Tag Queries**: Retrieve tags by ID, list all tags, find tags containing specific members
- **Percentage-Based Distribution**: Calculate deterministic percentage values for devices to enable percentage-based rollouts
- **Batch Operations**: Efficiently manage large numbers of devices in single API calls
- **Real-time Updates**: Immediate availability of tag changes across all services

### Service Architecture

The Tagging Service operates independently and provides direct tag management without intermediate caching layers, ensuring immediate consistency across the XConf platform.

---

## Authentication

All endpoints require authentication via JWT token or session-based authentication.

### Authentication Header Format

```
Authorization: Bearer <JWT_TOKEN>
```

### Authentication Requirements

- **JWT Tokens**: Must be valid and non-expired
- **Session-based**: Valid session credentials required
- **Role-based Access Control**: Operations require appropriate permissions
- **Service-to-Service**: SAT (Security Access Token) for inter-service communication

---

## Core Concepts

### Tags

A tag is a named grouping mechanism for devices. Tags can contain multiple device members and are identified by a unique string identifier.

**Tag Characteristics:**
- Unique identifier (string)
- List of member identifiers (device addresses, IDs, etc.)
- Last updated timestamp
- No size limitations (though practical limits apply)

### Members

Members are individual devices identified by:
- MAC addresses (e.g., `AA:BB:CC:DD:EE:FF`)
- Device IDs (e.g., `device_12345`)
- Account IDs (e.g., `account_67890`)
- Any custom string identifier

### Percentage Distribution

Percentage values are calculated deterministically using SipHash algorithm, enabling consistent percentage-based device selection for:
- Canary deployments
- A/B testing
- Gradual rollouts
- Statistical device sampling

---

## API Endpoints

### 1. Get All Tags

Retrieve all available tags in the system with optional full object expansion.

**Request:**
```
GET /taggingService/tags
GET /taggingService/tags?full
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `full` | flag | No | Returns complete tag objects with members instead of just tag IDs |

**Response (without `full`):**
```json
[
  "production_devices",
  "beta_testers",
  "development_units",
  "canary_group_1"
]
```

**Response (with `full`):**
```json
[
  {
    "id": "production_devices",
    "members": [
      "AA:BB:CC:DD:EE:01",
      "AA:BB:CC:DD:EE:02",
      "AA:BB:CC:DD:EE:03"
    ],
    "updated": 1699875600000
  },
  {
    "id": "beta_testers",
    "members": [
      "AA:BB:CC:DD:EE:10",
      "AA:BB:CC:DD:EE:11"
    ],
    "updated": 1699875500000
  }
]
```

**HTTP Status Codes:**
- `200 OK`: Successfully retrieved tags
- `500 Internal Server Error`: Server-side processing error

---

### 2. Get Tag by ID

Retrieve a specific tag with complete information including all members and metadata.

**Request:**
```
GET /taggingService/tags/{tag}
```

**Parameters:**

| Parameter | Type | Location | Required | Description |
|-----------|------|----------|----------|-------------|
| `tag` | string | Path | Yes | Tag identifier |

**Response:**
```json
{
  "id": "production_devices",
  "members": [
    "AA:BB:CC:DD:EE:01",
    "AA:BB:CC:DD:EE:02",
    "AA:BB:CC:DD:EE:03",
    "AA:BB:CC:DD:EE:04"
  ],
  "updated": 1699875600000
}
```

**HTTP Status Codes:**
- `200 OK`: Tag found and returned
- `400 Bad Request`: Invalid tag parameter
- `404 Not Found`: Tag does not exist
- `500 Internal Server Error`: Server processing error

---

### 3. Delete Tag

Delete a specific tag from the system completely.

**Request:**
```
DELETE /taggingService/tags/{tag}
```

**Parameters:**

| Parameter | Type | Location | Required | Description |
|-----------|------|----------|----------|-------------|
| `tag` | string | Path | Yes | Tag identifier to delete |

**Response:**
- No response body on successful deletion (HTTP 204)

**HTTP Status Codes:**
- `204 No Content`: Tag successfully deleted
- `400 Bad Request`: Invalid tag parameter
- `404 Not Found`: Tag not found
- `500 Internal Server Error`: Server processing error

---

### 4. Add Members to Tag

Add multiple members to an existing tag in a batch operation.

**Request:**
```
PUT /taggingService/tags/{tag}/members
Content-Type: application/json

[
  "AA:BB:CC:DD:EE:05",
  "AA:BB:CC:DD:EE:06",
  "device_id_12345"
]
```

**Parameters:**

| Parameter | Type | Location | Required | Description |
|-----------|------|----------|----------|-------------|
| `tag` | string | Path | Yes | Tag identifier |
| Body | array | Request Body | Yes | Array of member identifiers (max 1000) |

**Constraints:**
- Maximum 1000 members per request
- Exceeding limit returns 400 Bad Request
- Empty member list returns 400 Bad Request

**Response:**
- Success confirmation with no response body (HTTP 200 OK)

**HTTP Status Codes:**
- `200 OK`: Members successfully added
- `400 Bad Request`: Invalid input, batch size exceeded, or empty member list
- `404 Not Found`: Tag not found
- `500 Internal Server Error`: Server processing error

---

### 5. Remove Single Member from Tag

Remove a single member from a specific tag.

**Request:**
```
DELETE /taggingService/tags/{tag}/members/{member}
```

**Parameters:**

| Parameter | Type | Location | Required | Description |
|-----------|------|----------|----------|-------------|
| `tag` | string | Path | Yes | Tag identifier |
| `member` | string | Path | Yes | Member identifier to remove |

**Response:**
```json
{
  "id": "production_devices",
  "members": [
    "AA:BB:CC:DD:EE:01",
    "AA:BB:CC:DD:EE:03",
    "AA:BB:CC:DD:EE:04"
  ],
  "updated": 1699875700000
}
```

**HTTP Status Codes:**
- `200 OK`: Member successfully removed and tag returned
- `204 No Content`: Member successfully removed
- `400 Bad Request`: Invalid parameters
- `404 Not Found`: Tag or member not found
- `500 Internal Server Error`: Server processing error

---

### 6. Remove Multiple Members from Tag

Remove multiple members from a tag in a single batch operation.

**Request:**
```
DELETE /taggingService/tags/{tag}/members
Content-Type: application/json

[
  "AA:BB:CC:DD:EE:02",
  "AA:BB:CC:DD:EE:05",
  "device_id_old_123"
]
```

**Parameters:**

| Parameter | Type | Location | Required | Description |
|-----------|------|----------|----------|-------------|
| `tag` | string | Path | Yes | Tag identifier |
| Body | array | Request Body | Yes | Array of member identifiers to remove (max 1000) |

**Constraints:**
- Maximum 1000 members per request
- Exceeding limit returns 400 Bad Request

**Response:**
- No response body on successful removal (HTTP 204 No Content)

**HTTP Status Codes:**
- `204 No Content`: Members successfully removed
- `400 Bad Request`: Invalid input or batch size exceeded
- `404 Not Found`: Tag not found
- `500 Internal Server Error`: Server processing error

---

### 7. Get Tag Members

Retrieve all members of a specific tag.

**Request:**
```
GET /taggingService/tags/{tag}/members
```

**Parameters:**

| Parameter | Type | Location | Required | Description |
|-----------|------|----------|----------|-------------|
| `tag` | string | Path | Yes | Tag identifier |

**Response:**
```json
[
  "AA:BB:CC:DD:EE:01",
  "AA:BB:CC:DD:EE:02",
  "AA:BB:CC:DD:EE:03",
  "AA:BB:CC:DD:EE:04",
  "device_id_12345",
  "device_id_67890"
]
```

**HTTP Status Codes:**
- `200 OK`: Members successfully retrieved
- `400 Bad Request`: Invalid tag parameter
- `404 Not Found`: Tag not found
- `500 Internal Server Error`: Server processing error

---

### 8. Get Tags by Member

Retrieve all tags containing a specific member.

**Request:**
```
GET /taggingService/tags/members/{member}
```

**Parameters:**

| Parameter | Type | Location | Required | Description |
|-----------|------|----------|----------|-------------|
| `member` | string | Path | Yes | Member identifier to search for |

**Response:**
```json
[
  {
    "id": "production_devices",
    "members": [
      "AA:BB:CC:DD:EE:01",
      "AA:BB:CC:DD:EE:02",
      "other_device_ids..."
    ],
    "updated": 1699875600000
  },
  {
    "id": "priority_group",
    "members": [
      "AA:BB:CC:DD:EE:01",
      "priority_device_ids..."
    ],
    "updated": 1699875550000
  }
]
```

**HTTP Status Codes:**
- `200 OK`: Tags successfully retrieved
- `400 Bad Request`: Invalid member parameter
- `404 Not Found`: Member not found in any tags
- `500 Internal Server Error`: Server processing error

---

### 9. Calculate Percentage Value for Member

Calculate deterministic percentage value for a member using SipHash algorithm.

**Request:**
```
GET /taggingService/tags/members/{member}/percentages/calculation
```

**Parameters:**

| Parameter | Type | Location | Required | Description |
|-----------|------|----------|----------|-------------|
| `member` | string | Path | Yes | Member identifier for calculation |

**Algorithm Details:**
- **Name**: SipHash-2-4
- **Key Configuration**: Fixed keys for consistent results
- **Output Range**: 0-99 (deterministic percentage value)
- **Use Case**: Percentage-based device rollout strategies
- **Deterministic**: Same member always returns same percentage value

**Response:**
```json
67
```

**HTTP Status Codes:**
- `200 OK`: Percentage calculated successfully
- `400 Bad Request`: Invalid member parameter
- `500 Internal Server Error`: Server processing error

---

## Data Models

### Tag Object

```json
{
  "id": "string",           // Unique tag identifier (required)
  "members": ["string"],    // Array of member identifiers
  "updated": "integer"      // Unix timestamp (milliseconds) of last update
}
```

**Field Details:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier for the tag |
| `members` | array | No | Array of device MAC addresses, device IDs, or identifiers |
| `updated` | integer | No | Unix timestamp in milliseconds of last modification |

### Member Identifier Formats

Supported member identifier formats:

| Format | Example | Description |
|--------|---------|-------------|
| MAC Address | `AA:BB:CC:DD:EE:FF` | 48-bit hardware address in colon-separated hex |
| Device ID | `device_12345` | Alphanumeric device identifier |
| Account ID | `account_67890` | Account or subscription identifier |
| Custom String | `custom_identifier_1` | Any string-based identifier |

---

## Request/Response Examples

### Example 1: Create and Populate a Canary Tag

**Step 1: Add devices to canary tag**

```bash
curl -X PUT "http://localhost:9001/taggingService/tags/canary_group_1/members" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '[
    "AA:BB:CC:DD:EE:01",
    "AA:BB:CC:DD:EE:02",
    "AA:BB:CC:DD:EE:03"
  ]'
```

**Step 2: Verify tag members**

```bash
curl -X GET "http://localhost:9001/taggingService/tags/canary_group_1/members" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Step 3: Calculate percentages for devices**

```bash
curl -X GET "http://localhost:9001/taggingService/tags/members/AA:BB:CC:DD:EE:01/percentages/calculation" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### Example 2: Batch Update Production Devices

```bash
# Add devices
curl -X PUT "http://localhost:9001/taggingService/tags/production_devices/members" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '[
    "device_prod_001",
    "device_prod_002",
    "device_prod_003",
    "device_prod_004",
    "device_prod_005"
  ]'

# Remove outdated device
curl -X DELETE "http://localhost:9001/taggingService/tags/production_devices/members" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '[
    "device_prod_001"
  ]'

# Get all production devices
curl -X GET "http://localhost:9001/taggingService/tags/production_devices/members" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### Example 3: Find All Tags for a Device

```bash
curl -X GET "http://localhost:9001/taggingService/tags/members/device_prod_002" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

---

## Error Handling

### Common Error Responses

**400 Bad Request - Missing Tag**
```json
{
  "status": 400,
  "message": "tag is not specified"
}
```

**400 Bad Request - Empty Member List**
```json
{
  "status": 400,
  "message": "member list is empty"
}
```

**400 Bad Request - Batch Size Exceeded**
```json
{
  "status": 400,
  "message": "batch size 1500 exceeds the limit of 1000"
}
```

**404 Not Found**
```json
{
  "status": 404,
  "message": "production_devices tag not found"
}
```

**500 Internal Server Error**
```json
{
  "status": 500,
  "message": "Internal server error occurred while processing the request"
}
```

### HTTP Status Code Reference

| Code | Status | Description |
|------|--------|-------------|
| 200 | OK | Request successful |
| 204 | No Content | Successful deletion/removal |
| 400 | Bad Request | Invalid input, missing parameters, batch size exceeded |
| 401 | Unauthorized | Authentication failed or missing token |
| 404 | Not Found | Resource not found |
| 500 | Internal Server Error | Server-side processing error |

---

## Usage Patterns

### Pattern 1: Percentage-Based Rollout

```
1. Create/identify target tag: "firmware_v2.1_rollout"
2. Add devices to tag in batches of 1000
3. For each device, calculate percentage value
4. Select devices where percentage < 10 for initial rollout
5. Monitor rollout success
6. Gradually increase percentage threshold (10 → 25 → 50 → 100)
7. Remove tag when rollout complete
```

### Pattern 2: A/B Testing Scenarios

```
1. Create two tags: "feature_old_version" and "feature_new_version"
2. Add 50% of devices to each tag
3. Use percentage calculation for consistent split
4. Monitor metrics for both groups
5. Compare outcomes
6. Migrate all devices to winning version
7. Remove losing version tag
```

### Pattern 3: Device Segmentation by Model

```
1. Create model-specific tags:
   - "model_x1_devices"
   - "model_x2_devices"
   - "model_x3_devices"
2. Populate tags during device registration
3. Query tags by device to determine applicable configs
4. Remove devices from old model tags when replaced
```

### Pattern 4: Geographic Distribution

```
1. Create region-specific tags:
   - "region_us_west_devices"
   - "region_us_east_devices"
   - "region_eu_devices"
2. Assign devices based on registration location
3. Deploy region-specific configurations
4. Monitor regional performance metrics
```

---

## Best Practices

### Tag Naming Conventions

✓ **Do:**
- Use lowercase with underscores: `production_devices`, `beta_testers`
- Include environment prefix: `prod_`, `staging_`, `dev_`
- Use descriptive names: `model_x1_devices`, `region_us_west`
- Keep names concise (max 64 characters recommended)
- Use consistent naming patterns

✗ **Don't:**
- Use special characters (except underscores)
- Create ambiguous names without context
- Mix naming conventions in same system
- Use names that are too long

### Member Management

**Batch Operations:**
- Use batch operations for multiple members
- Maximum 1000 members per request
- Split large operations into multiple requests
- Prefer PUT/DELETE over individual operations

**Validation:**
- Validate member identifiers before adding
- Ensure consistency in identifier format
- Remove duplicate members before adding
- Check for typos in device IDs

**Monitoring:**
- Track tag membership changes
- Audit tag modifications
- Monitor tag sizes and growth
- Alert on unexpected changes

### Performance Considerations

**Query Optimization:**
- Use `/tags?full` only when needed
- Query specific tags rather than listing all
- Cache tag data locally when appropriate
- Batch read operations

**Scaling:**
- Split large tags if exceeding 10,000 members
- Use logical tag hierarchies for organization
- Monitor API response times
- Implement client-side caching

### Security Best Practices

**Authentication:**
- Always use valid JWT tokens
- Rotate tokens regularly
- Use secure token storage
- Implement token expiration

**Authorization:**
- Verify user permissions before operations
- Implement least-privilege access
- Audit all tag modifications
- Log sensitive operations

**Input Validation:**
- Validate all member identifiers
- Sanitize tag names
- Check batch sizes
- Verify request parameters

---

## Rate Limiting and Quotas

| Resource | Limit | Notes |
|----------|-------|-------|
| Batch Member Operations | 1000 members/request | Split into multiple requests if needed |
| Tag Name Length | 256 characters | Recommended max: 64 characters |
| Member ID Length | 256 characters | Depends on identifier format |
| Concurrent Connections | Per server config | Default: 100 connections |
| API Request Rate | Unlimited | Subject to authentication limits |
| Tag Count | No documented limit | Monitor system performance |
| Members per Tag | No documented limit | Recommend splitting if > 10,000 |

---

## Troubleshooting

### Common Issues

**Issue: 404 Tag Not Found**
- Verify tag name spelling and case sensitivity
- Check if tag was recently deleted
- List all available tags: `GET /taggingService/tags`

**Issue: 400 Batch Size Exceeded**
- Split requests into chunks of max 1000 members
- Use loop to add members in batches
- Validate batch size before sending

**Issue: 401 Unauthorized**
- Verify JWT token validity and expiration
- Check Authorization header format: `Bearer <TOKEN>`
- Regenerate token if expired

**Issue: Member Not Found in Percentage Calculation**
- Verify member exists in at least one tag
- Check member identifier format
- Ensure no special character escaping issues

**Issue: Slow Tag Queries**
- Reduce tag size by splitting into multiple tags
- Use specific tag queries instead of listing all
- Implement client-side caching
- Check network connectivity

---

## API Integration Guide

### Integration with XConf Admin Service

**Tag-based Configuration Targeting:**
- Define configuration rules using tag criteria
- Deploy different configurations to different tag groups
- Support gradual canary rollouts via percentage-based tags
- Enable automatic rollback by removing devices from tags

### Integration with XConf WebConfig Service

**Device Query on Configuration Request:**
- WebConfig queries tags to determine applicable configs
- Uses tag membership to evaluate targeting rules
- Applies configuration changes based on tag presence
- Supports multi-tag device membership

### Integration with Feature Control (RFC)

**Feature Rollout Management:**
- RFC rules can target devices by tag membership
- Percentage-based distribution within tags
- A/B testing across tag populations
- Feature graduation: dev tag → beta tag → production tag

### Integration with Telemetry

**Data Collection Segmentation:**
- Telemetry profiles can target specific device tags
- Analytics segmentation by tag-based device groups
- Performance monitoring per device cohort
- Custom metrics collection for tagged device populations

---

## Support and Resources

- **API Documentation**: This document
- **Integration Guide**: See "API Integration Guide" section above
- **Troubleshooting**: See "Troubleshooting" section above
- **XConf Main Documentation**: Reference overview.md
- **Best Practices**: See "Best Practices" section above

---

*Last Updated: January 2026*
*Version: 1.0*
