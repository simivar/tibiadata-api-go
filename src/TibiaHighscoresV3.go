package main

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

// Child of Highscores
type Highscore struct {
	Rank     int    `json:"rank"`            // Rank column
	Name     string `json:"name"`            // Name column
	Vocation string `json:"vocation"`        // Vocation column
	World    string `json:"world"`           // World column
	Level    int    `json:"level"`           // Level column
	Value    int    `json:"value"`           // Points/SkillLevel column
	Title    string `json:"title,omitempty"` // Title column (when category: loyalty)
}

// Child of JSONData
type Highscores struct {
	World         string      `json:"world"`
	Category      string      `json:"category"`
	Vocation      string      `json:"vocation"`
	HighscoreAge  int         `json:"highscore_age"`
	HighscoreList []Highscore `json:"highscore_list"`
}

//
// The base includes two levels: Highscores and Information
type HighscoresResponse struct {
	Highscores  Highscores  `json:"highscores"`
	Information Information `json:"information"`
}

var (
	HighscoresAgeRegex = regexp.MustCompile(`.*<div class="Text">Highscores.*Last Update: ([0-9]+) minutes ago.*`)
	SevenColumnRegex   = regexp.MustCompile(`<td>.*<\/td><td.*">(.*)<\/a><\/td><td.*>(.*)<\/td><td.*>(.*)<\/td><td>(.*)<\/td><td.*>(.*)<\/td><td.*>(.*)<\/td>`)
	SixColumnRegex     = regexp.MustCompile(`<td>.*<\/td><td.*">(.*)<\/a><\/td><td.*">(.*)<\/td><td>(.*)<\/td><td.*>(.*)<\/td><td.*>(.*)<\/td>`)
)

// TibiaHighscoresV3 func
func TibiaHighscoresV3(c *gin.Context) {
	// getting params from URL
	world := c.Param("world")
	category := c.Param("category")
	vocation := c.Param("vocation")

	// maybe return error on faulty vocation value?!

	// Adding fix for First letter to be upper and rest lower
	if strings.EqualFold(world, "all") {
		world = ""
	} else {
		world = TibiadataStringWorldFormatToTitleV3(world)
	}

	highscoreCategory := HighscoreCategoryFromString(category)

	// Sanitize of vocation input
	vocationName, vocationid := TibiaDataVocationValidator(vocation)

	// Getting data with TibiadataHTMLDataCollectorV3
	TibiadataRequest.URL = "https://www.tibia.com/community/?subtopic=highscores&world=" + TibiadataQueryEscapeStringV3(world) + "&category=" + strconv.Itoa(int(highscoreCategory)) + "&profession=" + TibiadataQueryEscapeStringV3(vocationid) + "&currentpage=400000000000000"
	BoxContentHTML, err := TibiadataHTMLDataCollectorV3(TibiadataRequest)

	// return error (e.g. for maintenance mode)
	if err != nil {
		TibiaDataAPIHandleOtherResponse(c, http.StatusBadGateway, "TibiaHighscoresV3", gin.H{"error": err.Error()})
		return
	}

	jsonData := TibiaHighscoresV3Impl(world, highscoreCategory, vocationName, BoxContentHTML)

	// return jsonData
	TibiaDataAPIHandleSuccessResponse(c, "TibiaHighscoresV3", jsonData)
}

func TibiaHighscoresV3Impl(world string, category HighscoreCategory, vocationName string, BoxContentHTML string) HighscoresResponse {
	// Loading HTML data into ReaderHTML for goquery with NewReader
	ReaderHTML, err := goquery.NewDocumentFromReader(strings.NewReader(BoxContentHTML))
	if err != nil {
		log.Fatal(err)
	}

	// Creating empty HighscoreData var
	var (
		HighscoreData                                                           []Highscore
		HighscoreDataVocation, HighscoreDataWorld, HighscoreDataTitle           string
		HighscoreDataRank, HighscoreDataLevel, HighscoreDataValue, HighscoreAge int
	)

	// getting age of data
	subma1 := HighscoresAgeRegex.FindAllStringSubmatch(string(BoxContentHTML), 1)
	HighscoreAge = TibiadataStringToIntegerV3(subma1[0][1])

	// Running query over each div
	ReaderHTML.Find(".TableContent tr").First().NextAll().Each(func(index int, s *goquery.Selection) {

		// Storing HTML into CreatureDivHTML
		HighscoreDivHTML, err := s.Html()
		if err != nil {
			log.Fatal(err)
		}

		// Regex the data table..
		var subma1 [][]string

		/*
			Tibia highscore table columns
			Achievment	=>	Rank		Name	Vocation	World		Level	Points
			Axe			=>	Rank		Name	Vocation	World		Level	Skill Level
			Charm		=>	Rank		Name	Vocation	World		Level	Points
			Club		=>	Rank		Name	Vocation	World		Level	Skill Level
			Distance	=>	Rank		Name	Vocation	World		Level	Skill Level
			Drome		=>	Rank		Name	Vocation	World		Level	Score
			Exp			=>	Rank		Name	Vocation	World		Level	Points
			Fishing		=>	Rank		Name	Vocation	World		Level	Skill Level
			Fist		=>	Rank		Name	Vocation	World		Level	Skill Level
			Goshnar		=>	Rank		Name	Vocation	World		Level	Points
			Loyalty		=>	Rank		Name	Title		Vocation	World	Level			Points
			Magic lvl	=>	Rank		Name	Vocation	World		Level	Skill Level
			Shield		=>	Rank		Name	Vocation	World		Level	Skill Level
			Sword		=>	Rank		Name	Vocation	World		Level	Skill Level
		*/

		if category == loyaltypoints {
			subma1 = SevenColumnRegex.FindAllStringSubmatch(HighscoreDivHTML, -1)
		} else {
			subma1 = SixColumnRegex.FindAllStringSubmatch(HighscoreDivHTML, -1)
		}

		if len(subma1) > 0 {

			// Debugging of what is in which column
			if TibiadataDebug {
				log.Println("1 -> " + subma1[0][1])
				log.Println("2 -> " + subma1[0][2])
				log.Println("3 -> " + subma1[0][3])
				log.Println("4 -> " + subma1[0][4])
				log.Println("5 -> " + subma1[0][5])
				if category == loyaltypoints {
					log.Println("6 -> " + subma1[0][6])
				}
			}

			HighscoreDataRank++
			if category == loyaltypoints {
				HighscoreDataTitle = subma1[0][2]
				HighscoreDataVocation = subma1[0][3]
				HighscoreDataWorld = subma1[0][4]
				HighscoreDataLevel = TibiadataStringToIntegerV3(subma1[0][5])
				HighscoreDataValue = TibiadataStringToIntegerV3(subma1[0][6])
			} else {
				HighscoreDataVocation = subma1[0][2]
				HighscoreDataWorld = subma1[0][3]
				HighscoreDataLevel = TibiadataStringToIntegerV3(subma1[0][4])
				HighscoreDataValue = TibiadataStringToIntegerV3(subma1[0][5])
			}

			HighscoreData = append(HighscoreData, Highscore{
				Rank:     HighscoreDataRank,
				Name:     TibiaDataSanitizeEscapedString(subma1[0][1]),
				Vocation: HighscoreDataVocation,
				World:    HighscoreDataWorld,
				Level:    HighscoreDataLevel,
				Value:    HighscoreDataValue,
				Title:    HighscoreDataTitle,
			})

		}
	})

	// Printing the HighscoreData data to log
	if TibiadataDebug {
		log.Println(HighscoreData)
	}

	categoryString, _ := category.String()

	//
	// Build the data-blob
	return HighscoresResponse{
		Highscores{
			World:         strings.Title(strings.ToLower(world)),
			Category:      categoryString,
			Vocation:      vocationName,
			HighscoreAge:  HighscoreAge,
			HighscoreList: HighscoreData,
		},
		Information{
			APIVersion: TibiadataAPIversion,
			Timestamp:  TibiadataDatetimeV3(""),
		},
	}
}
