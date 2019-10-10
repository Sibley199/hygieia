package main

import (
	"firebase.google.com/go"
	"google.golang.org/api/option"
	"google.golang.org/appengine"
	"google.golang.org/appengine/taskqueue"
	"log"
	"net/http"
)

func cronSitemapHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	opt := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Println("Failed to initialize firebase")
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Println("Failed to initialize firestore")
		log.Fatalln(err)
	}

	defer client.Close()
	docs, err := client.Collection("sitemap").Documents(ctx).GetAll()
	if err != nil {
		log.Fatalln(err)
	}

	for _, u := range docs {
		if u.Data()["shouldScrape"] == false { continue }

		route := u.Data()["route"]
		t := taskqueue.NewPOSTTask("/api/scrape/sitemap", map[string][]string{"url": {route.(string)}})
		if _, err := taskqueue.Add(ctx, t, "sitemaps"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
