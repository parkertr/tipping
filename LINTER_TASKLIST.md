# Linter Task List

## Rules
- Always ask the user for confirmation before running any git commands (add, commit, push, etc.).
- After each successful linter fix:
  1. Run tests automatically with `go test ./...` (without -v flag to only show failures)
  2. Confirm the linter rule has been fixed by running the linter
  3. Only then proceed with:
     - `git add .` to stage all changed files
     - `git commit --no-verify -m "Fix: <brief description of the fix>"` to commit the changes
     - Use `--no-verify` while there are linter issues to fix

## High-priority linter errors (examples, not exhaustive):

### 1. Line length exceeds 120 characters (lll) ✅ FIXED
- [x] `internal/infrastructure/api/middleware/auth.go:23`
  - Split AuthMiddleware function signature into multiple lines
- [x] `internal/infrastructure/repository/postgres/prediction_repository.go:103`
  - Split long error message in GetByUserAndMatch into multiple lines

### 2. Magic numbers (mnd) ✅ FIXED
- [x] `cmd/api/main.go:51,52,53,54,55,73`
  - Created constants for server timeouts and configuration
- [x] `internal/auth/jwt.go:40`
  - Added constant for token expiration time
- [x] `internal/domain/prediction.go:39`
  - Added constants for prediction points
- [x] `internal/domain/user.go:83`
  - Added constant for initial stats values
- [x] `internal/infrastructure/api/handlers/matches.go:275`
  - Added constant for initial score values

### 3. Structs missing json tags (musttag) ✅ FIXED
- [x] `internal/infrastructure/api/handlers/predictions.go:80,94,103`
  - Added json tags to anonymous structs in event handling
  - Added tags for MatchCreated, MatchScoreUpdated, and MatchStatusChanged events

### 4. Return both nil error and invalid value (nilnil) ✅ FIXED
- [x] `internal/infrastructure/repository/postgres/user_repository.go:209`
  - Updated queryUser function to return a proper error when no rows are found
  - Changed `return nil, nil` to `return nil, fmt.Errorf("user not found")`

### 5. No blank line before continue/break/return (nlreturn) ✅ FIXED
- [x] `cmd/import-fixtures/main.go:94,100,117`
  - Added blank lines before continue statements
- [x] `internal/infrastructure/api/handlers/auth_handler.go:319`
  - Added blank line before break statement
- [x] `internal/infrastructure/eventhandlers/match_handler.go:66,70,71,102,103,135`
  - Added blank lines before return statements
- [x] `internal/domain/match.go:76`
  - Added blank line before return statement
- [x] `internal/domain/prediction.go:62`
  - Added blank line before return statement
- [x] `internal/domain/user.go:34`
  - Added blank line before return statement
- [x] `internal/auth/jwt.go:52,65`
  - Added blank lines before return statements

### 6. Unused parameter (revive)
- [ ] `internal/auth/jwt.go:56`

### 7. Rows.Err must be checked (rowserrcheck)
- [ ] `internal/infrastructure/eventstore/postgres.go:64,125,186`

### 8. Use t.Cleanup instead of defer in tests (tparallel)
- [ ] `internal/infrastructure/api/server/server_test.go:31`

### 9. Use http.Method* instead of string (usestdlibvars)
- [ ] `internal/infrastructure/api/handlers/matches_test.go:43,82,106,141,172`
- [ ] `internal/infrastructure/api/handlers/predictions_test.go:38`
- [ ] `internal/infrastructure/api/server/middleware.go:25`

### 10. Use t.Context() in tests (usetesting)
- [ ] `internal/infrastructure/eventstore/eventstore_test.go:61,105,153,241`

### 11. Variable/parameter name too short (varnamelen)
- [ ] Multiple locations (see linter output)

### 12. Error returned from external package is unwrapped (wrapcheck)
- [ ] Multiple locations (see linter output)

### 13. Cuddling/whitespace issues (wsl)
- [ ] Multiple locations (see linter output)

---

**Note:**
- There are over 200 linter issues. This list contains the most common and high-priority types, but you should refer to the full linter output for all details and line numbers.
- Fixing these will improve code quality, readability, and maintainability.
