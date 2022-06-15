package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type DBModel struct {
	DB *sql.DB
}

// Returns one movie by id
func (m *DBModel) Get(id int) (*Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, title, description, year, release_date, rating, runtime, mpaa_rating, created_at, updated_at, coalesce(poster, '') from movies where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var movie Movie

	err := row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.Description,
		&movie.ReleaseYear,
		&movie.ReleaseDate,
		&movie.Raiting,
		&movie.Runtime,
		&movie.MPAARaiting,
		&movie.CreatedAt,
		&movie.UpdatedAt,
		&movie.Poster,
	)

	if err != nil {
		return nil, err
	}

	genres, err := m.getMovieGenre(movie.ID)
	if err != nil {
		return nil, err
	}

	movie.MovieGenre = genres

	return &movie, nil
}

// Returns all movies / error
func (m *DBModel) All(genre ...int) ([]*Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	where := ""
	if len(genre) > 0 {
		where = fmt.Sprintf("where id in (select movie_id from movies_genres where genre_id = %d)", genre[0])
	}

	query := fmt.Sprintf(`select id, title, description, year, release_date, rating, runtime, mpaa_rating, created_at, updated_at, coalesce(poster, '') from movies %s order by title`, where)

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []*Movie

	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.Description,
			&movie.ReleaseYear,
			&movie.ReleaseDate,
			&movie.Raiting,
			&movie.Runtime,
			&movie.MPAARaiting,
			&movie.CreatedAt,
			&movie.UpdatedAt,
			&movie.Poster,
		)

		if err != nil {
			return nil, err
		}

		genres, err := m.getMovieGenre(movie.ID)
		if err != nil {
			return nil, err
		}

		movie.MovieGenre = genres
		movies = append(movies, &movie)

	}

	return movies, nil

}

// returns in a map[int]string the genres for a givem movie
func (m *DBModel) getMovieGenre(movieID int) (map[int]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	genreQuery := `select 
				mg.id, mg.movie_id, mg.genre_id, g.genre_name
			from
				movies_genres mg
				left join genres g on (g.id = mg.genre_id)
			where
				mg.movie_id = $1
	`
	genreRows, _ := m.DB.QueryContext(ctx, genreQuery, movieID)

	genres := make(map[int]string)

	for genreRows.Next() {
		var mg MovieGenre
		err := genreRows.Scan(
			&mg.ID,
			&mg.MovieID,
			&mg.GenreID,
			&mg.Genre.GenreName,
		)

		if err != nil {
			return nil, err
		}

		genres[mg.ID] = mg.Genre.GenreName
	}
	genreRows.Close()

	return genres, nil
}

func (m *DBModel) GetAllGenres() ([]*Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, genre_name, created_at, updated_at from genres order by genre_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []*Genre

	for rows.Next() {
		var g Genre

		err = rows.Scan(
			&g.ID,
			&g.GenreName,
			&g.CreatedAt,
			&g.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		genres = append(genres, &g)
	}

	return genres, nil
}

func (m *DBModel) InsertMovie(movie Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	statement := `insert into movies (title, description, year, release_date, runtime, rating, mpaa_rating, created_at, updated_at, poster) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := m.DB.ExecContext(ctx, statement,
		movie.Title,
		movie.Description,
		movie.ReleaseYear,
		movie.ReleaseDate,
		movie.Runtime,
		movie.Raiting,
		movie.MPAARaiting,
		movie.CreatedAt,
		movie.UpdatedAt,
		movie.Poster,
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) UpdateMovie(movie Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	statement := `update movies set title = $1, description = $2, year = $3, release_date = $4, runtime = $5, rating = $6, mpaa_rating = $7, updated_at = $8, poster = $9 where id = $10`

	_, err := m.DB.ExecContext(ctx, statement,
		movie.Title,
		movie.Description,
		movie.ReleaseYear,
		movie.ReleaseDate,
		movie.Runtime,
		movie.Raiting,
		movie.MPAARaiting,
		movie.UpdatedAt,
		movie.Poster,
		movie.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) DeleteMovie(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `delete from movies where id = $1`

	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}
