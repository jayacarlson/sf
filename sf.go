package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/jayacarlson/dbg"
	"github.com/jayacarlson/pth"
)

type tokenMap map[string]string

var (
	bug         = dbg.Dbg{}
	tokRex      = regexp.MustCompile("((?s).*?)%(.)((?s).*)")
	recursive   bool
	bashHeader  bool
	hiddenFiles bool
	hiddenDirs  bool
	ignoreECase bool
	reverse     bool
	dontHomify  bool
	help        bool
	metahelp    bool
	tMap        tokenMap
	incList     string
	excList     string
	outputFile  string
	homeDir     string
	leadOutput  string
	tailOutput  string
	aLeadOutput string
	aTailOutput string
	dirOutput   string
	fileOutput  string
	include     string
	exclude     string
	fileCount   int64
	totalCount  int64 = 0
)

const (
	bashHead = `#!/bin/bash
#
#	sf %a
#
`
	helpString = `Usage of sf:  [args] [dir list]
  -?          Show help with output string meta-characters
  -b          Output BASH header at start (output file made executable)
  -D          Include hidden directories
  -F          Include hidden files
  -h          Do not ~/homify paths (simplify full paths to ~/ if possible)
  -I          Ignore case when filtering by file extension
  -r          Recurse into directories
  -s          Sort in decending order
  -o string   File to output data
  -i string   File filter by list of extensions (inclusive)
  -x string   File filter by list of extensions (exclusive)
  -d string   Per directory output (default: "" - limited metachars: OHrRdDpP)
  -f string   Output string per file (defaults to '%f' -- filepath)
  -L string   Startup lead output string (limited metachars: OH)
  -T string   Final tail output string (limited metachars: OHT)
  -l string   [dir list] directory lead output string (limited metachars: OHrR)
  -t string   [dir list] directory tail output string (limited metachars: OHrRCT)

    If no [dir list] given, the PWD (./) is used.

    To include/exclude files with no extension use '-'.
     e.g. -[ix] "go - txt"
`
	metaHelpString = `The supported %meta values are as follows:
  O   Origin path (PWD) from where sf was called, always homified
  H   Users real HOME path: /home/<user>
  r   [dir list] directory as given from [dir list]
  R   Full dirpath of given [dir list] directory (homified by default)
  p   Current dirpath, from [dir list] directory on down
  P   Current full dirpath, from [dir list] dir on down (homified by default)
  d   Latest directory name
  D   Current dirpath, below [dir list] directory
  s   The file/dir size
  c   File count inside the dir (dir output line)
  C   Dir count inside the dir (dir output line)
  T   Current total file count (over all [dir list] dirs)
 --- files only output
  f   Current filepath from [dir list] directory on down
  F   Current full filepath (homified by default)
  n   Current filename and extension as read: file.ext
  N   Current filename without any extension: file
  e   Current extension, without leading '.': ext
  E   Current extension, including the '.':  .ext
  c   Current file count inside the dir
  C   Current file count inside [dir list] dir

  NOTE: the meta values r, p, d, D, f, n, N, e & E can be prepended
   with a 'u' to uppercasify the value, or 'l' lowercasify the value.
   e.g.:  'Dir/File.Ext' can be adjusted to
          'DIR' | 'dir' / 'FILE' | 'file' / 'EXT' | 'ext'
`
)

func init() {
	flag.BoolVar(&metahelp, "?", false, "Show help with output string meta-characters")
	flag.BoolVar(&hiddenDirs, "D", false, "Include hidden directories")
	flag.BoolVar(&hiddenFiles, "F", false, "Include hidden files")
	flag.BoolVar(&dontHomify, "h", false, "Do not ~/homify paths (simplify full paths to ~/ if possible)")
	flag.BoolVar(&ignoreECase, "I", false, "Ignore case when filtering by file extension")
	flag.BoolVar(&help, "help", false, "Show arg help")
	flag.BoolVar(&recursive, "r", false, "Recurse into directories")
	flag.BoolVar(&bashHeader, "b", false, "Output BASH header at start")
	flag.BoolVar(&reverse, "s", false, "Sort in decending order")

	flag.StringVar(&outputFile, "o", "", "File to output data")
	flag.StringVar(&include, "i", "", "Filter by list of extensions (inclusive)")
	flag.StringVar(&exclude, "x", "", "Filter by list of extensions (exclusive)")
	flag.StringVar(&fileOutput, "f", "", "Output string per file (defaults to '%f' -- filepath)")
	flag.StringVar(&leadOutput, "L", "", "Startup lead output string")
	flag.StringVar(&tailOutput, "T", "", "Final tail output string")
	flag.StringVar(&aLeadOutput, "l", "", "[dir list] directory lead output string")
	flag.StringVar(&aTailOutput, "t", "", "[dir list] directory tail output string")
	flag.StringVar(&dirOutput, "d", "", "Per directory lead output string")

	homeDir = pth.AsRealPath("~")
	tMap = make(tokenMap)
	tMap["%"] = "%"
	tMap["H"] = homeDir
	tMap["O"] = homifyDir(pth.AsRealPath("."))
}

func (t tokenMap) replace(src string) string {
	if x := tokRex.FindStringSubmatch(src); x != nil {
		vl, ok := "", false
		if (x[2] == "0" || x[2] == " ") && (x[3][0] > '1' && x[3][0] < '9') { // 2..8 digits
			key := x[3][1:2]
			vl, ok = t[key]
			dbg.ChkTruX(ok, "Unknown replacement token: %%%s", key)
			dbg.ChkTruX(key=="c" || key=="C" || key=="T", "Illegal replacement token: %%%s", key)
			flen := int(x[3][0] - '0') - len(vl)
			x[3] = x[3][2:]
			if flen > 0 {
				vl = strings.Repeat(x[2],flen) + vl
			}
		} else {
			lc, uc := false, false
			if x[2] == "u" {
				uc = true
			} else if x[2] == "l" {
				lc = true
			}
			if uc || lc {
				x[2] = x[3][0:1]
				x[3] = x[3][1:]
				dbg.ChkTruX(-1 != strings.Index("rpdDfnNeE", x[2]), "Cannot change case for: %s", x[2])
			}
			vl, ok = t[x[2]]
			dbg.ChkTruX(ok, "Unknown replacement token: %%%s", x[2])
			if uc {
				vl = strings.ToUpper(vl)
			} else if lc {
				vl = strings.ToLower(vl)
			}
		}
		return x[1] + vl + t.replace(x[3])
	}
	return src
}

func (t tokenMap) output(outTo io.Writer, src string) {
	outLine := strings.Replace(t.replace(src), "\\n", "\n", -1)
	fmt.Fprintln(outTo, outLine)
}

func (t tokenMap) String() string {
	out := "[\n"
	for t, v := range t {
		out += fmt.Sprintf("  %s: %s\n", t, v)
	}
	return out + "]"
}

func homifyDir(theDir string) string {
	if !dontHomify {
		if len(theDir) >= len(homeDir) {
			if theDir[:len(homeDir)] == homeDir {
				theDir = "~" + theDir[len(homeDir):]
			}
		}
	}
	return theDir
}

func clearFileMetas() {
	delete(tMap, "c") // remove any previous 'file' metachars
	delete(tMap, "C")
	delete(tMap, "T")
	delete(tMap, "f")
	delete(tMap, "F")
	delete(tMap, "n")
	delete(tMap, "N")
	delete(tMap, "e")
	delete(tMap, "E")
}

func clearDirMetas() {
	delete(tMap, "c") // remove any previous 'dir' metachars
	delete(tMap, "C")
	delete(tMap, "d")
	delete(tMap, "D")
	delete(tMap, "p")
	delete(tMap, "P")
	delete(tMap, "s")
}

func handleFiles(outTo io.Writer, dirPath string, fileNames []string) error {
	var ext string
	var count int64 = 0
	for _, fileName := range fileNames {
		realPath := pth.AsRealPath(dirPath, fileName)

		fi, err := os.Stat(realPath)
		err = chkErr(err)
		if nil != err {
			if err == Err_NotExist {
				continue
			}
			dbg.Error("Error %v for file `%s`", err, realPath)
			return err
		}

		tMap["n"] = fileName
		_, tMap["N"], ext = pth.Split(realPath)
		tMap["E"] = ext
		if ext != "" {
			ext = ext[1:]
		}
		tMap["e"] = ext

		if ignoreECase {
			ext = strings.ToLower(ext)
		}

		count += 1
		fileCount += 1
		totalCount += 1
		tMap["c"] = strconv.FormatInt(count, 10)
		tMap["C"] = strconv.FormatInt(fileCount, 10)
		tMap["T"] = strconv.FormatInt(totalCount, 10)
		tMap["s"] = strconv.FormatInt(fi.Size(), 10)
		tMap["F"] = path.Clean(tMap["P"] + "/" + fileName)
		tMap["f"] = tMap["p"] + "/" + fileName
		tMap.output(outTo, fileOutput)
	}
	return nil
}

func handleDir(outTo io.Writer, dirRoot, dirPath, curDir string) error {
	theDirs := []string{}
	theFiles := []string{}
	realPath := pth.AsRealPath(dirRoot, dirPath, curDir)
	curPath := path.Join(dirPath, curDir)

	bug.Warning("handleDir: dirRoot: %s  dirPath: %s  curDir: %s", dirRoot, dirPath, curDir)
	//bug.Info("realPath: %s", realPath)
	//bug.Info("curPath:  %s", curPath)

	// validate latest realPath, (test dirs in recursion situation)
	//  should only possibly get Err_Permission
	fi, err := os.Stat(realPath)
	err = chkDirErr(realPath, err)
	if nil != err {
		return err
	}

	entries, err := ioutil.ReadDir(realPath)
	err = chkErr(err)
	if nil != err {
		if err == Err_Permission {
			dbg.Warning("Failed to open restricted dir: `%s`", realPath)
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if !hiddenDirs && len(entry.Name()) > 1 && entry.Name()[0] == '.' {
				continue
			}
			theDirs = append(theDirs, entry.Name())
		} else {
			if !entry.Mode().IsRegular() {
				continue
			}
			if !hiddenFiles && entry.Name()[0] == '.' {
				continue
			}
			if incList != "" || excList != "" {
				_, _, ext := pth.Split(entry.Name())
				if ext != "" {
					ext = ext[1:]
				} else {
					ext = "-"
				}
				if ignoreECase {
					ext = strings.ToLower(ext)
				}

				if incList != "" && -1 == strings.Index(incList, " "+ext+" ") {
					continue
				}
				if excList != "" && -1 != strings.Index(excList, " "+ext+" ") {
					continue
				}
			}
			theFiles = append(theFiles, entry.Name())
		}
	}

	if reverse {
		for b, e := 0, len(theDirs)-1; b < e; b, e = b+1, e-1 {
			theDirs[b], theDirs[e] = theDirs[e], theDirs[b]
		}
	}

	clearFileMetas()
	tMap["P"] = homifyDir(realPath)
	tMap["p"] = path.Join(tMap["r"], curPath)
	tMap["D"] = curPath
	tMap["d"] = curDir
	tMap["s"] = strconv.FormatInt(fi.Size(), 10)
	tMap["c"] = strconv.FormatInt(int64(len(theFiles)), 10)
	tMap["C"] = strconv.FormatInt(int64(len(theDirs)), 10)
	tMap["T"] = strconv.FormatInt(totalCount, 10)

	// output dir lead (argDir / recursive)
	if dirOutput != "" {
		tMap.output(outTo, dirOutput)
	}

	if fileOutput != "" {
		if reverse {
			for b, e := 0, len(theFiles)-1; b < e; b, e = b+1, e-1 {
				theFiles[b], theFiles[e] = theFiles[e], theFiles[b]
			}
		}
		err = handleFiles(outTo, realPath, theFiles)
		if nil != err {
			return err
		}
	} else {
		totalCount += int64(len(theFiles))
	}

	if 0 < len(theDirs) {
		for _, dirName := range theDirs {
			if recursive {
				err = handleDir(outTo, dirRoot, curPath, dirName)
				if nil != err {
					return err
				}
			} else if dirOutput != "" {
				realPath := pth.AsRealPath(dirRoot, curPath, dirName)
				fi, err := os.Stat(realPath)
				err = chkDirErr(realPath, err)
				if nil != err {
					return err
				}
				tMap["s"] = strconv.FormatInt(fi.Size(), 10)
				tMap["P"] = homifyDir(realPath)
				tMap["d"] = dirName
				tMap.output(outTo, dirOutput)
			}
		}
	}

	return nil
}

func processDir(outTo io.Writer, curDir string) error {
	curDir = path.Clean(curDir)
	dirRoot := pth.AsRealPath(curDir)
	if dirRoot[0] != '/' {
		dirRoot = pth.AsRealPath("./" + curDir)
	}
	tMap["R"] = homifyDir(dirRoot)
	tMap["r"] = homifyDir(curDir)
	fileCount = 0
	if aLeadOutput != "" {
		clearFileMetas()
		clearDirMetas()
		tMap.output(outTo, aLeadOutput)
	}
	err := handleDir(outTo, dirRoot, ".", ".")
	if aTailOutput != "" {
		clearFileMetas()
		clearDirMetas()
		tMap["T"] = strconv.FormatInt(totalCount, 10)
		tMap.output(outTo, aTailOutput)
	}
	return err
}

func main() {
	var outTo *os.File = os.Stdout // default output to stdout
//	bug.Enabled = true

	flag.Parse()
	if help {
		fmt.Printf("%s", helpString)
		return
	}
	if metahelp {
		fmt.Printf("%s", metaHelpString)
		return
	}
	if include != "" && exclude != "" {
		dbg.Fatal("Can only use -i or -x, not both")
	}
	if include != "" {
		if ignoreECase {
			include = strings.ToLower(include)
		}
		incList = " " + include + " "
	}
	if exclude != "" {
		if ignoreECase {
			exclude = strings.ToLower(exclude)
		}
		excList = " " + exclude + " "
	}
	if dirOutput == "" && fileOutput == "" {
		fileOutput = "%f"
	}

	if outputFile != "" {
		outputFile = pth.AsRealPath(outputFile)
		os.Remove(outputFile)
		mode := 0644
		if bashHeader {
			mode += 0100
		}
		file, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, os.FileMode(mode))
		if err != nil {
			dbg.Fatal("Failed to open output file %s", outputFile)
		}
		outTo = file
		defer outTo.Close()
	}

	if bashHeader {
		args := " "
		for _, v := range os.Args[1:] {
			if strings.Index(v, " ") >= 0 {
				args += `"` + v + `" `
			} else {
				args += v + " "
			}
		}
		tMap["a"] = args
		fmt.Fprintln(outTo, tMap.replace(bashHead))
		delete(tMap, "a")
	}

	dirs := flag.Args()
	if len(dirs) == 0 {
		dirs = append(dirs, "./")
	}

	if leadOutput != "" {
		tMap.output(outTo, leadOutput)
	}
	for _, curDir := range dirs {
		processDir(outTo, curDir)
	}
	if tailOutput != "" {
		clearFileMetas()
		clearDirMetas()
		tMap["T"] = strconv.FormatInt(totalCount, 10)
		tMap.output(outTo, tailOutput)
	}
}
