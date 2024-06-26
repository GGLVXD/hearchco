package googleimages

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/hearchco/hearchco/src/anonymize"
	"github.com/hearchco/hearchco/src/config"
	"github.com/hearchco/hearchco/src/search/bucket"
	"github.com/hearchco/hearchco/src/search/engines"
	"github.com/hearchco/hearchco/src/search/engines/_sedefaults"
	"github.com/rs/zerolog/log"
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

	// disable User Agent since Google Images responds with fake data if UA is correct
	col.UserAgent = ""

	pageRankCounter := make([]int, options.Pages.Max)

	col.OnResponse(func(e *colly.Response) {
		body := string(e.Body)
		index := strings.Index(body, "{\"ischj\":")

		if index == -1 {
			log.Error().
				Caller().
				Str("engine", Info.Name.String()).
				Str("body", body).
				Msg("Failed parsing response, couldn't find the start of JSON")
			return
		}

		body = body[index:]
		var jsonResponse JsonResponse
		if err := json.Unmarshal([]byte(body), &jsonResponse); err != nil {
			log.Error().
				Caller().
				Err(err).
				Str("engine", Info.Name.String()).
				Str("body", body).
				Msg("Failed parsing response, couldn't unmarshal JSON")
			return
		}

		pageIndex := _sedefaults.PageFromContext(e.Request.Ctx, Info.Name)
		page := pageIndex + options.Pages.Start + 1

		for _, metadata := range jsonResponse.ISCHJ.Metadata {
			origImg := metadata.OriginalImage
			thmbImg := metadata.Thumbnail
			resultJson := metadata.Result
			textInGridJson := metadata.TextInGrid

			// google images sometimes inverts original height and width
			if (thmbImg.Height > thmbImg.Width) != (origImg.Height > origImg.Width) {
				origImg.Height, origImg.Width = origImg.Width, origImg.Height
			}

			if resultJson.ReferrerUrl != "" && origImg.Url != "" && thmbImg.Url != "" {
				res := bucket.MakeSEImageResult(
					origImg.Url, resultJson.PageTitle, textInGridJson.Snippet,
					resultJson.SiteTitle, resultJson.ReferrerUrl, thmbImg.Url,
					origImg.Height, origImg.Width, thmbImg.Height, thmbImg.Width,
					Info.Name, page, pageRankCounter[pageIndex]+1,
					salt,
				)
				valid := bucket.AddSEResult(&res, Info.Name, relay, options, pagesCol, nEnabledEngines)
				if valid {
					pageRankCounter[pageIndex]++
				}
			} else {
				log.Error().
					Caller().
					Str("engine", Info.Name.String()).
					Str("jsonMetadata", fmt.Sprintf("%v", metadata)).
					Str("url", resultJson.ReferrerUrl).
					Str("original", origImg.Url).
					Str("thumbnail", thmbImg.Url).
					Msg("Couldn't find image URL")
			}
		}
	})

	retErrors := make([]error, 0, options.Pages.Max)

	// starts from at least 0
	for i := options.Pages.Start; i < options.Pages.Start+options.Pages.Max; i++ {
		colCtx := colly.NewContext()
		colCtx.Put("page", strconv.Itoa(i-options.Pages.Start))

		// dynamic params
		pageParam := "&tbm=isch&asearch=isch&async=_fmt:json,p:1,ijn:1"
		// i == 0 is the first page
		if i > 0 {
			pageParam = "&tbm=isch&asearch=isch&async=_fmt:json,p:1,ijn:" + strconv.Itoa(i*10)
		}

		urll := Info.URL + query + pageParam
		anonUrll := Info.URL + anonymize.String(query) + pageParam

		err := _sedefaults.DoGetRequest(urll, anonUrll, colCtx, col, Info.Name)
		if err != nil {
			retErrors = append(retErrors, err)
		}
	}

	col.Wait()
	pagesCol.Wait()

	return retErrors[:len(retErrors):len(retErrors)]
}
