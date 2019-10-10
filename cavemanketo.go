package main

import (
	"github.com/gocolly/colly"
	"strings"
)

func scrapeCavemanKetoMacros(e *colly.HTMLElement) Macros {
	nutritionDom := e.DOM.Find("item[itemprop=nutrition]")
	if len(nutritionDom.Nodes) == 0 {
		return Macros{}
	}

	return Macros{
		Calories: parseFloat(nutritionDom.Find("span[itemprop=calories]").Text()),
		Carbs:    parseFloat(nutritionDom.Find("span[itemprop=carbohydrateContent]").Text()),
		Fat:      parseFloat(nutritionDom.Find("span[itemprop=fatContent]").Text()),
		Fiber:	  parseFloat(nutritionDom.Find("span[itemprop=fiberContent").Text()),
		Protein:  parseFloat(nutritionDom.Find("span[itemprop=proteinContent]").Text()),
	}
}

func scrapeCavemanKetoRecipe(e *colly.HTMLElement, URL string) Recipe {
	t := RecipeTimes{
		Cook:  e.ChildText("time[itemprop=cookTime]"),
		Prep:  e.ChildText("time[itemprop=prepTime]"),
		Total: e.ChildText("time[itemprop=totalTime]"),
	}

	m := scrapeCavemanKetoMacros(e)

	var ingredients []string
	e.ForEach("li.ingredient", func(_ int, f *colly.HTMLElement) {
		ingredients = append(ingredients, f.Text)
	})

	var instructions []string
	e.ForEach("li.instruction", func(_ int, f *colly.HTMLElement) {
		instructions = append(instructions, f.Text)
	})

	yieldString := e.ChildText("span[itemprop=recipeYield]")
	yieldString =  strings.Split(yieldString, " ")[0]
	var yield float64
	if yieldString != "" {
		y := parseFloat(yieldString)
		yield = y
	}

	return Recipe {
		ID:           hash(URL),
		Title:        e.ChildText(".ERSName"),
		Route:        URL,
		Source:       getHost(URL),
		Yield:        yield,
		Ingredients:  ingredients,
		Instructions: instructions,
		Time:         t,
		Macros:       m,
	}
}

//func scrapeCavemanKetoRecipeHandler(_ http.ResponseWriter, r *http.Request) {
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
//	c.OnHTML(`div[id=cmkRecipe]`, func(e *colly.HTMLElement) {
//		log.Println("found selector")
//		appEngineLog.Infof(ctx, "URL: %s", URL)
//		t := RecipeTimes{
//			Cook:  e.ChildText("time[itemprop=cookTime]"),
//			Prep:  e.ChildText("time[itemprop=prepTime]"),
//			Total: e.ChildText("time[itemprop=totalTime]"),
//		}
//		log.Println(t)
//
//		m := scrapeCavemanKetoMacros(e)
//
//		var ingredients []string
//		e.ForEach("li.ingredient", func(_ int, f *colly.HTMLElement) {
//			ingredients = append(ingredients, f.Text)
//		})
//
//		var instructions []string
//		e.ForEach("li.instruction", func(_ int, f *colly.HTMLElement) {
//			instructions = append(instructions, f.Text)
//		})
//
//		yieldString := e.ChildText("span[itemprop=recipeYield]")
//		yieldString =  strings.Split(yieldString, " ")[0]
//		var yield float64
//		if yieldString != "" {
//			y := parseFloat(yieldString)
//			yield = y
//		}
//
//		r := Recipe {
//			ID:           hash(URL),
//			Title:        e.ChildText(".ERSName"),
//			Route:        URL,
//			Source:       getHost(URL),
//			Yield:        yield,
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
