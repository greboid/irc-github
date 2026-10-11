package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/sebdah/goldie/v2"
)

func Test_github_handlePackageEvent(t *testing.T) {
	tests := []string{"packages/published_1.json", "packages/deleted_1.json"}
	gold := goldie.New(t)
	for index := range tests {
		t.Run(tests[index], func(t *testing.T) {
			g := &githubPackageHandler{}
			hook := webhook{}
			err := getTestData(tests[index], &hook)
			if err != nil {
				t.Fatal("Unable to parse example data")
			}
			got := []byte(strings.Join(g.handlePackageEvent(hook), "\n"))
			gold.Assert(t, tests[index], got)
		})
	}
}

func Test_github_handlePackageEvent_unhandledAction(t *testing.T) {
	g := &githubPackageHandler{}
	got := g.handlePackageEvent(webhook{Action: "updated"})
	if len(got) != 0 {
		t.Errorf("handlePackageEvent() = %v, want no messages", got)
	}
}

func Test_github_handlePackageEvent_nilPackage(t *testing.T) {
	g := &githubPackageHandler{}
	got := g.handlePackageEvent(webhook{Action: "published"})
	if len(got) != 0 {
		t.Errorf("handlePackageEvent() = %v, want no messages", got)
	}
}

func Test_github_handlePackageEvent_noRepositoryFallsBackToOwner(t *testing.T) {
	g := &githubPackageHandler{}
	got := g.handlePackageEvent(webhook{
		Action: "created",
		Package: &Package{
			Name:    "myimage",
			Type:    "container",
			Version: "v1.0.0",
			Owner:   &User{Login: "testuser"},
			HtmlURL: "https://forgejo.example/testuser/-/packages/container/myimage/v1.0.0",
		},
		Sender: Sender{Login: "testuser"},
	})
	want := []string{"[testuser] testuser published container package myimage:v1.0.0 - https://forgejo.example/testuser/-/packages/container/myimage/v1.0.0"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("handlePackageEvent() = %v, want %v", got, want)
	}
}

func Test_github_handlePackageEvent_noSenderFallsBackToCreator(t *testing.T) {
	g := &githubPackageHandler{}
	got := g.handlePackageEvent(webhook{
		Action:     "created",
		Repository: Repository{FullName: "testuser/repo"},
		Package: &Package{
			Name:    "mymodule",
			Type:    "go",
			Version: "v1.2.3",
			Creator: &User{Login: "testuser"},
			HtmlURL: "https://forgejo.example/testuser/repo/packages/go/mymodule/v1.2.3",
		},
	})
	want := []string{"[testuser/repo] testuser published go package mymodule:v1.2.3 - https://forgejo.example/testuser/repo/packages/go/mymodule/v1.2.3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("handlePackageEvent() = %v, want %v", got, want)
	}
}

func Test_github_handlePackageEvent_noVersion(t *testing.T) {
	g := &githubPackageHandler{}
	got := g.handlePackageEvent(webhook{
		Action:     "published",
		Repository: Repository{FullName: "testuser/repo"},
		Sender:     Sender{Login: "testuser"},
		Package: &Package{
			Name:    "myimage",
			Type:    "container",
			HtmlURL: "https://forgejo.example/testuser/-/packages/container/myimage",
		},
	})
	want := []string{"[testuser/repo] testuser published container package myimage - https://forgejo.example/testuser/-/packages/container/myimage"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("handlePackageEvent() = %v, want %v", got, want)
	}
}
