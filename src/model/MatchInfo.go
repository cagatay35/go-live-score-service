package model

type MatchInfo struct {
	HomeTeam   string
	AwayTeam   string
	HomeGoals  int
	AwayGoals  int
	HomeScores []GoalScorer
	AwayScores []GoalScorer
	Scores     []GoalScorer
	State      string
}
