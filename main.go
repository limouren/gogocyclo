package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/token"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"code.google.com/p/gcfg"
)

const usageDoc = `Ignore cyclomatic analyse result from gocyclo.

Intended to be used as a step in CI to exclude functions that is known and
expected to have high cyclomatic complexity (e.g. function with big switches).
Read gocyclo output from stdin.

Usage:
  ogocylo [flags]

Flags:
  -config PATH   read configuration from PATH

Example:
  gocyclo -over 20 $GOPATH/src/database | gogocyclo -config gogocyclo.ini

Configuration file:
gogocyclo reads from gitconfig section named "gogocyclo". It consists of a
mutli-value key called "ignores" and take the following format:

  ` + "`Package`" + `.Func

where "Package" and "Func" is the second and third column of gocyclo output
respectively.

"Func" also accepts the wildcard character "*", which matches functions of
any names.
`

func usage() {
	fmt.Fprintln(os.Stderr, usageDoc)
	os.Exit(2)
}

var (
	configPath = flag.String("config", ".gogocyclo", "path to configuration file")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("gogocyclo: ")

	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) > 0 {
		log.Fatal("expect zero positional arguments")
	}

	config, err := configFromFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	var stats []statistic

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		stat := statistic{}
		line := scanner.Text()
		if err := stat.FromLine(line); err != nil {
			log.Fatal(err)
		}
		stats = append(stats, stat)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	var filteredStats []statistic
	for _, stat := range stats {
		if !config.Ignores.Match(&stat) {
			filteredStats = append(filteredStats, stat)
		}
	}

	if len(filteredStats) > 0 {
		for _, stat := range filteredStats {
			fmt.Println(stat)
		}
		os.Exit(1)
	}
}

// `Package`.Func
//
// NOTE(limouren): Wanted to use "Package".Func but gcfg remove double quotes
// automatically...
var ignorePattern = regexp.MustCompile("\\A`(.+)`.(.+)\\z")

func configFromFile(path string) (config *configuration, err error) {
	var gconfig gitconfig
	if err = gcfg.ReadFileInto(&gconfig, path); err != nil {
		return
	}

	config = &configuration{}
	for _, rawIgnore := range gconfig.GoGoCyclo.Ignores {
		submatches := ignorePattern.FindAllStringSubmatch(rawIgnore, -1)

		// first item being the matched string, the other two are what is in
		// the parenthesis
		if len(submatches) != 1 || len(submatches[0]) != 3 {
			err = fmt.Errorf("parse %s: ignores not in format `Package`.Func", rawIgnore)
			return
		}

		config.Ignores = append(config.Ignores, ignoreRule{
			PackageName: submatches[0][1],
			FuncName:    submatches[0][2],
		})
	}

	return
}

type configuration struct {
	Ignores ignoreRules
}

type ignoreRules []ignoreRule

func (rules ignoreRules) Match(stat *statistic) bool {
	for _, rule := range rules {
		if rule.Match(stat) {
			return true
		}
	}

	return false
}

type ignoreRule struct {
	PackageName string
	FuncName    string
}

func (ignore ignoreRule) Match(stat *statistic) bool {
	if strings.HasPrefix(stat.PackageName, ignore.PackageName) {
		if ignore.FuncName == "*" {
			return true
		}

		return stat.FuncName == ignore.FuncName
	}

	return false
}

type gitconfig struct {
	GoGoCyclo struct {
		Ignores []string
	}
}

type statistic struct {
	PackageName string
	FuncName    string
	Complexity  int
	Pos         token.Position
}

func (stat *statistic) FromLine(str string) (err error) {
	ss := strings.Split(str, " ")

	stat.Complexity, err = strconv.Atoi(ss[0])
	if err != nil {
		return
	}

	stat.PackageName = ss[1]
	stat.FuncName = ss[2]
	err = (*position)(&stat.Pos).FromText(ss[3])

	return
}

func (s statistic) String() string {
	return fmt.Sprintf("%d %s %s %s", s.Complexity, s.PackageName, s.FuncName, s.Pos)
}

type position token.Position

func (p *position) FromText(s string) (err error) {
	ss := strings.Split(s, ":")

	p.Filename = ss[0]

	p.Offset, err = strconv.Atoi(ss[1])
	if err != nil {
		return
	}
	p.Line, err = strconv.Atoi(ss[2])
	if err != nil {
		return
	}

	return
}
