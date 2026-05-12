package handlers

import (
	"crypto/rand"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WeatherResult struct {
	Season      string `json:"season"`
	Temperature string `json:"temperature"`
	Sky         string `json:"sky"`
	Precipitation string `json:"precipitation"`
	Wind        string `json:"wind"`
	Special     string `json:"special"`
	Description string `json:"description"`
}

var seasons = []string{"Spring", "Summer", "Autumn", "Winter"}

var temperatures = map[string][]string{
	"Spring": {"Cool (40-60°F)", "Mild (50-70°F)", "Warm (60-80°F)"},
	"Summer": {"Warm (60-80°F)", "Hot (70-90°F)", "Scorching (85-105°F)"},
	"Autumn": {"Cool (40-60°F)", "Crisp (35-55°F)", "Mild (50-70°F)"},
	"Winter": {"Freezing (10-30°F)", "Cold (20-40°F)", "Bitter Cold (-10-20°F)"},
}

var skies = []string{"Clear", "Partly Cloudy", "Overcast", "Foggy", "Hazy"}
var precipitations = []string{"None", "Light Drizzle", "Rain", "Heavy Rain", "Thunderstorm", "Snow", "Heavy Snow", "Sleet", "Hail"}
var winds = []string{"Calm", "Light Breeze", "Moderate Wind", "Strong Wind", "Gale", "Storm"}
var specialties = []string{
	"None",
	"Sudden temperature drop",
	"Unnatural fog",
	"Blood-red moon visible",
	"Distant storm on the horizon",
	"Strange lights in the sky",
	"Overwhelming floral scent",
	"Ash falling from the sky",
	"Swarm of insects or bats",
	"Eerie silence",
	"Shooting star",
	"Rainbow after rain",
	"Thick hazy smog",
	"Geomagnetic storm - compasses spin",
	"Miasma from the ground",
}

func randChoice[T any](arr []T) T {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(arr))))
	return arr[n.Int64()]
}

func HandleGenerateWeather(c *gin.Context) {
	season := c.DefaultQuery("season", "")
	if season == "" {
		season = randChoice(seasons)
	}
	tempOptions, ok := temperatures[season]
	if !ok {
		tempOptions = temperatures["Spring"]
	}
	temp := randChoice(tempOptions)
	sky := randChoice(skies)
	precip := randChoice(precipitations)
	wind := randChoice(winds)
	special := randChoice(specialties)

	// Adjust based on season
	if season == "Winter" && precip == "Rain" {
		precip = "Snow"
	}
	if season == "Summer" && precip == "Snow" {
		precip = "None"
	}
	if season == "Spring" && precip == "Heavy Snow" {
		precip = "Rain"
	}

	desc := ""

	if precip == "Thunderstorm" && wind == "Storm" {
		desc = "A violent thunderstorm rages! Lightning splits the sky as gale-force winds howl."
	} else if precip == "Thunderstorm" {
		desc = "Thunder rumbles overhead as rain pours down."
	} else if precip == "Heavy Rain" && wind == "Gale" {
		desc = "Driving rain lashes sideways in the strong wind. Visibility is poor."
	} else if precip == "Heavy Snow" {
		desc = "Snow falls heavily, blanketing the landscape."
	} else if precip == "Snow" {
		desc = "Gentle snowflakes drift down from the grey sky."
	} else if sky == "Foggy" {
		desc = "Thick fog obscures vision beyond a few feet."
	} else if sky == "Clear" && season == "Summer" {
		desc = "The sun beats down from a cloudless blue sky."
	} else if sky == "Clear" && season == "Winter" {
		desc = "The air is biting cold under a clear, bright sky."
	} else if precip == "None" && wind == "Calm" {
		desc = "The air is still and calm. A quiet day."
	} else {
		desc = "It is a typical day for this season."
	}

	if special != "None" {
		desc += " " + special + "."
	}

	c.JSON(http.StatusOK, WeatherResult{
		Season:        season,
		Temperature:   temp,
		Sky:           sky,
		Precipitation: precip,
		Wind:          wind,
		Special:       special,
		Description:   desc,
	})
}
