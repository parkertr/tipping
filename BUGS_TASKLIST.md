# BUGS TASK LIST

## Rules
- Always ask the user for confirmation before running any git commands (add, commit, push, etc.).
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
- **Status:** ✅ FIXED
- **Fix:** The test file already had the correct status code expectations:
  - `/api/auth/google` correctly expects 307 (Temporary Redirect) for OAuth redirect
  - `/api/auth/google/callback` correctly expects 400 (Bad Request) when query parameters are missing

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

---

## 5. internal/infrastructure/api/server
**Test:** `TestMiddleware`
- **Status:** ✅ FIXED
- **Fix:** Updated the middleware test to properly test authentication:
  - Added test cases for different authentication scenarios:
    - No auth header
    - Invalid auth header format
    - Invalid token
  - Each test case verifies the correct 401 Unauthorized response
  - Removed unused mock variable to fix linter error

## 6. internal/infrastructure/api/server
**Test:** `TestServerClose`
- **Status:** ✅ FIXED
- **Fix:** Updated the server's Close method to properly clean up resources:
  - Added cleanup for the event store if it implements io.Closer
  - Removed attempts to close repositories as they don't implement Close methods
  - The test now passes as the Close method properly handles cleanup

## 7. internal/infrastructure/api/server
**Test:** `TestNewServer`
- **Status:** ✅ FIXED
- **Fix:** Updated the test to properly set up mock database expectations:
  - Added ExpectPing to the mock database
  - Verified all expectations were met after server creation
  - The test now passes as the server is properly initialized with the mock database
