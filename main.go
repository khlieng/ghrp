package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
)

const cacheTTL = 5 * time.Minute

var (
	client       *github.Client
	releaseCache = map[string]*release{}
)

type release struct {
	*github.RepositoryRelease
	fetchedAt time.Time
}

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("GITHUB_TOKEN is required")
		os.Exit(0)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	client = github.NewClient(tc)

	fmt.Println(":" + port)
	http.ListenAndServe(":"+port, http.HandlerFunc(serve))
}

func serve(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(params) == 3 {
		owner := params[0]
		repo := params[1]
		query := params[2]

		if rel, ok := releaseCache[owner+"/"+repo]; ok &&
			time.Since(rel.fetchedAt) < cacheTTL {
			proxy(w, r, rel, query)
		} else {
			fetchLatest(w, r, owner, repo, query)
		}
	} else {
		fail(w, http.StatusBadRequest)
	}
}

func fetchLatest(w http.ResponseWriter, r *http.Request, owner, repo, query string) {
	ghRel, _, err := client.Repositories.GetLatestRelease(context.Background(), owner, repo)
	if err != nil {
		fail(w, http.StatusInternalServerError)
	} else {
		rel := &release{
			RepositoryRelease: ghRel,
			fetchedAt:         time.Now(),
		}
		releaseCache[owner+"/"+repo] = rel

		proxy(w, r, rel, query)
	}
}

func proxy(w http.ResponseWriter, r *http.Request, rel *release, query string) {
	for _, asset := range rel.Assets {
		if strings.Contains(asset.GetName(), query) {
			url, err := url.Parse(asset.GetBrowserDownloadURL())
			if err != nil {
				fail(w, http.StatusInternalServerError)
			}

			proxy := &httputil.ReverseProxy{
				Director: func(r *http.Request) {
					r.URL = url
				},
			}
			proxy.ServeHTTP(w, r)
			return
		}
	}

	fail(w, http.StatusNotFound)
}

func fail(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
