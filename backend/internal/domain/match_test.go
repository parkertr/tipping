package domain_test

import (
	"testing"
	"time"

	"github.com/parkertr/tipping/internal/domain"
)

func TestNewMatch(t *testing.T) {
	t.Parallel()

	id := "match1"
	homeTeam := "Arsenal"
	awayTeam := "Chelsea"
	date := time.Now()
	competition := "Premier League"

	match := domain.NewMatch(id, homeTeam, awayTeam, date, competition)

	if match.ID != id {
		t.Errorf("Expected ID %s, got %s", id, match.ID)
	}

	if match.HomeTeam != homeTeam {
		t.Errorf("Expected HomeTeam %s, got %s", homeTeam, match.HomeTeam)
	}

	if match.AwayTeam != awayTeam {
		t.Errorf("Expected AwayTeam %s, got %s", awayTeam, match.AwayTeam)
	}

	if match.Date != date {
		t.Errorf("Expected Date %v, got %v", date, match.Date)
	}

	if match.Competition != competition {
		t.Errorf("Expected Competition %s, got %s", competition, match.Competition)
	}

	if match.Status != domain.MatchStatusScheduled {
		t.Errorf("Expected Status %s, got %s", domain.MatchStatusScheduled, match.Status)
	}

	if match.Score == nil {
		t.Errorf("Expected Score to be set")
	}

	if match.Score.HomeGoals != 0 || match.Score.AwayGoals != 0 {
		t.Errorf("Expected initial Score to be 0-0")
	}
}

func TestUpdateScore(t *testing.T) {
	t.Parallel()

	match := domain.NewMatch("match1", "Arsenal", "Chelsea", time.Now(), "Premier League")
	homeGoals := 2
	awayGoals := 1

	match.UpdateScore(homeGoals, awayGoals)

	if match.Score == nil {
		t.Errorf("Expected Score to be set")
	}

	if match.Score.HomeGoals != homeGoals {
		t.Errorf("Expected HomeGoals %d, got %d", homeGoals, match.Score.HomeGoals)
	}

	if match.Score.AwayGoals != awayGoals {
		t.Errorf("Expected AwayGoals %d, got %d", awayGoals, match.Score.AwayGoals)
	}
}

func TestIsFinished(t *testing.T) {
	t.Parallel()

	match := domain.NewMatch("match1", "Arsenal", "Chelsea", time.Now(), "Premier League")
	match.Status = domain.MatchStatusFinished

	if !match.IsFinished() {
		t.Errorf("Expected IsFinished to be true")
	}

	match.Status = domain.MatchStatusLive
	if match.IsFinished() {
		t.Errorf("Expected IsFinished to be false")
	}
}

func TestIsLive(t *testing.T) {
	t.Parallel()

	match := domain.NewMatch("match1", "Arsenal", "Chelsea", time.Now(), "Premier League")
	match.Status = domain.MatchStatusLive

	if !match.IsLive() {
		t.Errorf("Expected IsLive to be true")
	}

	match.Status = domain.MatchStatusFinished
	if match.IsLive() {
		t.Errorf("Expected IsLive to be false")
	}
}
