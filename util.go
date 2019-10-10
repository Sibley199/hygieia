package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"firebase.google.com/go"
	"google.golang.org/api/option"
	appEngineLog "google.golang.org/appengine/log"
	"hash/fnv"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func getHost(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		log.Fatal(err)
	}

	parts := strings.Split(u.Hostname(), ".")
	return parts[len(parts)-2]
}

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	value := h.Sum32()
	return strconv.FormatUint(uint64(value), 10)
}

func parseFloat(s string) float64 {
	// handle fraction serving sizes
	s = strings.Trim(s, "\u00a0")
	s = strings.Trim(s, " ")
	if strings.Index(s, "/") == 1 {
		s = "1"
	}

	// handle serving ranges likes 20-22
	stringParts := strings.Split(s, "-")
	value, err  := strconv.ParseFloat(stringParts[0], 64)
	if err != nil {
		log.Fatalln(err)
	}
	return value
}

func getFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	opt := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Println("Failed to initialize firebase")
		return nil, err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Println("Failed to initialize firestore")
		return nil, err
	}
	return client, nil
}

func setFailedScrape(ctx context.Context, URL string) {
	client, err := getFirestoreClient(ctx)
	if err != nil {
		log.Fatalln("Failed to get Firestore client")
	}

	defer client.Close()

	ID := hash(URL)
	docRef := client.Collection("location").Doc(ID)

	var updates []firestore.Update
	updates = append(updates, firestore.Update{Path: "status", Value: "-1"})
	updates = append(updates, firestore.Update{Path: "lastScraped", Value: time.Now()})
	_, err = docRef.Update(ctx, updates)
	if err != nil {
		log.Fatalln(err)
	}
}

func saveRecipe(ctx context.Context, r Recipe) {
	client, err := getFirestoreClient(ctx)
	if err != nil {
		appEngineLog.Criticalf(ctx, "Failed to initialize firestore")
		log.Fatalln(err)
	}

	defer client.Close()

	batch := client.Batch()
	appEngineLog.Infof(ctx, "Setting recipe %s", r.ID)
	docRef := client.Collection("recipe").Doc(r.ID)
	batch.Set(docRef, r)

	updates := []firestore.Update{
		{Path: "lastScraped", Value: time.Now()},
		{Path: "status", Value: 1},
		{Path: "isRecipe", Value: true},
	}
	locationRef := client.Collection("location").Doc(hash(r.Route))
	batch.Update(locationRef, updates)

	batch.Commit(ctx)
}

func getSelector(ctx context.Context, ID string) string {
	client, err := getFirestoreClient(ctx)
	if err != nil {
		appEngineLog.Criticalf(ctx, "Failed to initialize firestore")
		log.Fatalln(err)
	}

	defer client.Close()

	snapshot, err := client.Doc("sitemap/" + ID).Get(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	return snapshot.Data()["selector"].(string)
}

func getLocation(ctx context.Context, ID string) *Location {
	client, err := getFirestoreClient(ctx)
	if err != nil {
		appEngineLog.Criticalf(ctx, "Failed to initialize firestore")
		log.Fatalln(err)
	}

	defer client.Close()

	snapshot, err := client.Doc("location/" + ID).Get(ctx)
	if err != nil {
		// TODO: How to check the specific type of error
		return nil
	}

	data := snapshot.Data()
	appEngineLog.Debugf(ctx, "LOCATION OBJECT %+v", data)
	lastScraped := data["lastScraped"]
	if lastScraped == nil {
		var t interface{} = time.Time{}
		lastScraped = t
	}

	return &Location{
		Id:           data["id"].(string),
		IsRecipe:     data["isRecipe"].(bool),
		Route:        data["route"].(string),
		Source:       data["source"].(string),
		Status:       data["status"].(int64),
		LastModified: data["lastModified"].(time.Time),
		LastScraped:  lastScraped.(time.Time),
	}
}

func getRecipeCount(urls []Location) int {
	count := 0
	for _, u := range urls {
		if u.IsRecipe { count ++}
	}
	return count
}

//func saveRoutes(ctx context.Context, urls []Location) {
//	opt := option.WithCredentialsFile("serviceAccountKey.json")
//	app, err := firebase.NewApp(ctx, nil, opt)
//	if err != nil {
//		log.Println("Failed to initialize firebase")
//		log.Fatalln(err)
//	}
//
//	client, err := app.Firestore(ctx)
//	if err != nil {
//		log.Println("Failed to initialize firestore")
//		log.Fatalln(err)
//	}
//
//	defer client.Close()
//
//	batch := client.Batch()
//
//	host := getHost(urls[0].Route)
//	total := len(urls)
//	recipes := getRecipeCount(urls)
//	other := total - recipes
//
//	recipeCountRef := client.Collection("counts").Doc(host)
//	batch.Set(recipeCountRef, map[string]interface{}{
//		"total": total,
//		"recipes": recipes,
//		"other": other,
//	}, firestore.MergeAll)
//
//	var updates []firestore.Update
//	updates = append(updates, firestore.Update{Path: host, Value: total})
//	locationCountRef := client.Collection("counts").Doc("locations")
//	log.Println(locationCountRef)
//	batch.Update(locationCountRef, updates)
//
//	batch.Commit(ctx)
//
//	for _, u := range urls {
//		var updates []firestore.Update
//		updates = append(updates, firestore.Update{Path: "route", Value: u.Route})
//		updates = append(updates, firestore.Update{Path: "isRecipe", Value: u.IsRecipe})
//
//		docRef := client.Collection("location").Doc(u.Id)
//		_, err := docRef.Update(ctx, updates)
//		if err != nil {
//			_, err := docRef.Set(ctx, u)
//			if err != nil {
//				log.Fatalln(err)
//			}
//		}
//	}
//}

func saveTotalCount(ctx context.Context, host string, counter int) {
	client, err := getFirestoreClient(ctx)
	if err != nil {
		appEngineLog.Criticalf(ctx, "Failed to initialize firestore")
		log.Fatalln(err)
	}

	defer client.Close()

	updates := []firestore.Update{{Path: "total", Value: counter}}
	client.Collection("counts").Doc(host).Update(ctx, updates)
}

func saveLocation(ctx context.Context, location Location) {
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

	client.Collection("location").Doc(location.Id).Set(ctx, location)

	//batch := client.Batch()
	//host := getHost(urls[0].Route)
	//batch.Update()
	//recipeCountRef := client.Collection("counts").Doc(host)
	//batch.Set(recipeCountRef, map[string]interface{}{
	//	"total": total,
	//	"recipes": recipes,
	//	"other": other,
	//}, firestore.MergeAll)

	//var updates []firestore.Update
	//updates = append(updates, firestore.Update{Path: host, Value: total})
	//sitemapRouteCountRef := client.Collection("counts").Doc("siteMapRoutes")
	//log.Println(sitemapRouteCountRef)
	//batch.Update(sitemapRouteCountRef, updates)
	//
	//batch.Commit(ctx)
	//
	//for _, u := range urls {
	//	var updates []firestore.Update
	//	updates = append(updates, firestore.Update{Path: "route", Value: u.Route})
	//	updates = append(updates, firestore.Update{Path: "isRecipe", Value: u.IsRecipe})
	//
	//	docRef := client.Collection("siteMapRoute").Doc(u.Id)
	//	_, err := docRef.Update(ctx, updates)
	//	if err != nil {
	//		_, err := docRef.Set(ctx, u)
	//		if err != nil {
	//			log.Fatalln(err)
	//		}
	//	}
	//}
}