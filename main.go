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

func filter[T comparable](s []T, f func(T) bool) []T {
	var acc []T
	for i := 0; i < len(s); i++ {
		if f(s[i]) {
			acc = append(acc, s[i])
		}
	}
	return acc
}
	
// func validatedPrompt(prompt string, f func(string) (string, error)) {
	
// }
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
	
func download(link Link) {
	res, _ := http.Get(link.Url)
	body, _ := io.ReadAll(res.Body)
	_ = os.WriteFile(link.Title, body, 0666)
}

func searchPrompt(list SearchGamesList) (string, error) {
	// TODO: Implement paging
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
		fmt.Println("ifdb-dl: download your favorite IFDB games through the terminal!\nThis is an interactive program, call the executable without arguments to begin.");
	default:
		// Get API data from server
		gameTUID, err := searchPrompt(gameSearch(firstOption));
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
			fmt.Println("One download link found. Downloading...")
			download(downloads[0])
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
			download(downloads[number])
		}		
	}
}
