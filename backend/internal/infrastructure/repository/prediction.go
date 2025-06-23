package repository

// PredictionFilters represents the filters that can be applied when listing predictions
type PredictionFilters struct {
	UserID  *string
	MatchID *string
}
