package main

import (
	"flag"
	"strings"

	"github.com/jayacarlson/cfg"
	"github.com/jayacarlson/dbg"
	"github.com/jayacarlson/txt"
)

func handleArgs(a, p string) {
	switch a {
	case "D":
		hiddenDirs = true
	case "F":
		hiddenFiles = true
	case "h":
		dontHomify = true
	case "I":
		ignoreECase = true
	case "r":
		recursive = true
	case "b":
		bashHeader = true
	case "s":
		reverse = true
	case "o":
		outputFile = p
	case "i":
		include = p
	case "x":
		exclude = p
	case "f":
		fileOutput = p
	case "H":
		leadOutput = p
	case "T":
		tailOutput = p
	case "l":
		aLeadOutput = p
	case "t":
		aTailOutput = p
	case "d":
		dirOutput = p
	}
}

func processArgs(t string, f func(a, p string)) {
	t = strings.TrimSpace(t)
	if "" == t {
		return
	}
	dbg.ChkTruX(0 == (1&strings.Count(t, "\"")), "Mismatch in quotes")
	dbg.ChkTruX('"' != t[0], "Missing flag")
	p := strings.Split(t, "\"")
	for i, a := range p {
		if i >= len(p) {
			break
		}
		a = txt.CleanSpaces(a)
		if "" == a {
			continue
		}
		s := strings.Split(a, " ")
		for n, e := range s {
			if n >= len(s) {
				break
			}
			if "" == e {
				continue
			}
			dbg.ChkTruX('-' == e[0], "Invalid argument")
			e = strings.TrimLeft(e, "-")
			l := flag.Lookup(e)
			dbg.ChkTruX(nil != l, "Unknown argument flag: %s", e)

			switch l.Usage {
			case "error":
				dbg.Fatal("Cannot use arg '%s' in config params", e)
			case "bool":
				f(e, "")
			case "string":
				if n == len(s)-1 {
					dbg.ChkTruX(i < len(p)-1, "Missing argument for '%s'", e)
					f(e, p[i+1])
					p = append(p[:i+1], p[i+2:]...)
				} else {
					f(e, s[n+1])
					s = append(s[:i+1], s[i+2:]...)
				}
			}
		}
	}
}

func readConfigFile(f string) {
	cfg.LoadConfigBlocks(f, func(l, d string) {
		switch l {
		case "params":
			cParams = strings.TrimSpace(d)
		case "head":
			cHead = strings.TrimSpace(d)
		case "tail":
			cTail = strings.TrimSpace(d)
		default:
			dbg.Fatal("Unknown config block: %s", l)
		}
	})
	if "" != cParams {
		processArgs(cParams, handleArgs)
	}
}
