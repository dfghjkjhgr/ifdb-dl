package main

import (
	"encoding/json"
	"net/http"
	"fmt"
	"os"
	"bufio"
	"io"
	"net/url"
	"log"
	"strconv"
	"strings"
)

type SearchGame struct {
	Tuid string `json:"tuid"`
	Title string `json:"title"`
	Devsys string `json:"devsys"`
	StarRating float32 `json:"starRating"`
	NumRatings int `json:"numRatings"`
	
}

type SearchGamesList struct {
	Games []SearchGame `json:"games"`
}

type ViewGame struct {
	Ifdb Ifdb `json:"ifdb"`
}

type Ifdb struct {
	Downloads Downloads `json:"downloads"`
}

type Downloads struct {
	Links []Link `json:"links"`
}

type Link struct {
	Url string `json:"url"`
	Title string `json:"title"`
	Desc string `json:"desc"`
	Format string `json:"format"`
	IsGame bool `json:"isGame"`
}

// Filters a list based on whether f applied to a member of s is true or false. Some people can't
// go without their functional convienences... : (
func filter[T comparable](s []T, f func(T) bool) []T {
	var acc []T
	for i := 0; i < len(s); i++ {
		if f(s[i]) {
			acc = append(acc, s[i])
		}
	}
	return acc
}

// 
func validatedPrompt(prompt string, f func(string) (string, error)) string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	text := scanner.Text()
	for {
		if newPrompt, err := f(text); err != nil {
			fmt.Print(newPrompt)
			text = scanner.Text()
			continue
		}
		break
	}
	return text
}
func gameSearch(term string) SearchGamesList {
	query := "https://ifdb.org/search?json&" + url.Values{"searchfor": {term}}.Encode()
	res, err := http.Get(query)
	if err != nil {
		log.Fatal("Could not connect to wifi");
	}
	body, err := io.ReadAll(res.Body)
	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	// Parse the JSON
	var list SearchGamesList
	err = json.Unmarshal(body, &list)
	if err != nil {
		fmt.Println("error:", err)
	}
	return list
}
	
func download(link Link, path string) {
	res, _ := http.Get(link.Url)
	body, _ := io.ReadAll(res.Body)
	err := os.WriteFile(path, body, 0666)
	if err != nil {
		fmt.Println("Failed to download file: do you have enough storage?")
	}
}

func ynPrompt(prompt string) bool {
	fmt.Printf(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	for {
		text := strings.ToLower(scanner.Text())
		switch {
		case text == "y" || text == "yes":
			return true
		case text == "n" || text == "no":
			return false
		default:
			fmt.Printf("Please enter yes or no: ")
			scanner.Scan()
		}
	}
}

func searchPrompt(list SearchGamesList) (string, error) {
	switch {
	case len(list.Games) >= 10:
		fmt.Println("Too many options, only showing first 10 results:")
		fmt.Println("(Use N and P to page)");
	case len(list.Games) == 0:
		fmt.Println("No search results found.")
		return "", fmt.Errorf("No search results found.")
	default:
		fmt.Println("Search results: ")
	}
	pagePlace := 0
	max := pagePlace + 10
	for i := pagePlace; i < min(len(list.Games), max); i++ {
		fmt.Printf("(%v): %v (%v, %v stars, %v ratings)\n", i,
			 list.Games[i].Title,
			 list.Games[i].Devsys,
			 list.Games[i].StarRating,
			 list.Games[i].NumRatings)
	}
	var num int
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Which game do you want to get? ")
		scanner.Scan()
		text := scanner.Text()
		var err error
		num, err = strconv.Atoi(text)
		if err != nil {
			if len(list.Games) >= 10 && strings.ToLower(text) == "n" {
				pagePlace += 10
				max = pagePlace + 10
				for i := pagePlace; i < min(len(list.Games), max); i++ {
					fmt.Printf("(%v): %v (%v, %v stars, %v ratings)\n", i,
						list.Games[i].Title,
						list.Games[i].Devsys,
						list.Games[i].StarRating,
						list.Games[i].NumRatings)
				}
				continue
			}
			if len(list.Games) >= 10 && strings.ToLower(text) == "p"   {
				pagePlace -= 10
				max = pagePlace + 10
				for i := pagePlace; i < min(len(list.Games), max); i++ {
					fmt.Printf("(%v): %v (%v, %v stars, %v ratings)\n", i,
						list.Games[i].Title,
						list.Games[i].Devsys,
						list.Games[i].StarRating,
						list.Games[i].NumRatings)
				}
				continue
			}
			if num < 0 || num >= len(list.Games) {
				fmt.Println("Please enter a number within an appropriate range.");
				continue
			} 
		}
		break
	}
	return list.Games[num].Tuid, nil
}

func main() {
	args := os.Args[1:]

	firstOption := ""

	if len(args) != 0 {
		firstOption = args[0]
	}

	switch firstOption {
	case "-h", "--help", "help":
		fmt.Println(`ifdb-dl: download your favorite IFDB games through the terminal!
This is an interactive program, call the executable without arguments to begin.
IFDB search syntax:
+word makes word mandatory - only items that contain this word will be listed.

-word makes word prohibited - only items that don't contain this word will be listed.

"phrase" searches for the exact phrase within the quotes: all of the words have to be matched in the exact order given.

author:name lists games by the named author or authors.

tag:tag name searches for games containing the given tag text. (What's a tag?)
Go to the IFDB tag list

series:name lists only games with the given series name.
Show all series names appearing in game listings

genre:genre name only shows games with the given genre. If a game's listing has multiple genres, it will match as long as genre name is found within the listing. For example, genre:western will match a game listed with genre "Science Fiction/Western/Romance."
Show all genres used in game listings

rating:low-high lists games with average ratings in the given range (inclusive). For example, rating:2.5-3.5 lists games rated from 2½ to 3½ stars. Leave out an endpoint for an open-ended search: rating:3- lists games with ratings 3 stars and above; rating:-2 lists games rated 2 stars and below.

#ratings:low-high lists games with a total number of ratings in the given range. For example, #ratings:3- lists games with three or more ratings.

ratingdev:low-high lists games with a ratings standard deviation in the given range.

#reviews:low-high lists games with a total number of member reviews in the given range. (This doesn't count editorial reviews.)

forgiveness:rating only shows games with the given "forgiveness" rating (on the Zarfian scale: Merciful, Polite, Tough, Nasty, Cruel - more information).

published:year-year only shows games with publication dates in the given range. For example, published:1990-2000 shows games published from 1990 to 2000. published:1990 lists only games published in 1990. published:1990- lists games published in 1990 or later, and published:-2000 lists games published in 2000 or earlier. You can also search for games published within the last few days. published:30d- searches for games published within the last 30 days. published:90d-30d searches for games published within the last 90 days, but more than 30 days ago.

added:year-year only shows games with listings added to the database on dates in the given range. For example, added:2007-2020 shows games added from 2007 to 2020. added:2007 shows games added in 2007. added:2007- shows games added in 2007 or later, and added:-2020 shows games added in 2020 or earlier. You can also search for games added within the last few days. added:30d- searches for games added within the last 30 days. added:90d-30d searches for games added within the last 90 days, but more than 30 days ago.

language:code lists games written in the given spoken language. You can use the English name of the language, or a two- or three-letter ISO-639 code ("en" for English, "fr" for French, etc).

system:name lists only games written with the given authoring system (TADS, Inform, Hugo, etc).

format:name lists only games with downloadable files available for the given format. To search for multiple system versions, use * as a wildcard: format:tads * searches for all TADS versions. Use an operating system name to search for native executables for that system.

downloadable:yes|no lists games that are/are not downloadable. A downloadable game is one that has at least one story file or application download link.

playtime:minimum-maximum lists games with an estimated play time in the given range. After each number, use h for hours or m for minutes. For example, playtime:2h15m-3h shows games with an estimated play time of anywhere from 2 hours and 15 minutes to 3 hours. Hours may include decimals (for example, 3.5h). playtime:1.5h- lists games with an estimated play time of at least 1 and a half hours. playtime:-45m lists games with an estimated play time of 45 minutes or less. playtime:1h searches for games with an estimated play time of 1 hour. playtime: with no text after it searches for games with no estimated play time.

bafs:id searches for the game with the given Baf's Guide ID.

ifid:xxx searches for a game with the given IFID.

tuid:xxx searches for a game with the given TUID.

authorid:id lists games by the author with the given id.

competitionid:id lists games in a competition with the given id. `);
	default:
		// Get API data from server
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("WELCOME TO IFDB-DL! Download your favorite IF games here! Type in your search or Ctrl-C to exit.")
		gameTUID, err := searchPrompt(gameSearch(Scanner.Scan().Text()));
		if err != nil {
			// if the program ends, that means that we didn't find any results...
			return;
		}
		query2 := "https://ifdb.org/viewgame?json&" + url.Values{"id": {gameTUID}}.Encode()
		res2, err := http.Get(query2)	
		if err != nil {
			log.Fatal("Something went wrong with getting the game.");
		}
		
		var game ViewGame
		body2, err := io.ReadAll(res2.Body)
		err = json.Unmarshal(body2, &game)
		if err != nil {
			fmt.Println("error:", err)
		}
		
		downloads := filter(game.Ifdb.Downloads.Links, func (l Link) bool {
			return l.IsGame
		})
		
		switch {
		case len(downloads) == 0:
			fmt.Println("No download links found... :(")
		case len(downloads) == 1:
			fmt.Printf("One download link found. (%v, %v)\n", downloads[0].Title, downloads[0].Format)
			if ynPrompt("Download? (y/n)") == true {
				scanner.Scan()
				download(downloads[0], scanner.Text())
			}
		default:
			var number int
			fmt.Println("There are multiple downloads avaliable:");
			for i := 0; i < len(downloads); i++ {
				fmt.Printf("(%v): %v (%v)\n", i, downloads[i].Title, downloads[i].Format)
			}
			for {	
				fmt.Print("Which one do you wish to download?: ")
				_, err = fmt.Scanf("%d", &number)
				if err != nil || (number < 0 || number >= len(downloads)) {
					fmt.Println("Please enter a number within an appropriate range.");
					continue
				}
				break
			}

			fmt.Printf("File path? (default: '/mnt/us/extensions/Gargoyle/games/%v", downloads[number].Title)
			scanner.Scan()
			download(downloads[number], scanner.Text())
		}		
	}
}
