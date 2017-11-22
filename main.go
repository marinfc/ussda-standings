package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"io"
	"regexp"
)

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

	pattern := tag + "\"\\]=(?P<tag>\\[{.*}\\])"
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

	scriptCode := getScriptCode(resp.Body)

	//gamesJSON :=getJSON(scriptCode, "games")
	divisionsJSON := getJSON(scriptCode, "divisions")
	clubsJSON := getJSON(scriptCode, "clubs")

	resp.Body.Close()

	//fmt.Printf("Games: %s\n", gamesJSON)
	fmt.Printf("Divisions: %s\n", divisionsJSON)
	fmt.Printf("Clubs: %s\n", clubsJSON)

	fmt.Println("Done")
}