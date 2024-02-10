package mojeek

import "github.com/hearchco/hearchco/src/search/engines"

var Info engines.Info = engines.Info{
	Domain:         "www.mojeek.com",
	Name:           engines.MOJEEK,
	URL:            "https://www.mojeek.com/search?q=",
	ResultsPerPage: 10,
}

var dompaths engines.DOMPaths = engines.DOMPaths{
	Result:      "ul.results-standard > li",
	Link:        "h2 > a.title",
	Title:       "h2 > a.title",
	Description: "p.s",
}

var Support engines.SupportedSettings = engines.SupportedSettings{
	Locale:     true,
	SafeSearch: true,
}
