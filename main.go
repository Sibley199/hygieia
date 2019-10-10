package main

import (
	"encoding/json"
	"google.golang.org/appengine/taskqueue"
	"log"
	"net/http"
	"time"

	"github.com/gocolly/colly"
	"google.golang.org/appengine"
	appEngineLog "google.golang.org/appengine/log"
)

func scheduleSitemapScrapeHandler(w http.ResponseWriter, r *http.Request) {
	URL := r.FormValue("url")
	if URL == "" {
		// TODO: How to return 400 error code here
		return
	}

	t := taskqueue.NewPOSTTask("/api/scrape/sitemap", map[string][]string{"url": {URL}})
	ctx := appengine.NewContext(r)
	if _, err := taskqueue.Add(ctx, t, "sitemaps"); err != nil {
		log.Fatalln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// accepts a url and lastModified timestamp and determines if we need to scrape the potential recipe
func scrapeLocationHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	URL := r.FormValue("url")
	if URL == "" {
		appEngineLog.Criticalf(ctx, "No URL passed in")
		// TODO: return 400 error code here to retry task?
		return
	}

	selector := r.FormValue("selector")
	if selector == "" {
		appEngineLog.Criticalf(ctx, "No selector passed in")
		// TODO: return 400 error code here to retry task?
		return
	}

	lastModified := r.FormValue("lastModified")
	t, err := time.Parse(time.RFC3339, lastModified)
	if err != nil {
		appEngineLog.Criticalf(ctx, "Failed to parse time: %s", lastModified)
		// TODO: return 400 error code here to retry task?
		return
	}

	// create a Collector with Appengine context
	// TODO: Investigate caching and extensions.RandomUserAgent(c)
	c := colly.NewCollector()
	c.Appengine(ctx)

	ID       := hash(URL)
	location := getLocation(ctx, ID)
	if location == nil {
		appEngineLog.Infof(ctx, "No location found, creating one now")
		location = &Location{
			Id:           ID,
			IsRecipe:     false,
			Route:        URL,
			Source:       getHost(URL),
			Status:       1,
			LastModified: t,
			LastScraped:  time.Time{},
		}
	}

	// if the url hasn't been modified since we last scraped the page then don't bother scraping
	if location.LastScraped.After(location.LastModified) {
		appEngineLog.Infof(ctx, "location last scraped is after last modified")
		return
	}

	// save location
	saveLocation(ctx, *location)

	// create task to scrape the potential recipe
	params  := map[string][]string{"url": {location.Route}, "selector": {selector}}
	task := taskqueue.NewPOSTTask("/api/scrape/recipe", params)
	if _, err := taskqueue.Add(ctx, task, "recipe"); err != nil {
		log.Fatalln(err)
		return
	}
}

type scrapeSitemapRequestMessage struct {
	URL      string
}

// accepts sitemap url and iterates through each url spinning off a task with location url and last modified timestamp
func scrapeSitemapHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	decoder := json.NewDecoder(r.Body)

	var body scrapeSitemapRequestMessage
	err := decoder.Decode(&body)
	if err != nil {
		appEngineLog.Criticalf(ctx, "Failed to parse json request")
		http.Error(w, "Failed to parse json request", http.StatusInternalServerError)
		return
	}

	if body.URL == "" {
		appEngineLog.Criticalf(ctx, "No URL passed in")
		http.Error(w, "Missing required `url` field", http.StatusBadRequest)
		return
	}

	// Create a Collector with Appengine context
	// TODO: Investigate caching, extensions.RandomUserAgent(c), and allowed domains
	c := colly.NewCollector()
	c.Appengine(ctx)

	host     := getHost(body.URL)
	selector := getSelector(ctx, host)
	if selector == "" {
		appEngineLog.Criticalf(ctx, "No selector found, can't scrape %s sitemap", host)
		// TODO: return error code here
		http.Error(w, "Couldn't find selector for request", http.StatusBadRequest)
		return
	}

	counter := 0
	// Find all locations in sitemap and create task to scrape location
	c.OnXML("//urlset/url", func(e *colly.XMLElement) {
		loc := e.ChildText("loc")
		lastModified := e.ChildText("lastmod")

		params := map[string][]string{"url": {loc}, "lastModified": {lastModified}, "selector": {selector}}
		t := taskqueue.NewPOSTTask("/api/scrape/sitemap/location", params)
		_, err := taskqueue.Add(ctx, t, "sitemaps");
		if err != nil {
			log.Fatalln(err)
			// TODO: Raise error or return bad response?
			return
		}
		counter++
	})

	c.Visit(body.URL)
	saveTotalCount(ctx, host, counter)
}

type scrapeRecipeRequestMessage struct {
	URL      string
	Selector string
}

func scrapeRecipeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := appengine.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var body scrapeRecipeRequestMessage
	err := decoder.Decode(&body)
	if err != nil {
		appEngineLog.Criticalf(ctx, "Failed to parse json request")
		http.Error(w, "Failed to parse json request", http.StatusInternalServerError)
		return
	}

	if body.URL == "" {
		appEngineLog.Criticalf(ctx, "No URL passed in")
		http.Error(w, "Missing required `url` field", http.StatusBadRequest)
		return
	}

	if body.Selector == "" {
		appEngineLog.Criticalf(ctx, "No selector passed in")
		http.Error(w, "Missing required `selector` field", http.StatusBadRequest)
		// TODO: return 400 error code here to retry task?
		return
	}

	c := colly.NewCollector()
	c.Appengine(ctx)

	var recipe Recipe
	c.OnHTML(body.Selector, func(e *colly.HTMLElement) {
		// TODO: if any fail raise error and set failed scrape
		switch host := getHost(body.URL); host {
		case "cavemanketo":
			recipe = scrapeCavemanKetoRecipe(e, body.URL)
		case "ibreatheimhungry":
			recipe = scrapeIBreatheImHungryRecipe(e, body.URL)
		case "ketoconnect":
			recipe = defaultRecipe(e, body.URL)
		case "ketogasm":
			recipe = defaultRecipe(e, body.URL)
		case "lowcarbmaven":
			recipe = defaultRecipe(e, body.URL)
		case "nobunplease":
			recipe = defaultRecipe(e, body.URL)
		case "tasteaholics":
			recipe = scrapeTasteaholicsRecipe(e, body.URL)
		case "ruled":
			recipe = defaultRecipe(e, body.URL)
		default:
			log.Fatalln("Unknown recipe host")
		}

		saveRecipe(ctx, recipe)
	})

	c.Visit(body.URL)
}

func main() {
	http.HandleFunc("/api/schedule/sitemap", scheduleSitemapScrapeHandler)

	http.HandleFunc("/api/scrape/sitemap", scrapeSitemapHandler)
	http.HandleFunc("/api/scrape/sitemap/location", scrapeLocationHandler)
	http.HandleFunc("/api/scrape/recipe", scrapeRecipeHandler)

	http.HandleFunc("/cron/sitemap", cronSitemapHandler)

	appengine.Main() // Starts the server to receive requests
}
