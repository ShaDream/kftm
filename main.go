package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-bittorrent/magneturi"
	"github.com/shadream/kftm/kinopoisk"
	"github.com/shadream/kftm/qbitorrent"
)

var (
	configPath string
	changeVar  bool
)

func init() {
	executable, _ := os.Executable()
	exePath, _ := filepath.EvalSymlinks(executable)

	defaultConfigPath := filepath.Join(filepath.Dir(exePath), "config.json")
	flag.StringVar(&configPath, "config", defaultConfigPath, "path to config")
	flag.BoolVar(&changeVar, "change", false, "change already created torrent")

	flag.Parse()
}

func main() {
	if changeVar {
		change()
	} else {
		run()
	}
}

func readConfig() (*Config, error) {
	cfgData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var config Config
	err = json.Unmarshal(cfgData, &config)
	if err != nil {
		return nil, fmt.Errorf("unmarshall json config: %w", err)
	}

	return &config, nil
}

func createQbitorrentClient(torrentConfig QbitorrentConfig) (*qbitorrent.Client, error) {
	client := qbitorrent.NewClient(torrentConfig.BaseUrl)
	err := client.Login(torrentConfig.Username, torrentConfig.Password)
	if err != nil {
		return nil, fmt.Errorf("login to qbitorrent: %w", err)
	}

	return client, nil
}

func change() {
	config, err := readConfig()
	if err != nil {
		log.Fatal(err)
	}

	tClient, err := createQbitorrentClient(config.Qbitorrent)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("paste torrent hash:")
	var hash string
	_, err = fmt.Scanln(&hash)
	if err != nil {
		log.Fatal(err)
	}

	kClient := kinopoisk.NewClient(config.KinopoiskToken)

	fmt.Print("paste kinopois url or id: ")
	var kinopoiskUrl string
	_, err = fmt.Scanln(&kinopoiskUrl)
	if err != nil {
		log.Fatal(err)
	}

	filmId := parseKinopoiskUrl(kinopoiskUrl)
	if filmId == -1 {
		log.Fatal("can not get kinopoisk film id")
	}

	fmt.Printf("kinopoisk film id: %d\n", filmId)

	movie, err := kClient.GetById(filmId)
	if err != nil {
		log.Fatal(err)
	}

	nfo := KinopoiskDtoToNfo(*movie)

	name := whitelistString(fmt.Sprintf("%s (%d)", movie.Name, movie.Premiere.World.Year()))

	fmt.Println("getting files...")

	for {
		content, err := tClient.GetTorrentContent(hash)
		if err != nil {
			log.Fatal(err)
		}

		if len(content) == 0 {
			time.Sleep(time.Second)
			continue
		}

		file := content[0]
		if len(content) != 1 {
			file = PickFile(content)
		}

		fileExt := path.Ext(*file.Name)
		err = tClient.RenameFile(qbitorrent.RenameTorrentFiles{
			Hash:    hash,
			OldPath: *file.Name,
			NewPath: fmt.Sprintf("%s/%s%s", name, name, fileExt),
		})
		if err != nil {
			log.Fatal(err)
		}

		if len(content) != 1 {
			dir := path.Dir(*file.Name)
			err = tClient.RenameFolder(qbitorrent.RenameTorrentFiles{
				Hash:    hash,
				OldPath: dir,
				NewPath: name,
			})
			if err != nil {
				log.Fatal(err)
			}

		}

		break
	}

	movieDir := filepath.Join(config.Qbitorrent.RealSavePath, name)

	for {
		_, err := os.ReadDir(movieDir)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		break
	}

	nfoPath := filepath.Join(config.Qbitorrent.RealSavePath, name, fmt.Sprintf("%s.nfo", name))
	nfoData, err := xml.Marshal(nfo)
	if err != nil {
		log.Fatal(err)
	}

	nfoData = []byte(xml.Header + string(nfoData))

	err = os.WriteFile(nfoPath, nfoData, 0o755)
	if err != nil {
		log.Fatal(err)
	}

	imagePath := filepath.Join(config.Qbitorrent.RealSavePath, name, "poster.jpg")

	downloadImage(movie.Poster.URL, imagePath)

	fmt.Println("all done!")
}

func run() {
	config, err := readConfig()
	if err != nil {
		log.Fatal(err)
	}

	tClient, err := createQbitorrentClient(config.Qbitorrent)
	if err != nil {
		log.Fatal(err)
	}

	kClient := kinopoisk.NewClient(config.KinopoiskToken)

	fmt.Printf("paste kinopois url or id: ")
	var kinopoiskUrl string
	_, err = fmt.Scanln(&kinopoiskUrl)
	if err != nil {
		log.Fatal(err)
	}

	filmId := parseKinopoiskUrl(kinopoiskUrl)
	if filmId == -1 {
		log.Fatal("can not get kinopoisk film id")
	}

	fmt.Printf("kinopoisk film id: %d\n", filmId)

	movie, err := kClient.GetById(filmId)
	if err != nil {
		log.Fatal(err)
	}

	nfo := KinopoiskDtoToNfo(*movie)

	fmt.Println("paste magnet link:")
	var magnetLink string
	_, err = fmt.Scanln(&magnetLink)
	if err != nil {
		log.Fatal(err)
	}

	magnetParsed, err := magneturi.Parse(magnetLink)
	if err != nil {
		log.Fatal(err)
	}

	hash, _ := strings.CutPrefix(magnetParsed.ExactTopics[0], "urn:btih:")

	err = tClient.CreateTorrentFileUrl(qbitorrent.AddTorrentsURLs{
		AutoTMM:  makePointer(true),
		Category: &config.Qbitorrent.Category,
		Urls:     &magnetLink,
	})
	if err != nil {
		log.Fatal(err)
	}

	name := whitelistString(fmt.Sprintf("%s (%d)", movie.Name, movie.Premiere.World.Year()))

	fmt.Println("getting files...")

	for {
		content, err := tClient.GetTorrentContent(hash)
		if err != nil {
			log.Fatal(err)
		}

		if len(content) == 0 {
			time.Sleep(time.Second)
			continue
		}

		file := PickFile(content)
		fileExt := path.Ext(*file.Name)
		err = tClient.RenameFile(qbitorrent.RenameTorrentFiles{
			Hash:    hash,
			OldPath: *file.Name,
			NewPath: fmt.Sprintf("%s/%s%s", name, name, fileExt),
		})
		if err != nil {
			log.Fatal(err)
		}

		if len(content) != 1 {
			dir := path.Dir(*file.Name)
			err = tClient.RenameFolder(qbitorrent.RenameTorrentFiles{
				Hash:    hash,
				OldPath: dir,
				NewPath: name,
			})
			if err != nil {
				log.Fatal(err)
			}

		}

		break
	}

	nfoPath := filepath.Join(config.Qbitorrent.RealSavePath, name, fmt.Sprintf("%s.nfo", name))
	nfoData, err := xml.Marshal(nfo)
	if err != nil {
		log.Fatal(err)
	}

	for {
		err = os.WriteFile(nfoPath, nfoData, 0o755)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second)
			continue
		}

		break
	}

	imagePath := filepath.Join(config.Qbitorrent.RealSavePath, name, "poster.jpg")

	downloadImage(movie.Poster.URL, imagePath)

	fmt.Println("all done!")
}

func PickFile(files []qbitorrent.TorrentsFiles) qbitorrent.TorrentsFiles {
	var index int
	fmt.Println("pick movie file:")
	for index, item := range files {
		fmt.Printf("%d) %s\t%s\n", index+1, *item.Name, humanize.Bytes(uint64(*item.Size)))
	}

	for {
		_, err := fmt.Scanln(&index)
		if err != nil {
			fmt.Println("wrong input, write index:")
			continue
		}
		if index < 1 || index > len(files) {
			fmt.Println("index is too small or too big. write index:")
			continue
		}

		break
	}

	return files[index-1]
}

var kinopoiskUrlRegex = regexp.MustCompile(`\.kinopoisk\.ru/film/(\d+)`)

func parseKinopoiskUrl(url string) int {
	id, err := strconv.Atoi(url)
	if err == nil {
		return id
	}

	match := kinopoiskUrlRegex.FindStringSubmatch(url)
	if len(match) == 0 {
		return -1
	}

	id, err = strconv.Atoi(match[1])
	if err != nil {
		return -1
	}

	return id
}

func KinopoiskDtoToNfo(dto kinopoisk.Movie) MovieNfo {
	director, _ := TakeOne(dto.Persons, func(item kinopoisk.Person) bool {
		return item.EnProfession == "director"
	})

	actors := Filter(dto.Persons, func(item kinopoisk.Person) bool {
		return item.EnProfession == "actor"
	})

	actorDto := Select(actors, func(item kinopoisk.Person) Actor {
		return Actor{
			Name:  flat(item.Name),
			Role:  flat(item.Description),
			Thumb: item.Photo,
		}
	})

	return MovieNfo{
		Title:         dto.Name,
		Originaltitle: dto.AlternativeName,
		Ratings: Ratings{
			Rating: []Rating{
				{
					Name:    "kinopoisk",
					Max:     "10",
					Value:   fmt.Sprintf("%f", dto.Rating.Kp),
					Default: "true",
					Votes:   fmt.Sprintf("%f", dto.Votes.Kp),
				},
				{
					Name:    "imdb",
					Max:     "10",
					Value:   fmt.Sprintf("%f", dto.Rating.Imdb),
					Default: "true",
					Votes:   fmt.Sprintf("%f", dto.Votes.Imdb),
				},
			},
		},
		Outline: dto.ShortDescription,
		Plot:    dto.Description,
		Tagline: dto.Slogan,
		Genre: strings.Join(Select(dto.Genres,
			func(item kinopoisk.Country) string { return item.Name }), ", "),
		Country:   Select(dto.Countries, func(item kinopoisk.Country) string { return item.Name }),
		Director:  *director.Name,
		Actor:     actorDto,
		Premiered: dto.Premiere.World.Format("2006-01-02"),
		// Thumb: []Thumb{
		// 	{
		// 		Aspect:  "poster",
		// 		Preview: dto.Poster.PreviewURL,
		// 		Text:    dto.Poster.URL,
		// 	},
		// },
	}
}

func TakeOne[T any](slice []T, selector func(T) bool) (T, bool) {
	var item T
	for _, item := range slice {
		if selector(item) {
			return item, true
		}
	}

	return item, false
}

func Filter[T any](slice []T, selector func(T) bool) []T {
	result := make([]T, 0)
	for _, item := range slice {
		if selector(item) {
			result = append(result, item)
		}
	}

	return result
}

func Select[T, V any](slice []T, f func(T) V) []V {
	result := make([]V, 0, len(slice))
	for _, item := range slice {
		result = append(result, f(item))
	}

	return result
}

func makePointer[T any](a T) *T {
	return &a
}

func flat(item *string) string {
	if item == nil {
		return ""
	}

	return *item
}

func downloadImage(url string, pathToSave string) {
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatal(err)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(pathToSave)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = jpeg.Encode(file, img, nil)
	if err != nil {
		log.Fatal(err)
	}
}

var whitelistRegex = regexp.MustCompile(`[^a-zA-Z\(\)0-9а-я-А-Я\ ]`)

func whitelistString(str string) string {
	return whitelistRegex.ReplaceAllString(str, "")
}
