package xsf

import (
	"regexp"
)

const lbTargetsRule = `^(([\w.\-|]+,){5}([\w.\-|]+))((;([\w.\-|]+,){5}([\w.\-|]+)))*$`

var lbTargetsRegexp *regexp.Regexp

func init() {
	lbTargetsRegexp = regexp.MustCompile(lbTargetsRule)
}

func checkLbTargets(lbTargets string) bool {
	return lbTargetsRegexp.MatchString(lbTargets)
}
