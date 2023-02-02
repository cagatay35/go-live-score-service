package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go-live-score-service/src/model"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type LiveScoreApp struct {
	websocketUpgrader websocket.Upgrader
	ginEngine         *gin.Engine
	connections       []*websocket.Conn
	wg                sync.WaitGroup
	liveScoreMsgChan  chan *model.MatchInfo
}

func main() {
	var app = &LiveScoreApp{}
	app.wg = sync.WaitGroup{}
	app.wg.Add(1)
	go app.startWebSocket()
	app.wg.Wait()
	app.liveScoreMsgChan = make(chan *model.MatchInfo)
	go app.startBroadcastService()
	matchInfo := &model.MatchInfo{
		HomeTeam:  "Arsenal",
		AwayTeam:  "Manchester United",
		HomeGoals: 0,
		AwayGoals: 0,
	}

	go func() {
		time.Sleep(time.Second * 10)
		go app.simulateMatch(matchInfo)
		app.liveScoreMsgChan <- matchInfo
	}()

	app.startWebServer()

}

func (app *LiveScoreApp) simulateMatch(matchInfo *model.MatchInfo) {
	var manUnitedRoster = []string{
		"David de Gea",
		"Victor Lindelöf",
		"Harry Maguire",
		"Luke Shaw",
		"Scott McTominay",
		"Bruno Fernandes",
		"Paul Pogba",
		"Marcus Rashford",
		"Anthony Martial",
		"Mason Greenwood",
	}

	var arsenalRoster = []string{
		"Bernd Leno",
		"Héctor Bellerín",
		"Rob Holding",
		"Kieran Tierney",
		"Thomas Partey",
		"Granit Xhaka",
		"Bukayo Saka",
		"Emile Smith Rowe",
		"Nicolas Pépé",
		"Pierre Aubameyang",
	}

	var min = 0

	for min = 0; min <= 90; min++ {

		var goalPossibility = generateRandomNumber(20, 1)

		if goalPossibility == 5 {

			var whichTeamScored = generateRandomNumber(3, 1)

			var scorer string
			if whichTeamScored == 1 {
				// home goal scores
				scorer = arsenalRoster[rand.Int()%len(arsenalRoster)]
				matchInfo.HomeGoals = matchInfo.HomeGoals + 1
				matchInfo.HomeScores = append(matchInfo.HomeScores, model.GoalScorer{
					Name:   scorer,
					Team:   "Arsenal",
					Minute: min,
				})

				matchInfo.Scores = append(matchInfo.Scores, model.GoalScorer{
					Name:   scorer,
					Team:   "home",
					Minute: min,
				})

			} else {
				// away goal scores
				scorer = manUnitedRoster[rand.Int()%len(manUnitedRoster)]
				matchInfo.AwayGoals = matchInfo.AwayGoals + 1
				matchInfo.AwayScores = append(matchInfo.AwayScores, model.GoalScorer{
					Name:   scorer,
					Team:   "Manchester United",
					Minute: min,
				})

				matchInfo.Scores = append(matchInfo.Scores, model.GoalScorer{
					Name:   scorer,
					Team:   "away",
					Minute: min,
				})
			}

			app.liveScoreMsgChan <- matchInfo

		}

		if min == 45 {
			matchInfo.State = "end of first half"
			app.liveScoreMsgChan <- matchInfo
		}
		if min == 90 {
			matchInfo.State = "end"
			app.liveScoreMsgChan <- matchInfo
		}

		time.Sleep(time.Second)
		fmt.Println(fmt.Printf("minute : %s\n", strconv.Itoa(min)))
	}

}

func (app *LiveScoreApp) startBroadcastService() {

	fmt.Println("waiting for new messages")

	for msg := range app.liveScoreMsgChan {

		fmt.Println("new message received")

		for _, conn := range app.connections {
			if err := conn.WriteJSON(msg); err != nil {
				fmt.Println("could not send initial result: %v\n", err)
				return
			}
			fmt.Println("message sent")
		}

	}

}
func (app *LiveScoreApp) startWebSocket() {
	app.websocketUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	app.wg.Done()
	fmt.Println("Web Socket starting ...")
}

func (app *LiveScoreApp) startWebServer() {
	// start web server
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(gin.Recovery())

	engine.LoadHTMLGlob("resources/**/*.html")

	engine.GET("/ws", func(c *gin.Context) {
		conn, err := app.websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
		fmt.Println("new client connected")
		if err != nil {
			c.String(http.StatusBadRequest, "Could not upgrade connection: %v", err)
			return
		}
		// add new connection to slice of connections
		app.connections = append(app.connections, conn)
	})

	engine.GET("/healthcheck", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"result": "UP"})
	})

	engine.GET("/home", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", nil)
	})

	app.ginEngine = engine

	fmt.Println("Web Server starting ...")
	engine.Run(":8080")
}

func generateRandomNumber(max int, min int) int {
	rand.Seed(time.Now().UnixNano())
	randId := rand.Intn(max-min) + 1
	return randId
}
