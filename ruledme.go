package main

import (
	"errors"
	"github.com/gocolly/colly"
	"regexp"
)

func scrapeRuledMeYieldAndMacros(e *colly.HTMLElement) (float64, Macros, error) {
	f := e.DOM.Find("table tr td")
	if f.Length() == 0 {
		return 0, Macros {}, errors.New("there are no macros in this 'recipe'")
	}

	lastTableRow := f.Slice(f.Length()-7, f.Length())
	re := regexp.MustCompile("\\(\\/?([0-9]+)\\)")
	match := re.FindStringSubmatch(lastTableRow.Eq(0).Text())

	yield := 1.0
	if len(match) > 0 {
		yield = parseFloat(match[1])
	}

	return yield, Macros {
		Calories: parseFloat(lastTableRow.Eq(1).Text()),
		Carbs:    parseFloat(lastTableRow.Eq(3).Text()),
		Fat:      parseFloat(lastTableRow.Eq(2).Text()),
		Protein:  parseFloat(lastTableRow.Eq(6).Text()),
		Fiber:    parseFloat(lastTableRow.Eq(5).Text()),
	}, nil
}

func scrapeRuleMeRecipe(e *colly.HTMLElement, URL string) Recipe {
	y, m, err := scrapeRuledMeYieldAndMacros(e)
	if err != nil {
		// TODO: raise error and set failed scrape
		//setFailedScrape(ctx, URL)
		//appEngineLog.Infof(ctx, "%s failed to scrape, possibly not a recipe", URL)
		//return
	}

	var ingredients []string
	e.ForEach("li.ingredient", func(_ int, f *colly.HTMLElement) {
		ingredients = append(ingredients, f.Text)
	})

	var instructions []string
	e.ForEach("li.instruction", func(_ int, f *colly.HTMLElement) {
		instructions = append(instructions, f.Text)
	})

	return Recipe {
		ID:           hash(URL),
		Title:        e.DOM.Find("div[itemprop=name]").Text(),
		Route:        URL,
		Source:       getHost(URL),
		Yield:        y,
		Ingredients:  ingredients,
		Instructions: instructions,
		Time:         RecipeTimes{},
		Macros:       m,
	}
}

//func scrapeRuledMeRecipeHandler(_ http.ResponseWriter, r *http.Request) {
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
//	c.OnHTML("div.post.category-keto-recipes", func(e *colly.HTMLElement) {
//		y, m, err := scrapeRuledMeYieldAndMacros(e)
//		if err != nil {
//			appEngineLog.Infof(ctx, "%s failed to scrape, possibly not a recipe", URL)
//			setFailedScrape(ctx, URL)
//			return
//		}
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
//		r := Recipe {
//			ID:           hash(URL),
//			Title:        e.DOM.Find("div[itemprop=name]").Text(),
//			Route:        URL,
//			Source:       getHost(URL),
//			Yield:        y,
//			Ingredients:  ingredients,
//			Instructions: instructions,
//			Time:         RecipeTimes{},
//			Macros:       m,
//		}
//
//		saveRecipe(ctx, r)
//	})
//
//	c.OnError(func(r *colly.Response, err error) {
//		appEngineLog.Criticalf(ctx, "Errored out")
//	})
//
//	c.Visit(URL)
//}
