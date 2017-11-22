package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"io"
	"regexp"
)

// types have to be strings because results contain non integer values
type game struct {
	ID string `json:"id"`
	HomeID string `json:"homeId"`
	AwayID string `json:"awayId"`
	HomeScore string `json:"homeScore"`
	AwayScore string `json:"awayScore"`
}

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
	"awayScore":"0","isPlayed":null,
	"isCancelled":false,
	"locationName":"Val Vista - Field 1",
	"locationAddress":"7350 Johnson Dr.",
	"locationAddress2":null,"locationCity":"Pleasanton",
	"locationState":"CA","locationZip":"94588","locationParentName":null,"jerseyColorsAreEnabled":"0","jerseyHaveBottom":"0","jerseyHavePattern":"0","homeTopColor1":null,"homeTopColor2":null,"homeBottomColor":null,"homeAgeGroup":"449475","homeOutsideClubName":null,"homeDivision":"130811","homeClub":"3129","awayTopColor1":null,"awayTopColor2":null,"awayBottomColor":null,"awayAgeGroup":"449475","awayOutsideClubName":null,"frontEndScores":0,"hideGameReports":1,"awayDivision":"130811","awayClub":"3143"
*/
type club struct {
	ID int `json:"id"`
	Name string `json:"name"`
}

type division struct {
	ID int `json:"id"`
	Name string `json:"name"`
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

	// Print some test data
	fmt.Printf("Game[0]: %s\nGame[1]: %s\n", games[0].ID, games[1].ID)
	fmt.Printf("Division[0]: %s\nDivision[1]: %s\n", divisions[0].Name, divisions[1].Name)
	fmt.Printf("Club[0]: %s\nClub[1]: %s\n", clubs[0].Name, clubs[1].Name)

	fmt.Println("Done")
}