package main

import (
	"github.com/gocolly/colly"
	"regexp"
	"strings"
)

func parseIBreatheImHungryMacroValue(s string) float64 {
	re := regexp.MustCompile("([0-9]+)")
	match := re.FindStringSubmatch(s)
	if len(match) == 0 {
		return 0
	}
	return parseFloat(match[1])
}

func scrapeIBreatheImHungryRecipe(e *colly.HTMLElement, URL string) Recipe {
	title := e.DOM.Find("h2").Text()
	yield := e.DOM.Find(".tasty-recipes-details li.yield").Text()
	yield =  strings.Split(yield, " ")[1]

	t := RecipeTimes{
		Prep:  e.DOM.Find(".tasty-recipes-prep-time").Text(),
		Cook:  e.DOM.Find(".tasty-recipes-cook-time").Text(),
		Total: e.DOM.Find(".tasty-recipes-total-time").Text(),
	}

	m := Macros{
		Calories: parseIBreatheImHungryMacroValue(e.DOM.Find(".tasty-recipes-calories").Text()),
		Carbs:    parseIBreatheImHungryMacroValue(e.DOM.Find(".tasty-recipes-carbohydrates").Text()),
		Fat:      parseIBreatheImHungryMacroValue(e.DOM.Find(".tasty-recipes-fat").Text()),
		Protein:  parseIBreatheImHungryMacroValue(e.DOM.Find(".tasty-recipes-protein").Text()),
		Fiber:    parseIBreatheImHungryMacroValue(e.DOM.Find(".tasty-recipes-fiber").Text()),
	}

	var ingredients []string
	e.ForEach(".tasty-recipe-ingredients li", func(_ int, f *colly.HTMLElement) {
		ingredients = append(ingredients, strings.Trim(f.Text, ""))
	})

	var instructions []string
	e.ForEach(".tasty-recipe-instructions li", func(_ int, f *colly.HTMLElement) {
		instructions = append(instructions, strings.Trim(f.Text, ""))
	})

	return Recipe {
		ID:           hash(URL),
		Title:        title,
		Route:        URL,
		Source:       getHost(URL),
		Yield:        parseFloat(yield),
		Ingredients:  ingredients,
		Instructions: instructions,
		Time:         t,
		Macros:       m,
	}
}

//func scrapeIBreatheImHungryRecipeHandler(_ http.ResponseWriter, r *http.Request) {
//	URL := r.FormValue("url")
//	if URL == "" {
//		log.Println("missing URL argument")
//		return
//	}
//
//	ctx := appengine.NewContext(r)
//	c := colly.NewCollector()
//	c.Appengine(ctx)
//
//	c.OnHTML("div.tasty-recipes-display", func(e *colly.HTMLElement) {
//		title := e.DOM.Find("h2").Text()
//		yield := e.DOM.Find(".tasty-recipes-details li.yield").Text()
//		yield =  strings.Split(yield, " ")[1]
//
//		t := RecipeTimes{
//			Prep:  e.DOM.Find(".tasty-recipes-prep-time").Text(),
//			Cook:  e.DOM.Find(".tasty-recipes-cook-time").Text(),
//			Total: e.DOM.Find(".tasty-recipes-total-time").Text(),
//		}
//
//		m := Macros{
//			Calories: parseIBreatheImHungryMacroValue(e.DOM.Find(".tasty-recipes-calories").Text()),
//			Carbs:    parseIBreatheImHungryMacroValue(e.DOM.Find(".tasty-recipes-carbohydrates").Text()),
//			Fat:      parseIBreatheImHungryMacroValue(e.DOM.Find(".tasty-recipes-fat").Text()),
//			Protein:  parseIBreatheImHungryMacroValue(e.DOM.Find(".tasty-recipes-protein").Text()),
//			Fiber:    parseIBreatheImHungryMacroValue(e.DOM.Find(".tasty-recipes-fiber").Text()),
//		}
//
//		var ingredients []string
//		e.ForEach(".tasty-recipe-ingredients li", func(_ int, f *colly.HTMLElement) {
//			ingredients = append(ingredients, strings.Trim(f.Text, ""))
//		})
//
//		var instructions []string
//		e.ForEach(".tasty-recipe-instructions li", func(_ int, f *colly.HTMLElement) {
//			instructions = append(instructions, strings.Trim(f.Text, ""))
//		})
//
//		r := Recipe {
//			ID:           hash(URL),
//			Title:        title,
//			Route:        URL,
//			Source:       getHost(URL),
//			Yield:        parseFloat(yield),
//			Ingredients:  ingredients,
//			Instructions: instructions,
//			Time:         t,
//			Macros:       m,
//		}
//
//		saveRecipe(ctx, r)
//	})
//
//	c.Visit(URL)
//}

