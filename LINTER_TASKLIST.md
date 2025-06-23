# Linter Task List

## High Priority
1. Unwrapped errors (wrapcheck)
   - [x] `internal/infrastructure/repository/postgres/user_repository.go:18,44,97,115,128,132,149`
   - [x] `internal/infrastructure/repository/postgres/match_repository.go:45,51,67,79,84,90,98,111,117,123,131`
   - [x] `internal/infrastructure/repository/postgres/prediction_repository.go:45,51,67,79,84,90,98,111,117,123,131`
   - [x] `internal/infrastructure/eventstore/postgres.go:22,32,49,66,79,87,93,99,105,114,131,144,152,158,164,170,179,196,209,217,223,229,235,244`
   - [x] `internal/infrastructure/eventhandlers/match_handler.go:45,51,67,79,84,90,98,111,117,123,131`

2. Short variable names (revive)
   - [x] `internal/infrastructure/api/server/server_test.go:31`
   - [x] `internal/infrastructure/api/handlers/auth_handler.go:45`
   - [x] `internal/infrastructure/api/handlers/matches.go:67`

3. Unused parameter (revive)
   - [x] `internal/auth/jwt.go:56`

4. Rows.Err must be checked (rowserrcheck)
   - [x] `internal/infrastructure/eventstore/postgres.go:64,125,186`

5. Use t.Cleanup instead of defer in tests (tparallel)
   - [x] `internal/infrastructure/api/server/server_test.go:31`

## Medium Priority
1. Package Documentation (godoc)
   - [x] `internal/auth/jwt.go`: Add package documentation

## Rules
- Run tests automatically after each change (without -v flag)
- Show only failed tests
- Commit changes after each fix
- Use descriptive commit messages
- Follow Go best practices
- Keep code clean and maintainable
- Document any complex changes
- Review changes before committing
- Run linter after each fix
- Update task list after each fix
- Treat all linter warnings as errors to be fixed
