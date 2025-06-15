# Bugs Task List

## Rules
- An item in this list is only considered fixed once the corresponding test passes.
- After each successful bug fix:
  1. Run `git add .` to stage all changed files
  2. Run `git commit --no-verify -m "Fix: <brief description of the fix>"` to commit the changes
  3. The `--no-verify` flag is used while we still have linter issues to fix

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
  - [x] In `server_test.go`, update the expected status codes to match the actual behavior:
    - `/api/auth/google` should expect 307 (Temporary Redirect) for OAuth redirect
    - `/api/auth/google/callback` should expect 400 (Bad Request) when missing required query params

---

## 3. internal/infrastructure/eventstore
**Test:** `TestGetEvents`
- **Status:** ✅ FIXED
- **Fix:** Updated JSON field names in test assertions to match struct tags in events package:
  - Changed "ID" to "id"
  - Changed "HomeTeam" to "homeTeam"
  - Changed "AwayTeam" to "awayTeam"
  - Changed "Competition" to "competition"
  - Changed "UserID" to "userId"
  - Changed "MatchID" to "matchId"
  - Changed "HomeGoals" to "homeGoals"
  - Changed "AwayGoals" to "awayGoals"
  - Fixed numeric type conversion for goals fields
