package main

import (
	"api/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/graphql-go/graphql"
)

var movies []*models.Movie

// GraphQL schema definition
var fields = graphql.Fields{
	"movie": &graphql.Field{
		Type:        movieType,
		Description: "Get movie by ID",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id, ok := p.Args["id"].(int)
			if ok {
				for _, movie := range movies {
					if movie.ID == id {
						return movie, nil
					}
				}
			}
			return nil, nil
		},
	},
	"list": &graphql.Field{
		Type:        graphql.NewList(movieType),
		Description: "Get all movies",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return movies, nil
		},
	},
	"search": &graphql.Field{
		Type:        graphql.NewList(movieType),
		Description: "Search movies by title",
		Args: graphql.FieldConfigArgument{
			"titleContains": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			var results []*models.Movie
			search, ok := p.Args["titleContains"].(string)
			if ok {
				for _, currentMovie := range movies {
					if strings.Contains(currentMovie.Title, search) {
						results = append(results, currentMovie)
					}
				}
			}
			return results, nil
		},
	},
}

var movieType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Movie",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"title": &graphql.Field{
			Type: graphql.String,
		},
		"description": &graphql.Field{
			Type: graphql.String,
		},
		"release_year": &graphql.Field{
			Type: graphql.Int,
		},
		"release_date": &graphql.Field{
			Type: graphql.DateTime,
		},
		"runtime": &graphql.Field{
			Type: graphql.Int,
		},
		"raiting": &graphql.Field{
			Type: graphql.Int,
		},
		"mpaa_raiting": &graphql.Field{
			Type: graphql.String,
		},
		"created_at": &graphql.Field{
			Type: graphql.DateTime,
		},
		"updated_at": &graphql.Field{
			Type: graphql.DateTime,
		},
		"poster": &graphql.Field{
			Type: graphql.String,
		},
	},
})

func (app *application) moviesGraphQL(w http.ResponseWriter, r *http.Request) {
	movies, _ = app.models.DB.All()

	q, err := io.ReadAll(r.Body)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	query := string(q)

	rootQuery := graphql.ObjectConfig{
		Name:   "RootQuery",
		Fields: fields,
	}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		app.errorJSON(w, errors.New("failed to create graphql schema"))
		log.Println(err)
		return
	}

	params := graphql.Params{Schema: schema, RequestString: query}
	resp := graphql.Do(params)
	if len(resp.Errors) > 0 {
		app.errorJSON(w, fmt.Errorf("failed: %+v", resp.Errors))
		return
	}

	j, _ := json.MarshalIndent(resp, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)

}
