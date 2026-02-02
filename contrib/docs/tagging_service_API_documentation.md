# XConf Tagging Service API

## Table of Contents

1. [Performance Information](#performance-information)
   - [Add Members API](#add-members-api)
   - [Delete Members API](#delete-members-api)
2. [API Endpoints](#api-endpoints)
   - [SAT Token Requirements](#sat-token-requirements)
   - [Get Tag by ID](#get-tag-by-id)
   - [Delete Tag by ID (Asynchronous)](#delete-tag-by-id-asynchronous)
   - [Add Members to Tag](#add-members-to-tag)
   - [Remove Members from Tag](#remove-members-from-tag)
   - [Remove Member from Tag](#remove-member-from-tag)
   - [Get Tag Members](#get-tag-members)
3. [XConf Rule Configuration with Tags](#xconf-rule-configuration-with-tags)

---

## Performance Information

### Add Members API

The Add Members API processes member additions in batches, with each batch supporting up to 2,000 members.

### Delete Members API

The Delete Members API also operates in batches of up to 2,000 members per batch.

---

## API Endpoints

### SAT Token Requirements

Client should have following SAT capabilities:
- `"x1:coast:cmtagds:assign"`
- `"x1:coast:cmtagds:read"`
- `"x1:coast:cmtagds:unassign"`
- `"x1:coast:xconf:read"`
- `"x1:coast:xconf:read:maclist"`
- `"x1:coast:xconf:write"`
- `"x1:coast:xconf:write:maclist"`

---

### Get Tag by ID

Returns representation of XConf tag by provided tag id

**Endpoint:**
```
GET /taggingService/tags/{id}
```

**Headers:**
```
Accept = application/json
Content-Type = application/json
Authorization = Bearer {SAT token}
```

**Response Status Codes:**
- `200 OK`
- `404 NOT FOUND`

**Response Body:**
```json
{
    "id": "test:tag:demotag",
    "description": "",
    "members": [
        "A2:A2:A2:A2:B2:B2"
    ],
    "updated": 1711651165855
}
```

---

### Delete Tag by ID (Asynchronous)

Deletes a tag and all its members asynchronously. The API returns immediately after validation, and the actual deletion is processed in the background.

**Endpoint:**
```
DELETE /taggingService/tags/{id}
```

**Headers:**
```
Accept = application/json
Content-Type = application/json
Authorization = Bearer {SAT token}
```

**Success Response (202 Accepted):**
The tag deletion request has been accepted and queued for processing.

**Status Code:** `202 Accepted`

**Response Body:**
```json
{
    "status": "accepted",
    "message": "Tag 'my-tag' deletion has been queued for processing",
    "tag": "my-tag"
}
```

#### Behavior

- **Immediate Response:** API returns 202 Accepted immediately after validating that the tag exists
- **Background Processing:** Tag deletion (including all members and buckets) happens asynchronously
- **No Status Tracking:** Currently no endpoint to check deletion progress (work is pending)
- **Error Handling:** Any errors during background deletion are logged server-side

#### Notes

- The 202 Accepted status indicates the request was valid and accepted, not that deletion is complete
- For large tags with many members, deletion may take several minutes
- Once accepted, the deletion cannot be cancelled

---

### Add Members to Tag

Adds new members to the tag. If tag does not exist – new tag is created in XConf. By default XConf does tag member normalization: whitespaces are trimmed, string data is set to upper case.

**Endpoint:**
```
PUT /taggingService/tags/{tag}/members
```

**Headers:**
```
Accept = application/json
Content-Type = application/json
Authorization = Bearer {SAT token}
```

**Request Body - list of members:**
```json
["A1:A1:A1:A1:B1:B1", "A2:A2:A2:A2:B2:B2"]
```

**Response Status Code:** `202 Accepted`

**Response Body - XConf tag entity with added members:**
```json
{
    "id": "test:tag:demotag",
    "description": "",
    "members": [
        "A1:A1:A1:A1:B1:B1",
        "A2:A2:A2:A2:B2:B2"
    ],
    "updated": 1711651165855
}
```

---

### Remove Members from Tag

Removes members from the tag. If all members are removed, the tag is automatically deleted.

**Endpoint:**
```
DELETE /taggingService/tags/{tag}/members
```

**Headers:**
```
Accept = application/json
Content-Type = application/json
Authorization = Bearer {SAT token}
```

**Request Body - list of members:**
```json
["A1:A1:A1:A1:B1:B1", "A2:A2:A2:A2:B2:B2"]
```

**Response Status Codes:**
- `404 NOT FOUND`
- `204 NO CONTENT`

---

### Remove Member from Tag

Removes member record from XDAS first, in case of success removes tag member from XConf. Remove API takes non-normalized data, normalization is done by XConf.

**Endpoint:**
```
DELETE /taggingService/tags/{tag}/members/{member}
```

**Headers:**
```
Accept = application/json
Content-Type = application/json
Authorization = Bearer {SAT token}
```

**Response Status Code:** `204 NO CONTENT`

---

### Get Tag Members

Retrieves all members of a specified tag. Supports both non-paginated (V1 compatible) and paginated responses.

**Endpoint:**
```
GET /taggingService/tags/{tag}/members
```

**Headers:**
```
Accept = application/json
Content-Type = application/json
Authorization = Bearer {SAT token}
```

#### Query Parameters (Optional - for pagination)

| Parameter | Type | Required | Default | Maximum | Description |
|-----------|------|----------|---------|---------|-------------|
| `limit` | integer | No | 500 | 5000 | Number of members to return per page. Must be a positive integer. If exceeds maximum, returns 400 Bad Request |
| `cursor` | string | No | - | - | Pagination cursor for retrieving the next page of results. Obtained from nextCursor field in the previous response |

**Note:** If either limit or cursor is provided, the endpoint returns a paginated response. Otherwise, it returns a non-paginated response.

#### Response Status Codes

- `200 OK`: Successfully retrieved members
- `206 Partial Content`: Response contains only first 100,000 members (tag has more than 100k members)
- `400 Bad Request`: Invalid tag or query parameters
- `404 Not Found`: Tag does not exist

#### Non-Paginated Response Body

```json
[
    "A2:A2:A2:A2:B2:B2"
]
```

**Important:** In non-paginated mode, if a tag has more than 100,000 members, the response will be truncated to the first 100,000 members and the status code will be 206 Partial Content. To retrieve all members of large tags, use paginated mode.

#### Paginated Mode

Used when limit and/or cursor query parameters are provided.

**Response Status Codes:**
- `200 OK`: Successfully retrieved page of members
- `400 Bad Request`: Invalid query parameters or tag parameter
- `404 Not Found`: Tag does not exist

**Response Body:**
```json
{
    "data": [
        "A2:A2:A2:A2:B2:B2",
        "C3:C3:C3:C3:D3:D3",
        "E4:E4:E4:E4:F4:F4"
    ],
    "nextCursor": "eyJidWNrZXQiOjEyLCJsYXN0S2V5IjoiQTI6QTI6QTI6QTI6QjI6QjIifQ==",
    "hasMore": true
}
```

**Response Fields:**
- `data` (array of strings): List of member identifiers in the current page
- `nextCursor` (string, optional): Cursor for the next page. Omitted if there are no more results.
- `hasMore` (boolean): Indicates whether more results are available
  - `true`: More members available, use nextCursor to retrieve next page
  - `false`: No more members to retrieve

##### Example Requests

**Non-Paginated (all members, up to 100k):**
```
GET /taggingService/tags/my-tag-123/members
```

**Paginated (first page with custom limit):**
```
GET /taggingService/tags/my-tag-123/members?limit=1000
```

**Paginated (subsequent page):**
```
GET /taggingService/tags/my-tag-123/members?limit=1000&cursor=eyJidWNrZXQiOjEyLCJsYXN0S2V5IjoiQTI6QTI6QTI6QTI6QjI6QjIifQ==
```

##### Pagination Workflow

1. Make initial request with optional limit parameter
2. Process the data array containing members
3. Check hasMore field:
   - If `true`: Use the nextCursor value as the cursor parameter in the next request
   - If `false`: All members have been retrieved
4. Repeat until hasMore is false

---

## XConf Rule Configuration with Tags

### Steps to Configure Rules with Tags

1. **Create New Firmware Rule with the tag as the condition using EXISTS operation.**

2. **Add needed MAC address or any other parameters to the tag using "Add member to tag" API:**

```bash
curl --location --request PUT 'http://<xconf-admin-url>/taggingService/tags/xconf:tag:usage:demo/members' \
  --header 'Authorization: Bearer <SAT token>' \
  --header 'Accept: application/json' \
  --header 'Content-Type: application/json' \
  --data '["BB:BB:BB:BB:BB:BB"]'
```

3. **Trigger /swu/xconf/ API to evaluate the rules, make sure that tag member from step 2 is present as in the request parameters of /swu/xconf/ query:**

```bash
curl --location 'http://<xconf-url>/xconf/swu/stb?model=TESTMODEL&eStbMac=BB%3ABB%3ABB%3ABB%3ABB%3ABB&firmwareVersion=TEST_VERSION'
```

**Example Response:**
```json
{
    "firmwareDownloadProtocol": "tftp",
    "firmwareFilename": "filename.t",
    "firmwareVersion": "TEST_VERSION_TAGGING_USAGE",
    "mandatoryUpdate": false,
    "rebootImmediately": false
}
```

---

*Document Version: 1.0*
*Last Updated: January 2026*
