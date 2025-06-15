# Bugs Task List

## Rule
- An item in this list is only considered fixed once the corresponding test passes.

## 1. internal/infrastructure/api/handlers
**Test:** `TestGetUserPredictions/User_has_no_predictions`
- **Error:**
  - Panic: `mock: I don't know what to return because the method call was unexpected.`
  - The test calls `GetEventsByType` on a mock, but the mock was not set up with `.On("GetEventsByType")`.
- **Task:**
  - [x] In `predictions_test.go`, set up the mock for `GetEventsByType` with `.On("GetEventsByType")` and a suitable `.Return(...)` for the test.
  - [x] Ensure each test case uses its own mock to avoid state leakage between tests.
  - [x] Verify the handler returns an empty array for users with no predictions.

---

## 2. internal/infrastructure/api/server
**Test:** `TestServerRoutes`
- **Errors:**
  - `TestServerRoutes/GET_/api/auth/google`: Expected status code 200, got 307
  - `TestServerRoutes/GET_/api/auth/google/callback`: Expected status code 200, got 400
- **Task:**
  - [ ] In `server_test.go`, review the test setup and handlers for `/api/auth/google` and `/api/auth/google/callback` to ensure the correct status codes are returned.
  - [ ] Update the tests or the handler logic so that the expected status codes match the actual behavior.

---

## 3. internal/infrastructure/eventstore
**Test:** `TestGetEvents`
- **Errors:**
  - `TestGetEvents/Get_match_events`: Multiple field mismatches in event data:
    - Expected ID match123, got <nil>
    - Expected HomeTeam Team A, got <nil>
    - Expected AwayTeam Team B, got <nil>
    - Expected Competition Premier League, got <nil>
  - `TestGetEvents/Get_prediction_events`: Multiple field mismatches in event data:
    - Expected ID pred123, got <nil>
    - Expected UserID user123, got <nil>
    - Expected MatchID match123, got <nil>
    - Expected HomeGoals 2, got <nil>
    - Expected AwayGoals 1, got <nil>
- **Task:**
  - [ ] In `eventstore_test.go`, review the event data handling in `TestGetEvents` to ensure proper event data is being saved and retrieved.
  - [ ] Check if the event data is being properly serialized/deserialized in the event store implementation.
  - [ ] Verify that the test data setup matches the expected event structure.
