package handlers

import (
	"crypto/rand"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WeatherResult struct {
	Season        string `json:"season"`
	Biome         string `json:"biome"`
	Temperature   string `json:"temperature"`
	Sky           string `json:"sky"`
	Precipitation string `json:"precipitation"`
	Wind          string `json:"wind"`
	Special       string `json:"special"`
	Description   string `json:"description"`
}

var seasons = []string{"Spring", "Summer", "Autumn", "Winter"}

var biomes = []string{"temperate", "arctic", "desert", "forest", "mountain", "swamp", "coast", "underdark"}

// biomeTemperatures overrides the season temperature options for biomes whose
// climate differs from the temperate default.
var biomeTemperatures = map[string]map[string][]string{
	"arctic": {
		"Spring": {"Freezing (0-20°F)", "Cold (10-30°F)"},
		"Summer": {"Cold (20-40°F)", "Cool (30-50°F)"},
		"Autumn": {"Freezing (5-25°F)", "Cold (15-35°F)"},
		"Winter": {"Bitter Cold (-30-0°F)", "Freezing (-10-15°F)"},
	},
	"desert": {
		"Spring": {"Hot (80-100°F)", "Warm (70-90°F)"},
		"Summer": {"Scorching (95-120°F)", "Hot (85-105°F)"},
		"Autumn": {"Hot (75-95°F)", "Warm (65-85°F)"},
		"Winter": {"Mild (55-75°F)", "Cool (45-65°F)"},
	},
	"underdark": {
		"Spring": {"Cool (50-60°F)"},
		"Summer": {"Cool (50-60°F)"},
		"Autumn": {"Cool (50-60°F)"},
		"Winter": {"Cool (50-60°F)"},
	},
}

// biomePrecipitations overrides the precipitation pool for biomes that cannot
// produce the full temperate range (desert never snows, arctic mostly does).
var biomePrecipitations = map[string][]string{
	"arctic":    {"None", "Snow", "Heavy Snow", "Sleet", "Hail"},
	"desert":    {"None", "None", "None", "Light Drizzle", "Hail"},
	"underdark": {"None", "None", "Light Drizzle"},
	"coast":     {"None", "Light Drizzle", "Rain", "Heavy Rain", "Thunderstorm"},
	"swamp":     {"None", "Light Drizzle", "Rain", "Heavy Rain", "Thunderstorm"},
}

func validBiome(biome string) bool {
	for _, b := range biomes {
		if b == biome {
			return true
		}
	}
	return false
}

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
	biome := c.DefaultQuery("biome", "temperate")
	if !validBiome(biome) {
		biome = "temperate"
	}
	tempOptions, ok := temperatures[season]
	if !ok {
		tempOptions = temperatures["Spring"]
	}
	if bt, ok := biomeTemperatures[biome]; ok {
		if opts, ok := bt[season]; ok {
			tempOptions = opts
		}
	}
	precipPool := precipitations
	if bp, ok := biomePrecipitations[biome]; ok {
		precipPool = bp
	}
	temp := randChoice(tempOptions)
	sky := randChoice(skies)
	precip := randChoice(precipPool)
	wind := randChoice(winds)
	special := randChoice(specialties)

	// Season adjustments only apply to the temperate default; other biomes
	// already constrain their own precipitation pools.
	if biome == "temperate" {
		if season == "Winter" && precip == "Rain" {
			precip = "Snow"
		}
		if season == "Summer" && precip == "Snow" {
			precip = "None"
		}
		if season == "Spring" && precip == "Heavy Snow" {
			precip = "Rain"
		}
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
		Biome:         biome,
		Temperature:   temp,
		Sky:           sky,
		Precipitation: precip,
		Wind:          wind,
		Special:       special,
		Description:   desc,
	})
}
