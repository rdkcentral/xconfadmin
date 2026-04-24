# XConf Tagging API Documentation

## Overview
The XConf Tagging API provides comprehensive tag management capabilities for RDK device grouping and configuration targeting. It enables dynamic device categorization through tag-based systems, supporting both individual member management and percentage-based distribution for advanced deployment strategies.

**Base URL**: `/taggingService`

## Authentication
All endpoints require authentication via JWT token or session-based authentication. Each request is validated for appropriate permissions based on the tagging operations being performed.

---

# Tag Management APIs

## Get All Tags
Retrieve all available tags in the system:

**GET** `http://<host>:<port>/taggingService/tags`

**Query Parameters:**
- `full` (optional): When present, returns complete tag objects with all members. Without this parameter, returns only tag IDs.

<details>
<summary><strong>Response Body (without full parameter):</strong> Array of tag IDs</summary>

```json
[
  "production_devices",
  "beta_testers", 
  "development_units",
  "canary_group_1"
]
```
</details>

<details>
<summary><strong>Response Body (with full parameter):</strong> Array of complete tag objects</summary>

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
</details>

Response Codes: 200, 500

---

## Get Tag by ID
Retrieve a specific tag by its identifier:

**GET** `http://<host>:<port>/taggingService/tags/{tag}`

**Path Parameters:**
- `tag` (required): Tag identifier

<details>
<summary><strong>Response Body:</strong> Tag object with complete information</summary>

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
</details>

Response Codes: 200, 400, 404

---

## Delete Tag
Delete a specific tag from the system:

**DELETE** `http://<host>:<port>/taggingService/tags/{tag}`

**Path Parameters:**
- `tag` (required): Tag identifier to delete

<details>
<summary><strong>Response Body:</strong> No content on successful deletion</summary>

```
No response body (HTTP 204 No Content)
```
</details>

Response Codes: 204, 400, 404

---

## Delete Tag Without Prefix (Development Only)
Delete a tag from XConf without prefix validation - intended for testing and cleanup purposes only:

**DELETE** `http://<host>:<port>/taggingService/tags/{tag}/noprefix`

**Path Parameters:**
- `tag` (required): Tag identifier to delete

**⚠️ Warning**: This endpoint should not be used in production environments as it bypasses standard prefix validation.

<details>
<summary><strong>Response Body:</strong> No content on successful deletion</summary>

```
No response body (HTTP 204 No Content)
```
</details>

Response Codes: 204, 400, 404

---

# Tag Member Management APIs

## Add Members to Tag
Add multiple members to an existing tag:

**PUT** `http://<host>:<port>/taggingService/tags/{tag}/members`

**Path Parameters:**
- `tag` (required): Tag identifier

**Request Limits:**
- Maximum batch size: 1000 members per request

<details>
<summary><strong>Request Body:</strong> Array of member identifiers</summary>

```json
[
  "AA:BB:CC:DD:EE:05",
  "AA:BB:CC:DD:EE:06",
  "AA:BB:CC:DD:EE:07",
  "device_id_12345"
]
```
</details>

<details>
<summary><strong>Response Body:</strong> Success confirmation</summary>

```
No response body (HTTP 200 OK)
```
</details>

Response Codes: 200, 400, 500

---

## Remove Member from Tag
Remove a single member from a specific tag:

**DELETE** `http://<host>:<port>/taggingService/tags/{tag}/members/{member}`

**Path Parameters:**
- `tag` (required): Tag identifier
- `member` (required): Member identifier to remove

<details>
<summary><strong>Response Body:</strong> Updated tag object</summary>

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
</details>

Response Codes: 204, 400, 404, 500

---

## Remove Multiple Members from Tag
Remove multiple members from a specific tag in a single operation:

**DELETE** `http://<host>:<port>/taggingService/tags/{tag}/members`

**Path Parameters:**
- `tag` (required): Tag identifier

**Request Limits:**
- Maximum batch size: 1000 members per request

<details>
<summary><strong>Request Body:</strong> Array of member identifiers to remove</summary>

```json
[
  "AA:BB:CC:DD:EE:02",
  "AA:BB:CC:DD:EE:05",
  "device_id_old_123"
]
```
</details>

<details>
<summary><strong>Response Body:</strong> No content on successful removal</summary>

```
No response body (HTTP 204 No Content)
```
</details>

Response Codes: 204, 400, 500

---

## Get Tag Members
Retrieve all members of a specific tag:

**GET** `http://<host>:<port>/taggingService/tags/{tag}/members`

**Path Parameters:**
- `tag` (required): Tag identifier

<details>
<summary><strong>Response Body:</strong> Array of member identifiers</summary>

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
</details>

Response Codes: 200, 400, 404, 500

---

## Get Tags by Member
Retrieve all tags that contain a specific member:

**GET** `http://<host>:<port>/taggingService/tags/members/{member}`

**Path Parameters:**
- `member` (required): Member identifier to search for

<details>
<summary><strong>Response Body:</strong> Array of tag objects containing the member</summary>

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
</details>

Response Codes: 200, 400, 500

---

# Percentage-Based Distribution APIs

## Calculate Percentage Value for Member
Calculate the percentage value for a specific member using SipHash algorithm:

**GET** `http://<host>:<port>/taggingService/tags/members/{member}/percentages/calculation`

**Path Parameters:**
- `member` (required): Member identifier for percentage calculation

**Algorithm Details:**
- Uses SipHash with fixed keys for consistent percentage calculation
- Returns deterministic percentage value (0-99) for the given member
- Enables percentage-based rollout strategies

<details>
<summary><strong>Response Body:</strong> Calculated percentage value</summary>

```json
67
```
</details>

Response Codes: 200, 400, 500

---

# Data Models

## Tag Object
Represents a tag entity with its associated members:

```json
{
  "id": "string",           // Unique tag identifier
  "members": ["string"],    // Array of member identifiers
  "updated": "integer"      // Unix timestamp of last update
}
```

**Field Descriptions:**
- `id`: Unique identifier for the tag (required)
- `members`: Array of device MAC addresses, device IDs, or other identifiers
- `updated`: Unix timestamp in milliseconds indicating when the tag was last modified

## Member Identifier Formats
Members can be identified using various formats:

- **MAC Address**: `AA:BB:CC:DD:EE:FF`
- **Device ID**: `device_12345`
- **Account ID**: `account_67890`
- **Custom Identifier**: Any string-based identifier

---

# Usage Examples

## Creating a Canary Deployment Tag

### Step 1: Add devices to a canary tag
```bash
curl -X PUT "http://localhost:9001/taggingService/tags/canary_group_1/members" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '[
    "AA:BB:CC:DD:EE:01",
    "AA:BB:CC:DD:EE:02", 
    "AA:BB:CC:DD:EE:03"
  ]'
```

### Step 2: Verify tag membership
```bash
curl -X GET "http://localhost:9001/taggingService/tags/canary_group_1/members" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Step 3: Calculate percentage for device
```bash
curl -X GET "http://localhost:9001/taggingService/tags/members/AA:BB:CC:DD:EE:01/percentages/calculation" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Managing Production Device Groups

### Add multiple devices to production tag
```bash
curl -X PUT "http://localhost:9001/taggingService/tags/production_devices/members" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '[
    "device_prod_001",
    "device_prod_002",
    "device_prod_003"
  ]'
```

### Remove devices from tag
```bash
curl -X DELETE "http://localhost:9001/taggingService/tags/production_devices/members" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '[
    "device_prod_001"
  ]'
```

### Find all tags for a specific device
```bash
curl -X GET "http://localhost:9001/taggingService/tags/members/device_prod_002" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

# Error Handling

## Common Error Responses

### 400 Bad Request
```json
{
  "status": 400,
  "message": "tag is not specified"
}
```

### 404 Not Found  
```json
{
  "status": 404,
  "message": "production_devices tag not found"
}
```

### 400 Batch Size Exceeded
```json
{
  "status": 400,
  "message": "batch size 1500 exceeds the limit of 1000"
}
```

### 400 Empty Member List
```json
{
  "status": 400,
  "message": "member list is empty"
}
```

## Error Codes Summary

| Code | Description | Common Causes |
|------|-------------|---------------|
| 200  | Success | Request processed successfully |
| 204  | No Content | Successful deletion or removal |
| 400  | Bad Request | Missing parameters, invalid input, batch size exceeded |
| 401  | Unauthorized | Invalid or missing authentication token |
| 404  | Not Found | Tag or member not found |
| 500  | Internal Server Error | Server processing error |

---

# Best Practices

## Tag Naming Conventions
- Use descriptive, lowercase names with underscores: `production_devices`, `beta_testers`
- Include environment prefixes: `prod_`, `staging_`, `dev_`
- Use meaningful groupings: `model_specific`, `region_based`, `feature_enabled`

## Member Management
- **Batch Operations**: Use batch add/remove operations for better performance
- **Validation**: Ensure member identifiers are valid before adding to tags
- **Monitoring**: Track tag membership changes for audit purposes
- **Cleanup**: Regularly remove inactive or decommissioned device members

## Percentage-Based Rollouts
- **Deterministic Distribution**: Use the percentage calculation API for consistent device assignment
- **Gradual Rollouts**: Start with small percentages and gradually increase
- **Monitoring**: Track rollout progress and device health during percentage-based deployments
- **Rollback Strategy**: Maintain ability to quickly remove devices from percentage-based tags

## Security Considerations
- **Authentication**: Always use proper JWT tokens for API access
- **Authorization**: Ensure users have appropriate permissions for tag operations
- **Input Validation**: Validate all member identifiers before processing
- **Rate Limiting**: Respect batch size limits to prevent system overload

---

# Integration with XConf Configuration Management

The Tagging API integrates seamlessly with XConf's configuration management system:

## Firmware Rules Integration
- Tags can be used as targeting criteria in firmware rules
- Percentage-based rollouts enable gradual firmware deployment
- Tag membership automatically affects device configuration eligibility

## DCM Configuration Targeting
- Device Control Manager settings can target specific tags
- Log upload policies can be applied to tagged device groups
- Diagnostic data collection can be configured per tag

## RFC Feature Management
- Feature flags can target devices based on tag membership
- A/B testing scenarios can leverage percentage-based tag distribution
- Feature rollouts can be controlled through dynamic tag membership

## Telemetry Profile Assignment
- Telemetry collection policies can be assigned to specific tags
- Analytics data can be segmented based on tag-based device groupings
- Performance monitoring can target tagged device populations

This tagging system provides the foundation for sophisticated device management, enabling precise control over configuration distribution, feature rollouts, and operational policies across diverse RDK device deployments.