package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"strings"

	realGithub "github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"

	"github.com/zegl/bazel_dependency_tools/internal/github"
	"github.com/zegl/bazel_dependency_tools/licenses"
	"github.com/zegl/bazel_dependency_tools/maven_jar"
	"github.com/zegl/bazel_dependency_tools/parse"
)

func main() {
	flagPrefixFilter := flag.String("prefix", "", "Only attempt to upgrade dependencies with this prefix, if prefix is empty (default) all dependencies will be upgraded")
	flagWorkspace := flag.String("workspace", "WORKSPACE", "Path to the WORKSPACE file")
	flagFindLicenses := flag.Bool("find-licenses", false, "Runin find licenses mode")
	flag.Parse()

	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	gitHubClient := github.NewGithubClient(realGithub.NewClient(tc))

	if *flagFindLicenses {
		findLicenses(*flagWorkspace, *flagPrefixFilter)
		return
	}

	lineReplacements := parse.ParseWorkspace(*flagWorkspace, *flagPrefixFilter, gitHubClient, maven_jar.NewestAvailable)

	rawContent, err := ioutil.ReadFile(*flagWorkspace)
	if err != nil {
		panic(err)
	}

	rows := strings.Split(string(rawContent), "\n")

	// Perform all replacements
	for _, r := range lineReplacements {
		rows[r.Line-1] = strings.Replace(rows[r.Line-1], r.Find, r.Substitution, -1)
	}

	// Write the new file
	err = ioutil.WriteFile(*flagWorkspace, []byte(strings.Join(rows, "\n")), 0777)
	if err != nil {
		panic(err)
	}
}

func findLicenses(workspace, prefixFilter string) {
	licenses.ParseWorkspace(workspace, prefixFilter)
}
