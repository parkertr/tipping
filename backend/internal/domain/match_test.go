package domain

import (
	"testing"
	"time"
)

func TestNewMatch(t *testing.T) {
	// Test case: Create new match
	id := "match123"
	homeTeam := "Team A"
	awayTeam := "Team B"
	date := time.Now()
	competition := "Premier League"

	match := NewMatch(id, homeTeam, awayTeam, date, competition)

	if match.ID != id {
		t.Errorf("expected ID %v, got %v", id, match.ID)
	}
	if match.HomeTeam != homeTeam {
		t.Errorf("expected HomeTeam %v, got %v", homeTeam, match.HomeTeam)
	}
	if match.AwayTeam != awayTeam {
		t.Errorf("expected AwayTeam %v, got %v", awayTeam, match.AwayTeam)
	}
	if !match.Date.Equal(date) {
		t.Errorf("expected Date %v, got %v", date, match.Date)
	}
	if match.Competition != competition {
		t.Errorf("expected Competition %v, got %v", competition, match.Competition)
	}
	if match.Status != MatchStatusScheduled {
		t.Errorf("expected Status %v, got %v", MatchStatusScheduled, match.Status)
	}
	if match.Score != nil {
		t.Errorf("expected Score to be nil, got %v", match.Score)
	}
}

func TestUpdateScore(t *testing.T) {
	// Test case: Update match score
	match := NewMatch("match123", "Team A", "Team B", time.Now(), "Premier League")

	homeGoals := 2
	awayGoals := 1
	match.UpdateScore(homeGoals, awayGoals)

	if match.Score == nil {
		t.Errorf("expected Score to be non-nil")
	}
	if match.Score.HomeGoals != homeGoals {
		t.Errorf("expected HomeGoals %v, got %v", homeGoals, match.Score.HomeGoals)
	}
	if match.Score.AwayGoals != awayGoals {
		t.Errorf("expected AwayGoals %v, got %v", awayGoals, match.Score.AwayGoals)
	}
}

func TestIsFinished(t *testing.T) {
	match := NewMatch("match123", "Team A", "Team B", time.Now(), "Premier League")

	// Test case 1: Match not finished
	if match.IsFinished() {
		t.Errorf("expected match to not be finished")
	}

	// Test case 2: Match finished
	match.Status = MatchStatusFinished
	if !match.IsFinished() {
		t.Errorf("expected match to be finished")
	}
}

func TestIsLive(t *testing.T) {
	match := NewMatch("match123", "Team A", "Team B", time.Now(), "Premier League")

	// Test case 1: Match not live
	if match.IsLive() {
		t.Errorf("expected match to not be live")
	}

	// Test case 2: Match live
	match.Status = MatchStatusLive
	if !match.IsLive() {
		t.Errorf("expected match to be live")
	}
}
