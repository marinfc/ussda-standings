package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"io"
	"regexp"
)

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
			return ""
		case tt == html.StartTagToken:
			t := z.Token()
	
			isScript := t.Data == "script"
			if isScript {
				fmt.Println("We found a script!")

				// Determine if this is the one we care about
				inner := z.Next()
				if inner == html.TextToken {
					text := (string)(z.Text())
					match, _ := regexp.MatchString("games", text)
					if match  {
						fmt.Println("Found Script we're looking for!")
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

func main() {

	/* Make HTTP Request */
	resp, _ := http.Get("http://www.ussoccerda.com/sam/standings/regevent/index.php?containerId=MzgzNDMwMA%3D%3D&partialGames=0")

	// Find the script tag that contains the json code we want to parse
	scriptCode := getScriptCode(resp.Body)
	resp.Body.Close()
	
	// Get the JSON data from the script tag and parse it
	//gamesJSON :=getJSON(scriptCode, "games")

	// division JSON
	divisionsJSON := getJSON(scriptCode, "divisions")
	divisions := []division{}
	if err := json.Unmarshal([]byte(divisionsJSON), &divisions); err != nil {
		fmt.Println("Error parsing Division JSON")
		panic(err)
	}
	fmt.Printf("Division[0]: %s\nDivision[1]: %s\n", divisions[0].Name, divisions[1].Name)

	// club JSON
	clubsJSON := getJSON(scriptCode, "clubs")
	clubs := []club{}
	if err := json.Unmarshal([]byte(clubsJSON), &clubs); err != nil {
		fmt.Println("Error parsing Club JSON")
		panic(err)
	}
	fmt.Printf("Club[0]: %s\nClub[1]: %s\n", clubs[0].Name, clubs[1].Name)

	//fmt.Printf("Games: %s\n", gamesJSON)

	fmt.Println("Done")
}