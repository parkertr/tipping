# Linter Task List

## High-priority linter errors (examples, not exhaustive):

### 1. Line length exceeds 120 characters (lll)
- [ ] `internal/infrastructure/api/middleware/auth.go:23`
- [ ] `internal/infrastructure/repository/postgres/prediction_repository.go:103`

### 2. Magic numbers (mnd)
- [ ] `cmd/api/main.go:51,52,53,54,55,73`
- [ ] `internal/auth/jwt.go:40`
- [ ] `internal/domain/prediction.go:39`
- [ ] `internal/domain/user.go:83`
- [ ] `internal/infrastructure/api/handlers/matches.go:275`

### 3. Structs missing json tags (musttag)
- [ ] `internal/infrastructure/api/handlers/predictions.go:80,94,103`

### 4. Return both nil error and invalid value (nilnil)
- [ ] `internal/infrastructure/repository/postgres/user_repository.go:209`

### 5. No blank line before continue/break/return (nlreturn)
- [ ] `cmd/import-fixtures/main.go:94,100,117`
- [ ] `internal/infrastructure/api/handlers/auth_handler.go:319`
- [ ] `internal/infrastructure/api/handlers/mocks/eventstore.go:18`
- [ ] `internal/infrastructure/api/handlers/mocks/mock_repository.go:18,23`

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
