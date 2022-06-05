package main

import (
	"errors"
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
	bug                                  = dbg.Dbg{}
	tokRex                               = regexp.MustCompile("((?s).*?)%(.)((?s).*)")
	tMap                                 tokenMap
	bashHeader, reverse, dontHomify      bool
	recursive, hiddenFiles, hiddenDirs   bool
	ignoreECase, help, cfgHelp, metaHelp bool
	configFile, outputFile               string
	homifiedProcessDir, homifiedRealPath string
	homeDir, currentFullPath             string
	include, exclude, incList, excList   string
	leadOutput, tailOutput               string
	aLeadOutput, aTailOutput             string
	dirOutput, fileOutput                string
	cParams, cHead, cTail                string
	cArgs                                [9]string
	fileCount                            int64
	totalCount                           int64 = 0
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
  -c file     Config file
  -?c         Show help on configuration file settings
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
  -L string   Startup leading output string (limited metachars: OH)
  -T string   Final trailing output string (limited metachars: OHT)
  -l string   Per [dir list] directory lead output string (limited metachars: OHrR)
  -t string   Per [dir list] directory tail output string (limited metachars: OHrRCT)

  -1 ... -9 string   Special case when using config files

    If no [dir list] given, the PWD (./) is used.

    Only -i OR -x can be used.  To include/exclude files with no extension use '-'.
     e.g. -[ix] "go - txt"
`
	metaHelpString = `The supported %meta values are as follows:
  O   Origin path (PWD) from where sf was called, always homified
  H   Users real HOME path: /home/<user>
  R   Root to [dir list] path (homified by default)
  r   Current [dir list] directory
  P   Current full dirpath (homified by default)
  p   Dirpath from [dir list] on down
  D   Dirpath below [dir list] directory
  d   Latest directory name
  s   The file/dir size
  c   File count inside the dir (dir output line)
  C   Dir count inside the dir (dir output line)
  T   Current total file count (over all [dir list] dirs)
 --- files only output
  f   Current [dir list] filepath
  F   Current full filepath (homified by default)
  n   Current filename and extension as read: 'file.ext'
  N   Current filename without any extension: 'file'
  e   Current extension, no leading '.':      'ext'
  E   Current file extention, with '.':       '.ext'
  c   Current file count inside the dir
  C   Current file count inside [dir list] dir

  NOTE: the meta values r, p, d, D, f, n, N, e & E can be prepended
   with a 'u' to uppercasify the value, or 'l' lowercasify the value.
   e.g.:  'Dir/File.Ext' can be adjusted to
          'DIR' | 'dir' / 'FILE' | 'file' / 'EXT' | 'ext'
           %ur     %lr     %un      %ln      %ue     %le
  NOTE: the meta values p, D & f can be prepended with '@' to replace
   the dir separator '/' with '@'.  (Cannot combine with 'u' & 'l')
   e.g.: %@f of 'dir/sub-dir/file' becomes 'dir@sub-dir@file'
`
	cfgHelpString = `Read the given 'configuration' file looking for 'params' 'head' and 'tail'
blocks.  The 'params' override any given command line arguments.  The 'head'
block is output after any bash header request (-b) and before any -H line,
followed by file/director processing, then any -T line and finally the
'tail' block is output.

In addition to the standard %meta characters (which can be over-ridden by
the 'params' field) the arguments -1 through -9 can be given on the command
line and can then be used as the meta characters %1 through %9 while SF is 
processing the 'head' and 'tail' blocks.

params <
-r -h -i "jpg jpeg png gif tiff" -I
>

head <
These are all the image files...
>

tail <
... All done
>
`
)

func init() {
	flag.BoolVar(&metaHelp, "?", false, "error")
	flag.BoolVar(&cfgHelp, "?c", false, "error")
	flag.BoolVar(&help, "help", false, "error")
	flag.BoolVar(&hiddenDirs, "D", false, "bool")
	flag.BoolVar(&hiddenFiles, "F", false, "bool")
	flag.BoolVar(&dontHomify, "h", false, "bool")
	flag.BoolVar(&ignoreECase, "I", false, "bool")
	flag.BoolVar(&recursive, "r", false, "bool")
	flag.BoolVar(&bashHeader, "b", false, "bool")
	flag.BoolVar(&reverse, "s", false, "bool")

	flag.StringVar(&outputFile, "o", "", "string")
	flag.StringVar(&include, "i", "", "string")
	flag.StringVar(&exclude, "x", "", "string")
	flag.StringVar(&fileOutput, "f", "", "string")
	flag.StringVar(&leadOutput, "L", "", "string")
	flag.StringVar(&tailOutput, "T", "", "string")
	flag.StringVar(&aLeadOutput, "l", "", "string")
	flag.StringVar(&aTailOutput, "t", "", "string")
	flag.StringVar(&dirOutput, "d", "", "string")

	flag.StringVar(&configFile, "c", "", "error")
	flag.StringVar(&cArgs[0], "1", "", "error")
	flag.StringVar(&cArgs[1], "2", "", "error")
	flag.StringVar(&cArgs[2], "3", "", "error")
	flag.StringVar(&cArgs[3], "4", "", "error")
	flag.StringVar(&cArgs[4], "5", "", "error")
	flag.StringVar(&cArgs[5], "6", "", "error")
	flag.StringVar(&cArgs[6], "7", "", "error")
	flag.StringVar(&cArgs[7], "8", "", "error")
	flag.StringVar(&cArgs[8], "9", "", "error")

	homeDir = pth.AsRealPath("~")
	tMap = make(tokenMap)
	tMap["%"] = "%"
	tMap.safeset("H", homeDir)
	tMap.safeset("O", homifyDir(pth.AsRealPath(".")))
}

func (t tokenMap) safeset(tok, str string) {
	str = strings.ReplaceAll(str, ` `, `\ `)
	str = strings.ReplaceAll(str, `(`, `\(`) // are these others needed?
	str = strings.ReplaceAll(str, `)`, `\)`)
	str = strings.ReplaceAll(str, `'`, `\'`)
	str = strings.ReplaceAll(str, `"`, `\"`)
	t[tok] = str
}

func (t tokenMap) replace(src string) string {
	if x := tokRex.FindStringSubmatch(src); x != nil {
		vl, ok := "", false
		if (x[2] == "0" || x[2] == " ") && (x[3][0] > '1' && x[3][0] <= '9') { // 2..9 digits
			key := x[3][1:2]
			vl, ok = t[key]
			dbg.ChkTruX(ok, "Unknown replacement token: %%%s", key)
			dbg.ChkTruX(key == "c" || key == "s" || key == "C" || key == "T",
				"Illegal replacement token: %%%s", key)
			flen := int(x[3][0]-'0') - len(vl)
			x[3] = x[3][2:]
			if flen > 0 {
				vl = strings.Repeat(x[2], flen) + vl
			}
		} else if x[2] == "@" && (x[3][0] == 'p' || x[3][0] == 'D' || x[3][0] == 'f') {
			x[2] = x[3][0:1]
			x[3] = x[3][1:]
			vl, _ = t[x[2]]
			vl = strings.ReplaceAll(vl, "/", "@")
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
				dbg.ChkTruX(-1 != strings.Index("rpdDfnNeE", x[2]),
					"Cannot change case for: %s", x[2])
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
	outLine := strings.ReplaceAll(t.replace(src), "\\n", "\n")
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

func addNumberArgs() {
	if "" != cArgs[0] {
		tMap["1"] = cArgs[0]
	}
	if "" != cArgs[1] {
		tMap["2"] = cArgs[1]
	}
	if "" != cArgs[2] {
		tMap["3"] = cArgs[2]
	}
	if "" != cArgs[3] {
		tMap["4"] = cArgs[3]
	}
	if "" != cArgs[4] {
		tMap["5"] = cArgs[4]
	}
	if "" != cArgs[5] {
		tMap["6"] = cArgs[5]
	}
	if "" != cArgs[6] {
		tMap["7"] = cArgs[6]
	}
	if "" != cArgs[7] {
		tMap["8"] = cArgs[7]
	}
	if "" != cArgs[8] {
		tMap["9"] = cArgs[8]
	}
}
func clearNumberArgs() {
	delete(tMap, "0") // remove any number metachars
	delete(tMap, "1")
	delete(tMap, "2")
	delete(tMap, "3")
	delete(tMap, "4")
	delete(tMap, "5")
	delete(tMap, "6")
	delete(tMap, "7")
	delete(tMap, "8")
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
	var nm, ext string
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

		tMap.safeset("n", fileName)
		_, nm, ext = pth.Split(realPath)

		tMap.safeset("N", nm)
		tMap.safeset("E", ext)
		if ext != "" {
			ext = ext[1:]
		}
		tMap.safeset("e", ext)

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
		tMap.safeset("F", path.Clean(homifiedRealPath+"/"+fileName))
		if tMap["p"] == "." {
			tMap.safeset("f", fileName)
		} else {
			tMap.safeset("f", currentFullPath+"/"+fileName)
		}
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
			if "" != incList || "" != excList {
				_, _, ext := pth.Split(entry.Name())
				if ext != "" {
					ext = ext[1:]
				} else {
					ext = "-"
				}
				if ignoreECase {
					ext = strings.ToLower(ext)
				}

				if "" != incList && -1 == strings.Index(incList, " "+ext+" ") {
					continue
				}
				if "" != excList && -1 != strings.Index(excList, " "+ext+" ") {
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
	homifiedRealPath = homifyDir(realPath)
	currentFullPath = path.Join(homifiedProcessDir, curPath)
	tMap.safeset("P", homifiedRealPath)
	tMap.safeset("p", currentFullPath)
	tMap.safeset("D", curPath)
	tMap.safeset("d", curDir)
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
				homifiedRealPath = homifyDir(realPath)
				tMap["s"] = strconv.FormatInt(fi.Size(), 10)
				tMap.safeset("P", homifiedRealPath)
				tMap.safeset("d", dirName)
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
	homifiedProcessDir = homifyDir(curDir)
	tMap.safeset("R", homifyDir(dirRoot))
	tMap.safeset("r", homifiedProcessDir)
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

var (
	Err_NotExist   = errors.New("File/dir doesn't exist")
	Err_Permission = errors.New("Permission Denied")
)

func chkErr(err error) error {
	if nil != err {
		switch t := err.(type) {
		case *os.PathError:
			if os.IsNotExist(err) {
				return Err_NotExist
			}
			if os.IsPermission(err) {
				return Err_Permission
			}
			dbg.Message("OS path err: %v", err)
		default:
			dbg.Message("Path err type: %v", t)
		}
	}
	return err
}

func chkDirErr(realPath string, err error) error {
	err = chkErr(err)
	switch err {
	case nil:
		return nil
	case Err_NotExist:
		dbg.Warning("Failed to stat dir: `%s`", realPath)
		return err
	case Err_Permission:
		dbg.Warning("Failed to open restricted dir: `%s`", realPath)
		return err
	default:
		dbg.Error("Error %v for dir `%s`", err, realPath)
		return err
	}
}

func chkFileErr(realPath string, err error) error {
	err = chkErr(err)
	switch err {
	case nil:
		return nil
	case Err_NotExist:
		dbg.Warning("Failed to stat file: `%s`", realPath)
		return nil
	case Err_Permission:
		dbg.Warning("Failed to stat restricted file: `%s`", realPath)
		return nil
	default:
		dbg.Error("Error %v for file `%s`", err, realPath)
		return err
	}
}

func main() {
	var outTo *os.File = os.Stdout // default output to stdout
	//	bug.Enabled = true

	flag.Parse()
	if help {
		fmt.Printf("%s", helpString)
		return
	}
	if metaHelp {
		fmt.Printf("%s", metaHelpString)
		return
	}
	if cfgHelp {
		fmt.Printf("%s", cfgHelpString)
		return
	}
	if "" != configFile {
		readConfigFile(configFile)
	}
	if "" != include && "" != exclude {
		dbg.Fatal("Can only use -i or -x, not both")
	}
	if "" != include {
		if ignoreECase {
			include = strings.ToLower(include)
		}
		incList = " " + include + " "
	}
	if "" != exclude {
		if ignoreECase {
			exclude = strings.ToLower(exclude)
		}
		excList = " " + exclude + " "
	}
	if "" == dirOutput && "" == fileOutput {
		fileOutput = "%f"
	}

	if "" != outputFile {
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
		if "" != cParams {
			args += "( " + cParams + " )"
		}
		tMap["a"] = args
		fmt.Fprintln(outTo, tMap.replace(bashHead))
		delete(tMap, "a")
	}
	if "" != cHead {
		addNumberArgs()
		fmt.Fprintln(outTo, tMap.replace(cHead))
		clearNumberArgs()
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
	if "" != cTail {
		addNumberArgs()
		fmt.Fprintln(outTo, tMap.replace(cTail))
		clearNumberArgs()
	}
}
