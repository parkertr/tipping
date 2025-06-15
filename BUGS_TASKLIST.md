# Bugs Task List

## Rule
- An item in this list is only considered fixed once the corresponding test passes.

## 1. internal/infrastructure/api/handlers
**Test:** `TestGetUserPredictions/User_has_no_predictions`
- **Error:**
  - Panic: `mock: I don't know what to return because the method call was unexpected.`
  - The test calls `GetEventsByType` on a mock, but the mock was not set up with `.On("GetEventsByType")`.
- **Task:**
  - [ ] In `predictions_test.go`, set up the mock for `GetEventsByType` with `.On("GetEventsByType")` and a suitable `.Return(...)` for the test.

---

## 2. internal/infrastructure/api/server
**Test:** `TestServerRoutes`
- **Errors:**
  - `TestServerRoutes/GET_/api/auth/google`: Expected status code 200, got 307
  - `TestServerRoutes/GET_/api/auth/google/callback`: Expected status code 200, got 400
- **Task:**
  - [ ] In `server_test.go`, review the test setup and handlers for `/api/auth/google` and `/api/auth/google/callback` to ensure the correct status codes are returned.
  - [ ] Update the tests or the handler logic so that the expected status codes match the actual behavior.
