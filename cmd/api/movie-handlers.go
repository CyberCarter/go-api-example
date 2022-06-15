package main

import (
	"api/models"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (app *application) getOneMovie(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.logger.Println(errors.New("invalid id parameter"))
		app.errorJSON(w, err)
		return
	}

	movie, err := app.models.DB.Get(id)
	if err != nil {
		app.logger.Println(err)
	}

	err = app.writeJSON(w, http.StatusOK, movie, "movie")
	if err != nil {
		app.logger.Println(err)
	}

}

func (app *application) getAllMovies(w http.ResponseWriter, r *http.Request) {

	movies, err := app.models.DB.All()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, movies, "movies")
	if err != nil {
		app.errorJSON(w, err)
		return
	}

}

func (app *application) getAllGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := app.models.DB.GetAllGenres()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, genres, "genres")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}

func (app *application) getAllMoviesByGenre(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	genreID, err := strconv.Atoi(params.ByName("genre_id"))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	movies, err := app.models.DB.All(genreID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, movies, "movies")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}

func (app *application) deleteMovie(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.models.DB.DeleteMovie(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	ok := jsonResponse{
		OK:      true,
		Message: "Movie has been deleted",
	}

	err = app.writeJSON(w, http.StatusOK, ok, "response")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}

type MoviePayload struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ReleaseYear string `json:"release_year"`
	ReleaseDate string `json:"release_date"`
	Runtime     string `json:"runtime"`
	Raiting     string `json:"raiting"`
	MPAARaiting string `json:"mpaa_raiting"`
}

func (app *application) editMovie(w http.ResponseWriter, r *http.Request) {
	var payload MoviePayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var movie models.Movie

	if payload.ID != "0" {
		id, _ := strconv.Atoi(payload.ID)
		m, _ := app.models.DB.Get(id)
		movie = *m
		movie.UpdatedAt = time.Now()
	}

	movie.ID, _ = strconv.Atoi(payload.ID)
	movie.Title = payload.Title
	movie.Description = payload.Description
	movie.ReleaseDate, _ = time.Parse("2006-01-02", payload.ReleaseDate)
	movie.ReleaseYear = movie.ReleaseDate.Year()
	movie.Runtime, _ = strconv.Atoi(payload.Runtime)
	movie.Raiting, _ = strconv.Atoi(payload.Raiting)
	movie.MPAARaiting = payload.MPAARaiting
	movie.CreatedAt = time.Now()
	movie.UpdatedAt = time.Now()

	if movie.Poster == "" {
		movie = getPoster(movie)
	}

	if movie.ID == 0 {
		err = app.models.DB.InsertMovie(movie)
		if err != nil {
			app.errorJSON(w, err)
			return
		}
	} else {
		err = app.models.DB.UpdateMovie(movie)
		if err != nil {
			app.errorJSON(w, err)
			return
		}
	}

	ok := jsonResponse{
		OK:      true,
		Message: "inserted into db",
	}

	err = app.writeJSON(w, http.StatusOK, ok, "response")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}

func getPoster(movie models.Movie) models.Movie {
	type TheMovieDB struct {
		Page    int `json:"page"`
		Results []struct {
			Adult            bool    `json:"adult"`
			BackdropPath     string  `json:"backdrop_path"`
			GenreIds         []int   `json:"genre_ids"`
			ID               int     `json:"id"`
			OriginalLanguage string  `json:"original_language"`
			OriginalTitle    string  `json:"original_title"`
			Overview         string  `json:"overview"`
			Popularity       float64 `json:"popularity"`
			PosterPath       string  `json:"poster_path"`
			ReleaseDate      string  `json:"release_date"`
			Title            string  `json:"title"`
			Video            bool    `json:"video"`
			VoteAverage      float64 `json:"vote_average"`
			VoteCount        int     `json:"vote_count"`
		} `json:"results"`
		TotalPages   int `json:"total_pages"`
		TotalResults int `json:"total_results"`
	}

	client := &http.Client{}
	key := "c24bc10db92d9428ae369ca2331b8b3d"
	movieDBURL := "https://api.themoviedb.org/3/search/movie?api_key="
	fullUrl := movieDBURL + key + "&query=" + url.QueryEscape(movie.Title)
	log.Println(fullUrl)

	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		log.Println(err)
		return movie
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return movie
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return movie
	}

	var responseObj TheMovieDB

	json.Unmarshal(bodyBytes, &responseObj)

	if len(responseObj.Results) > 0 {
		movie.Poster = responseObj.Results[0].PosterPath
	}

	return movie
}
