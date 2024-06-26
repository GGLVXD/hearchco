package mojeek

import (
	"context"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/hearchco/hearchco/src/anonymize"
	"github.com/hearchco/hearchco/src/config"
	"github.com/hearchco/hearchco/src/search/bucket"
	"github.com/hearchco/hearchco/src/search/engines"
	"github.com/hearchco/hearchco/src/search/engines/_sedefaults"
)

type Engine struct{}

func New() Engine {
	return Engine{}
}

func (e Engine) Search(ctx context.Context, query string, relay *bucket.Relay, options engines.Options, settings config.Settings, timings config.CategoryTimings, salt string, nEnabledEngines int) []error {
	ctx, err := _sedefaults.Prepare(ctx, Info, Support, options, settings)
	if err != nil {
		return []error{err}
	}

	col, pagesCol := _sedefaults.InitializeCollectors(ctx, Info.Name, options, settings, timings, relay)

	pageRankCounter := make([]int, options.Pages.Max)

	col.OnHTML(dompaths.Result, func(e *colly.HTMLElement) {
		linkText, titleText, descText := _sedefaults.FieldsFromDOM(e.DOM, dompaths, Info.Name)

		pageIndex := _sedefaults.PageFromContext(e.Request.Ctx, Info.Name)
		page := pageIndex + options.Pages.Start + 1

		res := bucket.MakeSEResult(linkText, titleText, descText, Info.Name, page, pageRankCounter[pageIndex]+1)
		valid := bucket.AddSEResult(&res, Info.Name, relay, options, pagesCol, nEnabledEngines)
		if valid {
			pageRankCounter[pageIndex]++
		}
	})

	retErrors := make([]error, 0, options.Pages.Max)

	// static params
	localeParam := getLocale(options)
	safeSearchParam := getSafeSearch(options)

	// starts from at least 0
	for i := options.Pages.Start; i < options.Pages.Start+options.Pages.Max; i++ {
		colCtx := colly.NewContext()
		colCtx.Put("page", strconv.Itoa(i-options.Pages.Start))

		// dynamic params
		pageParam := ""
		// i == 0 is the first page
		if i > 0 {
			pageParam = "&s=" + strconv.Itoa(i*10+1)
		}

		urll := Info.URL + query + pageParam + localeParam + safeSearchParam
		anonUrll := Info.URL + anonymize.String(query) + pageParam + localeParam + safeSearchParam

		err := _sedefaults.DoGetRequest(urll, anonUrll, colCtx, col, Info.Name)
		if err != nil {
			retErrors = append(retErrors, err)
		}
	}

	col.Wait()
	pagesCol.Wait()

	return retErrors[:len(retErrors):len(retErrors)]
}

func getLocale(options engines.Options) string {
	spl := strings.SplitN(strings.ToLower(options.Locale), "_", 2)
	return "&lb=" + spl[0] + "&arc=" + spl[1]
}

func getSafeSearch(options engines.Options) string {
	if options.SafeSearch {
		return "&safe=1"
	}
	return "&safe=0"
}
