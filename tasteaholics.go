package main

import (
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

func parseTasteaholicsRecipeTimes(e *colly.HTMLElement) RecipeTimes {

	cookTimeValue := e.DOM.Find(".wpurp-recipe-cook-time").First().Text()
	cookTimeText := e.DOM.Find(".wpurp-recipe-cook-time-text").First().Text()

	prepTimeValue := e.DOM.Find(".wpurp-recipe-prep-time").First().Text()
	prepTimeText := e.DOM.Find(".wpurp-recipe-prep-time-text").First().Text()

	totalTimeValue := e.DOM.Find(".wpurp-recipe-total-time").First().Text()
	totalTimeText := e.DOM.Find(".wpurp-recipe-total-time-text").First().Text()

	return RecipeTimes{
		Cook:  cookTimeValue + " " + cookTimeText,
		Prep:  prepTimeValue + " " + prepTimeText,
		Total: totalTimeValue + " " + totalTimeText,
	}
}

func parseTasteaholicsMacroValue(s string, v string, m *Macros) {
	if strings.Contains(s, "calories") {
		m.Calories = parseFloat(v)
	} else if strings.Contains(s, "fat") {
		m.Fat = parseFloat(v)
	} else if strings.Contains(s, "carbs") {
		m.Carbs = parseFloat(v)
	} else if strings.Contains(s, "protein") {
		m.Protein = parseFloat(v)
	} else if strings.Contains(s, "fiber") {
		m.Fiber = parseFloat(v)
	}
}

func scrapeTasteaholicsMacros(e *colly.HTMLElement) Macros {
	macroText := e.DOM.Find(".wpurp-recipe-description").First().Text()
	macroParts := strings.Split(macroText, "\n")

	macros := Macros{}
	for _, s := range macroParts {
		info := strings.ToLower(s)
		re := regexp.MustCompile("([0-9]+.?[0-9]+)")
		match := re.FindStringSubmatch(info)
		if len(match) == 2 {
			parseTasteaholicsMacroValue(info, match[1], &macros)
		}
	}

	return macros
}

func scrapeTasteaholicsRecipe(e *colly.HTMLElement, URL string) Recipe {
	m := scrapeTasteaholicsMacros(e)

	var ingredients []string
	e.ForEach("li.wpurp-recipe-ingredient > span.wpurp-box", func(_ int, f *colly.HTMLElement) {
		ingredients = append(ingredients, strings.Trim(f.Text, ""))
	})

	var instructions []string
	e.ForEach("span.wpurp-recipe-instruction-group", func(_ int, f *colly.HTMLElement) {
		instructions = append(instructions, strings.Trim(f.Text, ""))
	})

	title := e.DOM.Find(".wpurp-recipe-title").First().Text()
	yield := e.DOM.Find(".wpurp-recipe-servings").First().Text()
	return Recipe {
		ID:           hash(URL),
		Title:        title,
		Route:        URL,
		Source:       getHost(URL),
		Yield:        parseFloat(yield),
		Ingredients:  ingredients,
		Instructions: instructions,
		Time:         parseTasteaholicsRecipeTimes(e),
		Macros:       m,
	}
}

//func scrapeTasteaholicsRecipeHandler(_ http.ResponseWriter, r *http.Request) {
//	URL := r.FormValue("url")
//	if URL == "" {
//		log.Println("missing URL argument")
//		return
//	}
//
//	ctx := appengine.NewContext(r)
//	appEngineLog.Infof(ctx, "Scraping %s", URL)
//
//	c := colly.NewCollector()
//	c.Appengine(ctx)
//
//	c.OnHTML("div.wpurp-container", func(e *colly.HTMLElement) {
//		m := scrapeTasteaholicsMacros(e)
//
//		var ingredients []string
//		e.ForEach("li.wpurp-recipe-ingredient > span.wpurp-box", func(_ int, f *colly.HTMLElement) {
//			ingredients = append(ingredients, strings.Trim(f.Text, ""))
//		})
//
//		var instructions []string
//		e.ForEach("span.wpurp-recipe-instruction-group", func(_ int, f *colly.HTMLElement) {
//			instructions = append(instructions, strings.Trim(f.Text, ""))
//		})
//
//		title := e.DOM.Find(".wpurp-recipe-title").First().Text()
//		yield := e.DOM.Find(".wpurp-recipe-servings").First().Text()
//		r := Recipe {
//			ID:           hash(URL),
//			Title:        title,
//			Route:        URL,
//			Source:       getHost(URL),
//			Yield:        parseFloat(yield),
//			Ingredients:  ingredients,
//			Instructions: instructions,
//			Time:         parseTasteaholicsRecipeTimes(e),
//			Macros:       m,
//		}
//
//		saveRecipe(ctx, r)
//	})
//
//	c.Visit(URL)
//}

