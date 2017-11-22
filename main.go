package main

import (
	"strconv"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"io"
	"regexp"
	"sort"
	//"time"
)

// types have to be strings because results contain non integer values
type game struct {
	ID			string 		`json:"id"`
	HomeTeamID 	string 		`json:"homeId"`
	AwayTeamID 	string 		`json:"awayId"`
	AwayName    string      `json:"awayName"`
	HomeName    string      `json:"homeName"`
	HomeScore 	string 		`json:"homeScore"`
	AwayScore 	string 		`json:"awayScore"`
	HomeClubID 	string 		`json:"homeClub"`
	AwayClubID 	string 		`json:"awayClub"`
	HomeDivisionID string	`json:"homeDivision"`
	AwayDivisionID string	`json:"awayDivision"`
	//StartDate 	time.Time 	`json:"startDate"`
	IsPlayed	string		`json:"isPlayed"`
}

const marinFC = "3139"
// 4000300 U14
// 4000301 U12
// 4000302 U13


/*
{"id":"4154357",
	"gameDivision":null,
	"startDate":"2017-04-28 00:00:00",
	"endDate":"2017-04-28 01:00:00",
	"homeId":"3994742",
	"awayId":"4000781",
	"locationId":"75640",
	"scheduleId":"278533",
	"homeName":"Ballistic United SC U-12",
	"homeScore":"0",
	"awayName":"Placer United SC U-12",
	"awayScore":"0",
	"isPlayed":null,
	"isCancelled":false,
	"locationName":"Val Vista - Field 1",
	"locationAddress":"7350 Johnson Dr.",
	"locationAddress2":null,
	"locationCity":"Pleasanton",
	"locationState":"CA",
	"locationZip":"94588",
	"locationParentName":null,
	"jerseyColorsAreEnabled":"0",
	"jerseyHaveBottom":"0",
	"jerseyHavePattern":"0",
	"homeTopColor1":null,
	"homeTopColor2":null,
	"homeBottomColor":null,
	"homeAgeGroup":"449475",
	"homeOutsideClubName":null,
	"homeDivision":"130811",
	"homeClub":"3129",
	"awayTopColor1":null,
	"awayTopColor2":null,
	"awayBottomColor":null,
	"awayAgeGroup":"449475",
	"awayOutsideClubName":null,
	"frontEndScores":0,
	"hideGameReports":1,
	"awayDivision":"130811",
	"awayClub":"3143"
*/

type club struct {
	ID int `json:"id"`
	Name string `json:"name"`
}

type division struct {
	ID int `json:"id"`
	Name string `json:"name"`
}

type standing struct {
	DivisionID string
	ClubID string
	TeamID string
	Name string
	Wins int
	Ties int
	Losses int
	Points int
	GoalsFor int
	GoalsAgainst int
	Games []game
}

// Sort Interface
type byPoints []standing

func (a byPoints) Len() int           { return len(a) }
func (a byPoints) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPoints) Less(i, j int) bool { 
	// Note: Just compare points. Don't worry about other tie breakers
    if a[i].Points > a[j].Points {
       return true
	}
	
	return false
}

func getScriptCode(r io.Reader) string {
	z := html.NewTokenizer(r)
	
	for {
		tt := z.Next()
	
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			panic("Unable to find script section we're looking for")
		case tt == html.StartTagToken:
			t := z.Token()
	
			isScript := t.Data == "script"
			if isScript {

				// Determine if this is the one we care about
				inner := z.Next()
				if inner == html.TextToken {
					text := (string)(z.Text())
					match, _ := regexp.MatchString("games", text)
					if match  {
						return text
					}
				}
			}
		}
	}	
}

func getJSON(script string, tag string) string {

	pattern := tag + "\"\\]=(?P<tag>\\[{[^;]*}\\])"
	scriptTag := regexp.MustCompile(pattern)

	matches := scriptTag.FindStringSubmatch(script)
	if len(matches) > 0 {
		return matches[1]
	}

	return ""
}

func parseJSONs(gamesJSON string, clubsJSON string, divisionsJSON string) ([]game, []club, [] division) {

	games := []game{}
	if err := json.Unmarshal([]byte(gamesJSON), &games); err != nil {
		fmt.Println("Error parsing Games JSON")
		panic(err)
	}

	divisions := []division{}
	if err := json.Unmarshal([]byte(divisionsJSON), &divisions); err != nil {
		fmt.Println("Error parsing Division JSON")
		panic(err)
	}

	clubs := []club{}
	if err := json.Unmarshal([]byte(clubsJSON), &clubs); err != nil {
		fmt.Println("Error parsing Club JSON")
		panic(err)
	}

	return games, clubs, divisions
}

func getStanding(standings map[string]*standing, g game, isAway bool) *standing {

	var ok bool
	var val *standing
	var teamID string
	if isAway {
		teamID = g.AwayTeamID
	} else {
		teamID = g.HomeTeamID
	}

	if val, ok = standings[teamID]; !ok {
		val = &standing{}
		val.TeamID = teamID
		if isAway {
			val.ClubID = g.AwayClubID
			val.DivisionID = g.AwayDivisionID
			val.Name = g.AwayName
		} else {
			val.ClubID = g.HomeClubID
			val.DivisionID = g.HomeDivisionID
			val.Name = g.HomeName
		}
		standings[teamID] = val
	}

	return val
}

func updateStandingPoints(s *standing) {
	s.Points = s.Wins * 3 + s.Ties
}

func createStandings(divisions []division, clubs []club, games []game) map[string]*standing {
	
	standings := map[string]*standing{}
	for _, g := range games {
		// Ignore games that haven't been played
		if g.IsPlayed == "" {
			continue
		}
		// initialize the home and away standing if it doesn't already exist
		awayStanding := getStanding(standings, g, true)
		homeStanding := getStanding(standings, g, false)

		// Associate game with standing
		awayStanding.Games = append(awayStanding.Games, g)
		homeStanding.Games = append(homeStanding.Games, g)
		
		// Update wins / losses / ties
		if g.AwayScore > g.HomeScore {
			awayStanding.Wins++
			homeStanding.Losses++
		} else if g.AwayScore < g.HomeScore {
			homeStanding.Wins++
			awayStanding.Losses++
		} else {
			awayStanding.Ties++
			homeStanding.Ties++
		}

		// Update Goals for and against
		awayScore, _ := strconv.Atoi(g.AwayScore)
		homeScore, _ := strconv.Atoi(g.HomeScore)
		awayStanding.GoalsFor += awayScore
		awayStanding.GoalsAgainst += homeScore
		homeStanding.GoalsFor += homeScore
		homeStanding.GoalsAgainst += awayScore

		// Calculate Points
		updateStandingPoints(homeStanding)
		updateStandingPoints(awayStanding)
	}

	return standings
}

func showTeamStanding(standings map[string]*standing, teamID string) {
	teamStanding := standings[teamID]
	divisionID := teamStanding.DivisionID

	// Grab all teams in the division
	var divisionStandings []standing
	for _, s := range standings {
		if s.DivisionID == divisionID {
			divisionStandings = append(divisionStandings, *s)
		}
	}

	// Sort by points
	sort.Sort(byPoints(divisionStandings))

	// Show the standings
	fmt.Printf("%50s\tWins\tLosses\tTies\tPoints\tGF\tGA\tDiff\n", "Team")
	for _, s := range divisionStandings {
		fmt.Printf("%50s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n", s.Name, s.Wins, s.Losses, s.Ties, s.Points, s.GoalsFor, s.GoalsAgainst, s.GoalsFor - s.GoalsAgainst)
	}
	fmt.Printf("\n")
	
	// TODO: Sort the games by date played
	// Show games for the requested team
	fmt.Printf("%s Games\n", teamStanding.Name)
	fmt.Printf("%50s vs %50s\tScore\n", "Home", "Away")
	for _, g := range teamStanding.Games {
		fmt.Printf("%50s vs %50s\t%s - %s\n", g.HomeName, g.AwayName, g.HomeScore, g.AwayScore)
	}
	fmt.Printf("\n\n")
	
}

func main() {

	// Make HTTP Request
	resp, _ := http.Get("http://www.ussoccerda.com/sam/standings/regevent/index.php?containerId=MzgzNDMwMA%3D%3D&partialGames=0")

	// Find the script tag that contains the json code we want to parse
	scriptCode := getScriptCode(resp.Body)
	resp.Body.Close()
	
	///////////////////////////////////////////////////////
	// Get the JSON data from the script tag and parse it
	gamesJSON :=getJSON(scriptCode, "games")
	divisionsJSON := getJSON(scriptCode, "divisions")
	clubsJSON := getJSON(scriptCode, "clubs")
	games, clubs, divisions := parseJSONs(gamesJSON, clubsJSON, divisionsJSON)

	// create standings data
	standings := createStandings(divisions, clubs, games)
	showTeamStanding(standings, "4000301")
	showTeamStanding(standings, "4000302")
	showTeamStanding(standings, "4000300")
}