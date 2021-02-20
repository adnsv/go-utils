package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/adnsv/go-utils/git"
	"github.com/adnsv/go-utils/version"
	"github.com/blang/semver/v4"
	cli "github.com/jawher/mow.cli"
	"github.com/tcnksm/go-input"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const invalidInput = "invalid input, please try again (Ctrl+C to exit)"

var ui = &input.UI{
	Writer: os.Stdout,
	Reader: os.Stdin,
}

func main() {

	app := cli.App("rtag", "rtag is a git tag management utility that helps making consistent release tags")

	app.Action = func() {

		wd, _ := os.Getwd()
		stats, err := git.Stat(wd)

		if stats != nil {
			fmt.Println("repo info:")
			fmt.Println("- branch:      ", stats.Branch)
			fmt.Println("- author date: ", stats.AuthorDate)
			fmt.Println("- hash:        ", stats.Hash)
		}

		if err == git.ErrNoTags {
			tag := "v0.0.1"
			comment := "tagging as v0.0.1"

			fmt.Println()
			fmt.Println("the repo does not yet have any tags assigned")
			fmt.Println("you can assign the first tag manually, for example:")
			fmt.Println()
			fmt.Printf("    git tag -a %s -m \"%s\"\n", tag, comment)
			fmt.Println()
			fmt.Println("or this utility can do it for you")
			fmt.Println()
			fmt.Printf("ready create first tag '%s' with comment '%s'\n", tag, comment)
			err := performTagging(tag, comment)
			check(err)
			return
		} else if err != nil {
			log.Fatalf("failed to obtain git stats: %v", err)
		}

		oldtag := stats.Description.Tag

		fmt.Println("- last tag:    ", oldtag)
		vprefix := len(oldtag) > 0 && oldtag[0] == 'v'

		if stats.Description.AdditionalCommits > 0 {
			fmt.Println("! additional commits:", stats.Description.AdditionalCommits)
		}

		if stats.Dirty {
			fmt.Println("! modified since last commit (use 'git status' for mode detail)")
		}

		vi, err := git.ParseVersion(stats.Description)
		if err != nil {
			fmt.Println()
			fmt.Printf("ERROR: last tag '%s' does not conform to semantic version syntax\n", stats.Description.Tag)
			fmt.Println("this utility expects your repository to be tagged with semantic tags")
			fmt.Println("see https://semver.org for more information")
			fmt.Println("exiting now")
			os.Exit(1)
		}

		fmt.Println("- semantic ver:", vi.Semantic)
		fmt.Println("- version quad:", vi.Quad.String())

		fmt.Println()

		if stats.Dirty {
			fmt.Println("commit your changes before updating the tag")
			os.Exit(1)
		}

		if stats.Description.AdditionalCommits == 0 {
			fmt.Printf("your repo state is already tagged as '%s'\n", stats.Description.Tag)
			fmt.Print("still want to ")
			if !askYN("proceed [y/n]? ") {
				fmt.Println("exiting")
				os.Exit(2)
			}
		}

		actions := collectActions(vi.Semantic)
		if len(actions) > 0 {
			fmt.Println()
			fmt.Println("available actions:")
			for i, a := range actions {
				if a.showPRchoice {
					fmt.Printf("%d: %s '%s' ...\n", i+1, a.desc, a.ver.String())
				} else {
					fmt.Printf("%d: %s '%s'\n", i+1, a.desc, a.ver.String())
				}
			}

			choice := 0
			fmt.Print("make a choice: ")
			ask(fmt.Sprintf("type a number [1 ... %d]: ", len(actions)), func(s string) bool {
				v, err := strconv.Atoi(s)
				choice = int(v)
				if err == nil && choice >= 1 && choice <= len(actions) {
					return true
				}
				return false
			})

			action := actions[choice-1]
			newver := action.ver

			if action.showPRchoice {
				fmt.Println()
				fmt.Println("select (pre-)release type")
				fmt.Printf("- 'alpha'   for '%s'\n", withPR(action.ver, "alpha", 1).String())
				fmt.Printf("- 'beta'    for '%s'\n", withPR(action.ver, "beta", 1).String())
				fmt.Printf("- 'rc'      for '%s'\n", withPR(action.ver, "rc", 1).String())
				fmt.Printf("- 'release' for '%s'\n", withoutPR(action.ver).String())

				choice := ""
				ask("type 'alpha', 'beta', 'rc', or 'release': ", func(s string) bool {
					choice = s
					if choice == "alpha" || choice == "beta" || choice == "rc" || choice == "release" {
						return true
					}
					fmt.Println("invalid input, please try again (Ctrl+C to exit)")
					return false
				})
			}

			tag := newver.String()
			if vprefix {
				tag = "v" + tag
			}
			comment := fmt.Sprintf("tagging as %s", tag)
			fmt.Println()
			fmt.Printf("ready to update tag '%s'->'%s' with comment '%s'\n", oldtag, tag, comment)
			err := performTagging(newver.String(), comment)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("mission accomplished")
		}
	}

	app.Run(os.Args)
}

func ask(prompt string, validate func(s string) bool) {
	s := ""
	for {
		fmt.Printf(prompt)
		fmt.Scanln(&s)
		if validate(s) {
			break
		}
		fmt.Println(invalidInput)
		fmt.Println()
	}
}

func askYN(prompt string) bool {
	ret := false
	ask(prompt, func(s string) bool {
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			ret = true
			return true
		} else if s == "n" || s == "no" {
			ret = false
			return true
		}
		return false
	})
	return ret
}

func performTagging(tag string, comment string) error {

	if !askYN("proceed [y/n]? ") {
		fmt.Println("exiting without changes")
		os.Exit(2)
	}

	fmt.Println()
	fmt.Printf("executing: 'git tag -a %s -m \"%s\"\n", tag, comment)

	cmd := exec.Command("git", "tag", "-a", tag, "-m", comment)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("your local repo is now tagged as '%s'\n", tag)
	fmt.Printf("- to push it to remote:  'git push --tags'\n")
	fmt.Printf("- to undo local changes: 'git tag -d %s'\n", tag)
	fmt.Println()
	fmt.Printf("rolling back changes after the push:\n")
	fmt.Printf("- delete local:  'git tag -d %s'\n", tag)
	fmt.Printf("- delete remote: 'git push --delete origin %s'\n", tag)
	fmt.Println()
	fmt.Println("this utility can push the new tag for you")
	if !askYN("proceed with push [y/n]? ") {
		fmt.Println("exiting")
		os.Exit(2)
	}

	fmt.Println("executing: 'git push -tags'")

	cmd = exec.Command("git", "push", "--tags")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}

type action struct {
	desc         string
	ver          semver.Version
	showPRchoice bool
}

func makePR(s string, n uint64) []semver.PRVersion {
	return []semver.PRVersion{
		{VersionStr: s},
		{VersionNum: n, IsNum: true},
	}
}

func withPR(v version.Semantic, pr string, pn uint64) version.Semantic {
	v.Pre = makePR(pr, pn)
	return v
}

func withoutPR(v version.Semantic) version.Semantic {
	v.Pre = v.Pre[:0]
	return v
}

func collectActions(v version.Semantic) []action {
	v.Build = nil
	ret := []action{}

	if len(v.Pre) == 0 {
		if n := v; n.IncrementPatch() == nil {
			ret = append(ret, action{"increment patch", n, true})
		}
		if n := v; n.IncrementMinor() == nil {
			ret = append(ret, action{"increment minor", n, true})
		}
		if n := v; n.IncrementMajor() == nil {
			ret = append(ret, action{"increment major", n, true})
		}
	} else {
		pr := v.Pre[0].VersionStr
		pn := uint64(0)
		if len(v.Pre) > 1 && v.Pre[1].IsNum {
			pn = v.Pre[1].VersionNum
		}
		ret = append(ret, action{"bump '" + pr + "'", withPR(v, pr, pn+1), false})
		if pr == "alpha" {
			ret = append(ret, action{"upgrade 'alpha' to 'beta'", withPR(v, "beta", 1), false})
			ret = append(ret, action{"upgrade 'alpha' to 'rc'", withPR(v, "rc", 1), false})
			// do not allow to go from alpha to release
		}
		if pr == "beta" {
			ret = append(ret, action{"upgrade 'beta' to 'rc'", withPR(v, "rc", 1), false})
			ret = append(ret, action{"make release", withoutPR(v), false})
		}
		if pr == "rc" {
			ret = append(ret, action{"make release", withoutPR(v), false})
		}
	}

	return ret
}
