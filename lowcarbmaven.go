package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"regexp"
	"strings"
)

func parseMacroValue(s string, v string, m *Macros) {
	if strings.Contains(s, "calories") {
		m.Calories = parseFloat(v)
	} else if strings.Contains(s, "total fat") {
		m.Fat = parseFloat(v)
	} else if strings.Contains(s, "total carbohydrates") {
		m.Carbs = parseFloat(v)
	} else if strings.Contains(s, "protein") {
		m.Protein = parseFloat(v)
	} else if strings.Contains(s, "fiber") {
		m.Fiber = parseFloat(v)
	}
}

func scrapeMacros(e *colly.HTMLElement) Macros {
	nutritionItems := e.DOM.Find(".nutrition-main")
	macros := Macros{}
	if nutritionItems.Length() == 0 {
		return macros
	}

	nutritionItems.Each(func(_ int, s *goquery.Selection) {
		re    := regexp.MustCompile("([0-9]+)")
		info  := strings.ToLower(s.Text())
		match := re.FindStringSubmatch(info)
		if 0 < len(match) {
			value := match[1]
			parseMacroValue(info, value, &macros)
		}
	})

	nutritionSubItems := e.DOM.Find(".nutrition-sub")
	if 0 < nutritionSubItems.Length() {
		nutritionSubItems.Each(func(_ int, s *goquery.Selection) {
			info := strings.ToLower(s.Text())
			re := regexp.MustCompile("([0-9]+)")
			match := re.FindStringSubmatch(info)
			value := match[1]

			parseMacroValue(info, value, &macros)
		})
	}

	return macros
}

func parseTitle(s string) string {
	var title = ""
	titleParts := strings.Split(s, ",")
	l := len(titleParts)
	if 1 == l {
		title = titleParts[0]
	} else if 0 < l {
		title = titleParts[l-1]
	}
	return title
}

func defaultRecipe(e *colly.HTMLElement, URL string) Recipe {
	times := e.DOM.Find(".wprm-recipe-time")
	var totalTimeIndex = 2
	if times.Length() == 4 {
		totalTimeIndex = 3
	}
	t := RecipeTimes{
		Cook:  times.Eq(1).Text(),
		Prep:  times.Eq(0).Text(),
		Total: times.Eq(totalTimeIndex).Text(),
	}

	m := scrapeMacros(e)

	var ingredients []string
	e.ForEach("li.wprm-recipe-ingredient", func(_ int, f *colly.HTMLElement) {
		ingredients = append(ingredients, strings.Trim(f.Text, ""))
	})

	var instructions []string
	e.ForEach("div.wprm-recipe-instruction-text", func(_ int, f *colly.HTMLElement) {
		instructions = append(instructions, strings.Trim(f.Text, ""))
	})

	titleText := e.DOM.Find(".wprm-recipe-name").Text()
	yield := 1.0
	yieldString := e.DOM.Find(".wprm-recipe-servings").Text()
	if yieldString != "" {
		yield = parseFloat(yieldString)
	}

	return Recipe {
		ID:           hash(URL),
		Title:        parseTitle(titleText),
		Route:        URL,
		Source:       getHost(URL),
		Yield:        yield,
		Ingredients:  ingredients,
		Instructions: instructions,
		Time:         t,
		Macros:       m,
	}
}
