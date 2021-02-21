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

var ui = &input.UI{
	Writer: os.Stdout,
	Reader: os.Stdin,
}

func main() {
	prefix := "AUTO"
	allowDirty := false

	app := cli.App("rtag", "rtag is a git tag management utility that helps making consistent release tags")

	app.Spec = "[--prefix=<ver-prefix>] [--allow-dirty]"

	app.StringOptPtr(&prefix, "p prefix", "AUTO", "prefix for new tags")
	app.BoolOptPtr(&allowDirty, "allow-dirty", false, "allow tagging of repos that contain uncommited chanves")

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
			if prefix == "AUTO" {
				prefix = "v"
			}
			tag := prefix + "0.0.1"
			comment := fmt.Sprintf("tagging as %s", ppTag(tag))

			fmt.Println()
			fmt.Println("the repo does not yet have any tags assigned")
			fmt.Println("you can assign the first tag manually, for example:")
			fmt.Println()
			fmt.Printf("    git tag -a %s -m \"%s\"\n", tag, comment)
			fmt.Println()
			fmt.Println("or this utility can do it for you")
			fmt.Println()
			fmt.Printf("ready create first tag %s %s:\n", ppTag(tag), ppDim("(with comment '"+comment+"')"))
			err := performTagging(tag, comment)
			check(err)
			return
		} else if err != nil {
			log.Fatalf("failed to obtain git stats: %v", err)
		}

		if stats.Dirty {
			fmt.Println("! modified since last commit (use 'git status' for mode detail)")
			fmt.Println()
			fmt.Println("commit your changes before updating the tag")
			os.Exit(1)
		}

		oldtag := stats.Description.Tag
		vi, err := git.ParseVersion(stats.Description)
		if err != nil {
			var altOldTag string
			altOldTag, vi, err = git.LastSemanticTag(wd)
			if err == nil {
				fmt.Printf("last tag '%s' does not conform to semantic version syntax\n", oldtag)
				fmt.Printf("however, there is an older tag '%s' that can be used instead\n", altOldTag)
				fmt.Println()
				if !askYN("proceed with '" + altOldTag + "' as base " + ppInp("/", "y", "n") + "? ") {
					fmt.Println("exiting")
					os.Exit(2)
				}
				oldtag = altOldTag
			} else {
				fmt.Println()
				fmt.Printf("ERROR: last tag '%s' does not conform to semantic version syntax\n", stats.Description.Tag)
				fmt.Println("this utility expects your repository to be tagged with semantic tags")
				fmt.Println("see https://semver.org for more information")
				fmt.Println("exiting now")
				os.Exit(1)
			}
		}

		fmt.Println("- last tag:    ", oldtag)
		if prefix == "AUTO" {
			prefix = "v"
			if len(oldtag) > 1 {
				d := strings.IndexAny(oldtag, "0123456789")
				if d == 0 {
					prefix = ""
				} else if d == 1 && oldtag[0] == 'v' || oldtag[0] == 'V' {
					prefix = oldtag[:1]
				}
			}

			if prefix == "" {
				fmt.Printf("- auto prefix:  no\n")
			} else {
				fmt.Printf("- auto prefix:  with %q\n", prefix)
			}
		}

		stats.Description.Tag = strings.TrimPrefix(stats.Description.Tag, prefix)

		if stats.Description.AdditionalCommits > 0 {
			fmt.Println("- additional commits:", stats.Description.AdditionalCommits)
		}

		fmt.Println("- semantic ver:", vi.Semantic)
		fmt.Println("- version quad:", vi.Quad.String())

		fmt.Println()

		if stats.Description.AdditionalCommits == 0 {
			fmt.Printf("your repo state is already tagged as %s\n", ppTag(oldtag))
			fmt.Print("still want to ")
			if !askYN("proceed " + ppInp("/", "y", "n") + "? ") {
				fmt.Println("exiting")
				os.Exit(2)
			}
		}

		actions := collectActions(vi.Semantic)
		if len(actions) > 0 {
			fmt.Println()
			fmt.Println("available actions:")
			for i, a := range actions {
				desc := a.desc
				postdesc := ""
				if i := strings.IndexByte(desc, '|'); i >= 0 {
					postdesc = " " + ppDim("("+desc[i+1:]+")")
					desc = desc[:i]
				}
				if a.showPRchoice {
					fmt.Printf("%d: %s %s%s ...\n", i+1, desc, ppTag(prefix+a.ver.String()), postdesc)
				} else {
					fmt.Printf("%d: %s %s%s\n", i+1, desc, ppTag(prefix+a.ver.String()), postdesc)
				}
			}

			choice := 0
			fmt.Print("make a choice: ")
			ask("type a number "+ppInp("...", "1", strconv.Itoa(len(actions)))+": ", func(s string) bool {
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
				fmt.Printf("- 'alpha'   for %s\n", ppTag(prefix+withPR(action.ver, "alpha", 1).String()))
				fmt.Printf("- 'beta'    for %s\n", ppTag(prefix+withPR(action.ver, "beta", 1).String()))
				fmt.Printf("- 'rc'      for %s\n", ppTag(prefix+withPR(action.ver, "rc", 1).String()))
				fmt.Printf("- 'release' for %s\n", ppTag(prefix+withoutPR(action.ver).String()))

				choice := ""
				ask(fmt.Sprintf("type %s: ", ppInp("/", "alpha", "beta", "rc", "release")),
					func(s string) bool {
						choice = s
						if choice == "alpha" || choice == "beta" || choice == "rc" || choice == "release" {
							return true
						}
						fmt.Println("invalid input, please try again (Ctrl+C to exit)")
						return false
					})
				if choice != "release" {
					newver.Pre = makePR(choice, 1)
				} else {
					newver.Pre = newver.Pre[:0]
				}
			}

			tag := prefix + newver.String()
			comment := fmt.Sprintf("tagging as %s", tag)
			fmt.Println()
			fmt.Printf("ready to update tag %s->%s %s:\n", ppTag(oldtag), ppTag(tag), ppDim("(with comment '"+comment+"')"))
			err := performTagging(tag, comment)
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
		fmt.Println("invalid input, please try again (Ctrl+C to exit)")
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

	if !askYN("proceed " + ppInp("/", "y", "n") + "? ") {
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
	fmt.Printf("your local repo is now tagged as %s\n", ppTag(tag))
	fmt.Printf("- to push it to remote:  'git push origin %s'\n", tag)
	fmt.Printf("- to undo local changes: 'git tag -d %s'\n", tag)
	fmt.Println()
	fmt.Printf("rolling back changes after the push:\n")
	fmt.Printf("- delete local:  'git tag -d %s'\n", tag)
	fmt.Printf("- delete remote: 'git push --delete origin %s'\n", tag)
	fmt.Println()
	fmt.Println("this utility can push the new tag for you")
	if !askYN("proceed with push " + ppInp("/", "y", "n") + "? ") {
		fmt.Println("exiting")
		os.Exit(2)
	}

	fmt.Printf("executing: 'git push origin %s'\n", tag)

	cmd = exec.Command("git", "push", "origin", tag)
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
			ret = append(ret, action{"increment patch|backwards compatible bug fixes", n, true})
		}
		if n := v; n.IncrementMinor() == nil {
			ret = append(ret, action{"increment minor|backwards compatible new functionality", n, true})
		}
		if n := v; n.IncrementMajor() == nil {
			ret = append(ret, action{"increment major|incompatible API changes", n, true})
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

var pretty = true

const (
	vtReset     = "\x1b[0m"
	vtBold      = "\x1b[1m"
	vtDim       = "\x1b[2m"
	vtItalic    = "\x1b[3m"
	vtUnderline = "\x1b[4m"
)

func ppTag(s string) string {
	if pretty {
		return fmt.Sprintf(vtUnderline+"%s"+vtReset, s)
	} else {
		return fmt.Sprintf("'%s'", s)
	}
}

func ppDim(s string) string {
	if pretty {
		return fmt.Sprintf(vtDim+"%s"+vtReset, s)
	}
	return s
}

func ppInp(sep string, opts ...string) string {
	//if pretty {
	//	return vtDim + "[" + vtReset + vtBold + strings.Join(opts, vtReset+vtDim+sep+vtReset+vtBold) + vtReset + vtDim + "]" + vtReset
	//}
	return "[" + strings.Join(opts, sep) + "]"
}
