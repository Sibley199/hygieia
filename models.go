package main

import "time"

type Location struct {
	Id 			 string    `firestore:"id"`
	IsRecipe     bool      `firestore:"isRecipe"`
	Route		 string    `firestore:"route"`
	Source 		 string    `firestore:"source"`
	LastModified time.Time `firestore:"lastModified"`
	LastScraped  time.Time `firestore:"lastScraped,omitempty"`
	Status       int64     `firestore:"status"`
}

type Macros struct {
	Calories     float64 `firestore:"calories"`
	Carbs        float64 `firestore:"carbs"`
	Fat          float64 `firestore:"fat"`
	Protein      float64 `firestore:"protein"`
	Fiber        float64 `firestore:"fiber"`
}

type Ingredient struct {
	Amount      float64 `firestore:"amount"`
	Name        string  `firestore:"name"`
	Measurement string  `firestore:"measurement"`
}

type RecipeTimes struct {
	Cook  string `firestore:"cook"`
	Prep  string `firestore:"prep"`
	Total string `firestore:"total"`
}

type Recipe struct {
	ID           string      `firestore:"id"`
	Title        string      `firestore:"title"`
	Route        string      `firestore:"route"`
	Source       string      `firestore:"source"`
	Macros       Macros      `firestore:"macros"`
	Yield        float64     `firestore:"yield"`
	Ingredients  []string    `firestore:"ingredients"`
	Instructions []string    `firestore:"instructions"`
	Time         RecipeTimes `firestore:"time"`
}
