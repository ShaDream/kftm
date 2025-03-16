package kinopoisk

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	baseAddress = "https://api.kinopoisk.dev"
	apiVersion  = "v1.4"
)

var ErrBadResponse = errors.New("bad response")

type Client struct {
	token  string
	client *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:  token,
		client: &http.Client{},
	}
}

func (c *Client) get(endpoint string, opts map[string]string) (*http.Response, error) {
	var err error
	endpoint, err = url.JoinPath(baseAddress, apiVersion, endpoint)
	if err != nil {
		return nil, fmt.Errorf("bad request url: %w", err)
	}

	request, err := http.NewRequest(http.MethodGet, endpoint, nil)

	var query url.Values
	for key, value := range opts {
		query.Add(key, value)
	}

	request.URL.RawQuery = query.Encode()

	request.Header.Add("X-API-KEY", c.token)

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("do get request: %w", err)
	}

	return response, nil
}

func (c *Client) GetById(id int) (*Movie, error) {
	endpoint := fmt.Sprintf("movie/%d", id)
	response, err := c.get(endpoint, nil)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, wrapWrongStatusCode(response.StatusCode)
	}

	var result Movie
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("decode movie dto: %w", err)
	}

	return &result, nil
}

func wrapWrongStatusCode(statusCode int) error {
	return fmt.Errorf("wrong status code %d: %w", statusCode, ErrBadResponse)
}

type Movie struct {
	ID                int64        `json:"id"`
	ExternalID        ExternalID   `json:"externalId"`
	Name              string       `json:"name"`
	AlternativeName   string       `json:"alternativeName"`
	EnName            interface{}  `json:"enName"`
	Names             []Name       `json:"names"`
	Type              string       `json:"type"`
	TypeNumber        int64        `json:"typeNumber"`
	Year              int64        `json:"year"`
	Description       string       `json:"description"`
	ShortDescription  string       `json:"shortDescription"`
	Slogan            string       `json:"slogan"`
	Status            interface{}  `json:"status"`
	Rating            Rating       `json:"rating"`
	Votes             Rating       `json:"votes"`
	MovieLength       int64        `json:"movieLength"`
	TotalSeriesLength interface{}  `json:"totalSeriesLength"`
	SeriesLength      interface{}  `json:"seriesLength"`
	RatingMPAA        string       `json:"ratingMpaa"`
	AgeRating         int64        `json:"ageRating"`
	Poster            Backdrop     `json:"poster"`
	Backdrop          Backdrop     `json:"backdrop"`
	Genres            []Country    `json:"genres"`
	Countries         []Country    `json:"countries"`
	Persons           []Person     `json:"persons"`
	Premiere          Premiere     `json:"premiere"`
	Watchability      Watchability `json:"watchability"`
	Top10             interface{}  `json:"top10"`
	Top250            int64        `json:"top250"`
	IsSeries          bool         `json:"isSeries"`
	TicketsOnSale     bool         `json:"ticketsOnSale"`
	Lists             []string     `json:"lists"`
	Networks          interface{}  `json:"networks"`
	CreatedAt         time.Time    `json:"createdAt"`
	UpdatedAt         time.Time    `json:"updatedAt"`
	Fees              Fees         `json:"fees"`
	Videos            Videos       `json:"videos"`
	Logo              Backdrop     `json:"logo"`
	IsTmdbChecked     bool         `json:"isTmdbChecked"`
}

type Backdrop struct {
	URL        string `json:"url"`
	PreviewURL string `json:"previewUrl"`
}

type Country struct {
	Name string `json:"name"`
}

type ExternalID struct {
	KpHD string `json:"kpHD"`
	Imdb string `json:"imdb"`
	Tmdb int64  `json:"tmdb"`
}

type Fees struct {
	Russia Russia `json:"russia"`
	Usa    Russia `json:"usa"`
	World  Russia `json:"world"`
}

type Russia struct {
	Value    int64  `json:"value"`
	Currency string `json:"currency"`
}

type Name struct {
	Name     string      `json:"name"`
	Language string      `json:"language"`
	Type     interface{} `json:"type"`
}

type Person struct {
	ID           int64        `json:"id"`
	Photo        string       `json:"photo"`
	Name         *string      `json:"name"`
	EnName       *string      `json:"enName"`
	Description  *string      `json:"description"`
	Profession   Profession   `json:"profession"`
	EnProfession EnProfession `json:"enProfession"`
}

type Premiere struct {
	Country interface{} `json:"country"`
	Cinema  interface{} `json:"cinema"`
	Bluray  interface{} `json:"bluray"`
	DVD     interface{} `json:"dvd"`
	Digital time.Time   `json:"digital"`
	Russia  time.Time   `json:"russia"`
	World   time.Time   `json:"world"`
}

type Rating struct {
	Kp                 float64 `json:"kp"`
	Imdb               float64 `json:"imdb"`
	FilmCritics        float64 `json:"filmCritics"`
	RussianFilmCritics float64 `json:"russianFilmCritics"`
	Await              float64 `json:"await"`
}

type Videos struct {
	Trailers []Trailer `json:"trailers"`
}

type Trailer struct {
	URL  string `json:"url"`
	Name string `json:"name"`
	Site string `json:"site"`
	Type string `json:"type"`
}

type Watchability struct {
	Items []Item `json:"items"`
}

type Item struct {
	Name string `json:"name"`
	Logo Logo   `json:"logo"`
	URL  string `json:"url"`
}

type Logo struct {
	URL string `json:"url"`
}

type EnProfession string

const (
	Actor    EnProfession = "actor"
	Composer EnProfession = "composer"
	Designer EnProfession = "designer"
	Director EnProfession = "director"
	Editor   EnProfession = "editor"
	Operator EnProfession = "operator"
	Producer EnProfession = "producer"
	Writer   EnProfession = "writer"
)

type Profession string

const (
	Актеры      Profession = "актеры"
	Композиторы Profession = "композиторы"
	Монтажеры   Profession = "монтажеры"
	Операторы   Profession = "операторы"
	Продюсеры   Profession = "продюсеры"
	Режиссеры   Profession = "режиссеры"
	Сценаристы  Profession = "сценаристы"
	Художники   Profession = "художники"
)
