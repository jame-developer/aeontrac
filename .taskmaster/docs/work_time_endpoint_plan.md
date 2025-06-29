# Plan for New Work Time REST Endpoint

This document outlines the plan for creating a new REST endpoint to add work time entries.

## 1. Endpoint Definition

-   **HTTP Method:** `POST`
-   **URL Path:** `/worktime`

This follows the RESTful convention of using `POST` to create a new resource. The path `/worktime` is concise and clearly indicates the resource being manipulated.

## 2. Request Payload

The request body will be a JSON object containing the details of the work time entry.

**Structure:**

```json
{
  "date": "2025-06-29",
  "start": "2025-06-29T09:00:00Z",
  "stop": "2025-06-29T17:30:00Z",
  "comment": "Worked on implementing the new REST endpoint."
}
```

**Fields:**

-   `date` (string, required): The date of the work entry in `YYYY-MM-DD` format. This will be used to find or create the correct `AeonDay`.
-   `start` (string, required): The start time of the work entry in RFC3339 format.
-   `stop` (string, required): The end time of the work entry in RFC3339 format.
-   `comment` (string, optional): A description of the work performed.

**Note on Identifiers:** The initial request does not include `userId` or `projectId` fields. The current data model (`AeonVault`) appears to be for a single user's context, so these are omitted for consistency. If multi-user or per-project tracking is needed in the future, the model and this endpoint will need to be extended.

## 3. Response Payloads

### Success Response

-   **Status Code:** `201 Created`
-   **Body:** A JSON object representing the newly created `AeonUnit`, including its server-generated UUID and calculated duration.

**Structure:**

```json
{
  "id": "a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6",
  "start": "2025-06-29T09:00:00Z",
  "stop": "2025-06-29T17:30:00Z",
  "duration": "8h30m0s",
  "type": "WORK",
  "comment": "Worked on implementing the new REST endpoint."
}
```

### Error Responses

-   **Status Code:** `400 Bad Request`
    -   **Reason:** Sent when the request payload is malformed or fails validation (e.g., invalid date format, `stop` time is before `start` time).
    -   **Body:**
        ```json
        {
          "error": "Invalid request payload",
          "details": "Validation error message here"
        }
        ```

-   **Status Code:** `500 Internal Server Error`
    -   **Reason:** Sent when an unexpected error occurs on the server (e.g., unable to read from or write to the data store).
    -   **Body:**
        ```json
        {
          "error": "Internal server error"
        }
        ```

## 4. Handler Logic

The request handler will perform the following steps:

1.  **Bind and Validate:** Parse the incoming JSON request into a `WorkTimeRequest` struct. Validate the data types, formats, and logical constraints (e.g., `stop` > `start`).
2.  **Load Data:** Load the `AeonVault` from the persistent data store.
3.  **Find or Create Day:** Use the `date` from the request to find the corresponding `AeonDay` in the `AeonVault`. If it does not exist, create and initialize a new `AeonDay` instance.
4.  **Create Work Unit:**
    -   Generate a new `uuid.UUID` for the work entry.
    -   Create a new `models.AeonUnit` struct.
    -   Set the `Start` and `Stop` times from the request.
    -   Set the `Type` to `"WORK"`.
    -   Set the `Comment` from the request.
    -   Calculate the `Duration`.
5.  **Update Day:** Add the new `AeonUnit` to the `units` map of the `AeonDay`.
6.  **Recalculate Totals:** Recalculate the `TotalHours` and `OvertimeHours` for the affected `AeonDay`.
7.  **Save Data:** Persist the updated `AeonVault` back to the data store.
8.  **Respond:** Return a `201 Created` status with the JSON representation of the newly created `AeonUnit`.

## 5. Affected Files

The following files will be created or modified to implement this feature:

1.  **`internal/api/router/router.go`** (Modified):
    -   Add a new route `r.POST("/worktime", handlers.AddWorkTimeHandler)`.

2.  **`internal/api/handlers/worktime.go`** (New File):
    -   Create this new file to house the `AddWorkTimeHandler(c *gin.Context)` function. This handler will contain the logic described in the "Handler Logic" section.

3.  **`pkg/models/requests.go`** (New or Modified File):
    -   A new file could be created for request-specific models, or an existing one could be used.
    -   Define a `WorkTimeRequest` struct for binding the incoming JSON payload.

4.  **`internal/service/worktime_service.go`** (New File):
    -   Create a new service to encapsulate the business logic of adding a work time entry, separating it from the HTTP handler. It would contain a function like `AddWorkTimeEntry(request models.WorkTimeRequest) (*models.AeonUnit, error)`.

5.  **`internal/store/vault.go`** (Modified, assumed to exist):
    -   The existing data access layer will likely need no changes if it already provides generic `LoadVault()` and `SaveVault(vault *models.AeonVault)` functions. If not, these functions will need to be implemented.