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
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const invalidInput = "invalid input, please try again (type Ctrl+C to exit)"

func main() {

	app := cli.App("rtag", "rtag is a git tag management utility that helps making consistent release tags")

	app.Action = func() {
		wd, _ := os.Getwd()
		stats, err := git.Stat(wd)
		if err != nil {
			log.Fatalf("failed to obtain git stats: %v", err)
		}

		fmt.Println("repo info:")
		fmt.Println("- branch:", stats.Branch)
		fmt.Println("- author date:", stats.AuthorDate)
		fmt.Println("- hash:", stats.Hash)
		fmt.Println("- last tag:", stats.Description.Tag)

		if stats.Description.AdditionalCommits > 0 {
			fmt.Println("! additional commits:", stats.Description.AdditionalCommits)
		}

		if stats.Dirty {
			fmt.Println("! modified since last commit (use 'git status' for mode detail)")
		}

		vi, err := git.ParseVersion(stats.Description)
		check(err)

		fmt.Println("- semantic version:", vi.Semantic)
		fmt.Println("- version quad:", vi.Quad.String())

		fmt.Println()

		if stats.Dirty {
			fmt.Println("commit your changes before updating the tag")
			os.Exit(1)
		}

		if stats.Description.AdditionalCommits == 0 {
			fmt.Printf("your repo state is already tagged as '%s'\n", stats.Description.Tag)
			yn := ""
			for {
				fmt.Printf("do you still want to proceed (Y/N)?")
				fmt.Scanln(&yn)
				yn = strings.ToLower(yn)
				if yn == "n" || yn == "no" {
					fmt.Println("ok, exiting")
					os.Exit(2)
				} else if yn == "y" || yn == "yes" {
					break
				} else {
					fmt.Println(invalidInput)
				}
			}
		}

		actions := collectActions(vi.Semantic)
		if len(actions) > 0 {
			fmt.Println()
			fmt.Println("actions:")
			for i, a := range actions {
				if a.showPRchoice {
					fmt.Printf("%d: %s '%s' ...\n", i+1, a.desc, a.ver.String())
				} else {
					fmt.Printf("%d: %s '%s'\n", i+1, a.desc, a.ver.String())
				}
			}

			choice := 0
			for {
				fmt.Printf("make a choice, type 1 ... %d: ", len(actions))

				var s string
				fmt.Scanln(&s)

				if len(s) == 0 {
					fmt.Println(invalidInput)
					continue
				}
				v, err := strconv.Atoi(s)
				if err != nil || v < 1 || int(v) > len(actions) {
					fmt.Println(invalidInput)
					continue
				}
				choice = int(v)
				break
			}

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
				for {
					fmt.Printf("type 'alpha', 'beta', 'rc', or 'release': ")
					fmt.Scanln(&choice)

					if choice != "alpha" && choice != "beta" && choice != "rc" && choice != "release" {
						fmt.Println(invalidInput)
						continue
					}

					if choice == "release" {
						newver.Pre = newver.Pre[:0]
					} else {
						newver.Pre = makePR(choice, 1)
					}
					break
				}
			}

			comment := fmt.Sprintf("tagging as %s", newver.String())
			fmt.Println()
			fmt.Printf("ready to update current tag '%s' -> '%s'\n", vi.Semantic.String(), newver.String())
			fmt.Printf("will comment as: '%s'\n", comment)
			yn := ""
			for {
				fmt.Printf("proceed (Y/N)? ")
				fmt.Scanln(&yn)
				yn = strings.ToLower(yn)
				if yn == "n" || yn == "no" {
					fmt.Println("ok, exiting without changes")
					os.Exit(2)
				} else if yn == "y" || yn == "yes" {
					break
				} else {
					fmt.Println(invalidInput)
				}
			}

			fmt.Println()
			fmt.Printf("executing: 'git tag -a %s -m \"%s\"\n", newver.String(), comment)

			cmd := exec.Command("git", "tag", "-a", newver.String(), "-m", comment)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				os.Exit(1)
			}

			fmt.Println()
			fmt.Printf("your local repo is now tagged as '%s'\n", newver.String())
			fmt.Printf("you can push these changes to the remote now\n")
			fmt.Printf("or do additional testing, then:\n")
			fmt.Printf("- execute 'git push --tags' to push\n")
			fmt.Printf("- execute 'git tag -d %s' to undo\n", newver.String())
			fmt.Println()
			fmt.Printf("rolling back changes after the push requires one mode step:\n")
			fmt.Printf("- execute 'git tag -d %s' to delete the local tag\n", newver.String())
			fmt.Printf("- execute 'git push --delete origin %s' to delete the remote tag\n", newver.String())
			fmt.Println()

			for {
				fmt.Print("push the new tag to the remote repository (Y/N)?")
				fmt.Scanln(&yn)
				yn = strings.ToLower(yn)
				if yn == "n" || yn == "no" {
					fmt.Println("ok, exiting")
					os.Exit(0)
				} else if yn == "y" || yn == "yes" {
					break
				} else {
					fmt.Println(invalidInput)
				}
			}

			fmt.Println("executing: 'git push -tags'")

			cmd = exec.Command("git", "push", "--tags")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				os.Exit(1)
			}

			fmt.Printf("mission accomplished")
		}
	}

	app.Run(os.Args)
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
