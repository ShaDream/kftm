package main

import "encoding/xml"

type Ratings struct {
	Text   string   `xml:",chardata"`
	Rating []Rating `xml:"rating"`
}

type Rating struct {
	Text    string `xml:",chardata"`
	Name    string `xml:"name,attr"`
	Max     string `xml:"max,attr"`
	Default string `xml:"default,attr"`
	Value   string `xml:"value"`
	Votes   string `xml:"votes"`
}

type Thumb struct {
	Text    string `xml:",chardata"`
	Spoof   string `xml:"spoof,attr"`
	Cache   string `xml:"cache,attr"`
	Aspect  string `xml:"aspect,attr"`
	Preview string `xml:"preview,attr"`
}

type Actor struct {
	Text  string `xml:",chardata"`
	Name  string `xml:"name"`
	Role  string `xml:"role"`
	Order string `xml:"order"`
	Thumb string `xml:"thumb"`
}

type MovieNfo struct {
	XMLName xml.Name `xml:"movie"`
	Text    string   `xml:",chardata"`
	// The title for the movie
	Title         string  `xml:"title"`
	Originaltitle string  `xml:"originaltitle"`
	Sorttitle     string  `xml:"sorttitle"`
	Ratings       Ratings `xml:"ratings"`
	// Should be short, will be displayed on a single line (scraped from IMDB only)
	Outline string `xml:"outline"`
	// Can contain more information on multiple lines, will be wrapped
	Plot string `xml:"plot"`
	// Short movie slogan. "The true story of a real fake" is the tagline for "Catch me if you can"
	Tagline string `xml:"tagline"`
	// Path to available Movie Posters. Not needed when using local artwork.
	//
	// Example use of aspect="":
	//  - <thumb aspect="banner"
	//  - <thumb aspect="clearart"
	//  - <thumb aspect="clearlogo"
	//  - <thumb aspect="discart"
	//  - <thumb aspect="keyart"
	//  - <thumb aspect="landscape"
	//  - <thumb aspect="poster"
	Thumb     []Thumb  `xml:"thumb"`
	Genre     string   `xml:"genre"`
	Country   []string `xml:"country"`
	Director  string   `xml:"director"`
	Premiered string   `xml:"premiered"`
	// Note: Kodi v17: Tag deprecated, use <premiered> tag instead. Note: Kodi v20: Use <premiered> tag only.
	Year   string  `xml:"year"`
	Studio string  `xml:"studio"`
	Actor  []Actor `xml:"actor"`
}
