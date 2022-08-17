package main

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jayacarlson/dbg"
	"github.com/jayacarlson/tst"
)

type bufRW struct {
	buffer *bytes.Buffer
}

func (s bufRW) Write(p []byte) (int, error) {
	return s.buffer.Write(p)
}

func (s bufRW) Read(p []byte) (int, error) {
	return s.buffer.Read(p)
}

func (s bufRW) Size() int64 {
	return int64(s.buffer.Len())
}

func (s bufRW) Reset() {
	s.buffer.Reset()
}

func (s bufRW) MD5Sum() string {
	data := make([]byte, s.Size())
	s.buffer.Read(data)
	sum := fmt.Sprintf("%x", md5.Sum(data))
	if showTestData {
		dbg.Message("%s", string(data))
	}
	if showTestSums {
		dbg.Info("%s", sum)
	}
	return sum
}

var (
	showTestData bool
	showTestSums bool
	outTo        bufRW
)

func init() {
	flag.BoolVar(&showTestData, "stt", false, "Show generated test text")
	flag.BoolVar(&showTestSums, "sts", false, "Show generated test MD5 sum")
	outTo.buffer = bytes.NewBuffer(make([]byte, 0))
}

func initFunc() {
	fileOutput, dirOutput = "%f", ""
	aLeadOutput, aTailOutput = "", ""
	// leadOutput, tailOutput not testable as that happens in 'main'
	ignoreECase = false
	hiddenFiles = false
	hiddenDirs = false
	recursive = false
	reverse = false
	incList = ""
	excList = ""
	totalCount = 0
}

func finiFunc() {
	outTo.Reset()
}

func setIELists(ignore bool, include, exclude string) {
	ignoreECase = ignore
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
}

// ========================================================================= //
//	Can't use %P, %R or %F as they are unique to each user's dir structure

func testDirsNonRecursive() (string, string, bool) {
	fileOutput = ""
	dirOutput = "r: %r   p: %p   D: %D   d: %d"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "dd32871c2c927dad0a655ac8f4155b1a"
}

func testDirsRecursive() (string, string, bool) {
	recursive = true
	fileOutput = ""
	dirOutput = "r: %r   p: %p   D: %D   d: %d"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "546f6d03637df3138ceb857cf0ee44b1"
}

func testDirsReverseRecursive() (string, string, bool) {
	reverse = true
	recursive = true
	fileOutput = ""
	dirOutput = "r: %r   p: %p   D: %D   d: %d"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "de20c3392d493f80247b422e5bdbe34a"
}

func testDirsAlterCase() (string, string, bool) {
	recursive = true
	fileOutput = ""
	dirOutput = " r: %r   p: %p   D: %D   d: %d\nlr: %lr  lp: %lp  lD: %lD  ld: %ld\nur: %ur  up: %up  uD: %uD  ud: %ud"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "648c63c6bd4cc350519d734e6978c618"
}

func testDirsBashOutput() (string, string, bool) {
	recursive = true
	fileOutput = ""
	// bashHeader processed by 'main', but we can get argDir lead/tail
	aLeadOutput = "# Going to run 'someTool' on dirs in '%r'"
	aTailOutput = "# Done with '%r'"
	dirOutput = "someTool %p someOtherDir/%D"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "6ae1c460dea9c5de52d7ec8b3924a0bd"
}

func testFilesNonRecursive() (string, string, bool) {
	fileOutput = "f: %f  n: %n  N: %N  e: %e  E: %E  c: %c  C: %C"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "7d639395050a94b5aa53229b0b16289b"
}

func testFilesRecursive() (string, string, bool) {
	recursive = true
	fileOutput = "f: %f  n: %n  N: %N  e: %e  E: %E  c: %c  C: %C"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "58ec0905a533b1f2fb525c2147d92313"
}

func testFilesReverseRecursive() (string, string, bool) {
	reverse = true
	recursive = true
	fileOutput = "f: %f  n: %n  N: %N  e: %e  E: %E  c: %c  C: %C"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "076668796119c36c488f1d2d44f2fceb"
}

func testFilesNoExt() (string, string, bool) {
	recursive = true
	setIELists(false, "-", "")
	fileOutput = "f: %f  n: %n  N: %N  e: %e  E: %E  c: %c  C: %C"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "b86e1dd278bfde0201c0f6ed256a2ff6"
}

func testFilesIncExt() (string, string, bool) {
	recursive = true
	setIELists(false, "ext ex1 ex2", "")
	fileOutput = "f: %f  n: %n  N: %N  e: %e  E: %E  c: %c  C: %C"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "2f5f9c5c394fa05a4263c70e55337ad6"
}

func testFilesExcExt() (string, string, bool) {
	recursive = true
	setIELists(false, "", "ext ex1 ex2")
	fileOutput = "f: %f  n: %n  N: %N  e: %e  E: %E  c: %c  C: %C"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "afe7e561ba3a61beb191f940a5eb5848"
}

func testFilesIncExtIgCase() (string, string, bool) {
	recursive = true
	setIELists(true, "ext -", "")
	fileOutput = "f: %f  n: %n  N: %N  e: %e  E: %E  c: %c  C: %C"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "b681cc0fa83baf30f7565d80e459df3f"
}

func testFilesExcExtIgCase() (string, string, bool) {
	recursive = true
	setIELists(true, "", "ext -")
	fileOutput = "f: %f  n: %n  N: %N  e: %e  E: %E  c: %c  C: %C"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "4b89da8ee6ac1cded5119fb4231411c6"
}

func testFilesAlterCase() (string, string, bool) {
	recursive = true
	fileOutput = " f: %f   n: %n   N: %N   e: %e   E: %E  c: %c  C: %C\nuf: %uf  un: %un  uN: %uN  ue: %ue  uE: %uE\nlf: %lf  ln: %ln  lN: %lN  le: %le  lE: %lE"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "6da7b87df46dd0b4afe8dce9c93315d2"
}

func testFilesBashOutput() (string, string, bool) {
	recursive = true
	// bashHeader processed by 'main', but we can get argDir lead/tail
	aLeadOutput = "# Going to run 'someTool' on files in '%r'"
	aTailOutput = "# Done with '%r'"
	fileOutput = "someTool %f someOtherDir/%D/x-%n-x"
	processDir(outTo, "testdata")
	sum := outTo.MD5Sum()
	return dbg.IAm(), "", sum != "539d5a154f00b639f5234c882d9e5940"
}

func TestDirs(t *testing.T) {
	if tst.Testing(dbg.IAm(), "", true) {
		tst.Func(t, testDirsNonRecursive)
		tst.Func(t, testDirsRecursive)
		tst.Func(t, testDirsReverseRecursive)
		tst.Func(t, testDirsAlterCase)
		tst.Func(t, testDirsBashOutput)
	}
}

func TestFiles(t *testing.T) {
	if tst.Testing(dbg.IAm(), "", true) {
		tst.Func(t, testFilesNonRecursive)
		tst.Func(t, testFilesRecursive)
		tst.Func(t, testFilesReverseRecursive)
		tst.Func(t, testFilesNoExt)
		tst.Func(t, testFilesIncExt)
		tst.Func(t, testFilesExcExt)
		tst.Func(t, testFilesIncExtIgCase)
		tst.Func(t, testFilesExcExtIgCase)
		tst.Func(t, testFilesAlterCase)
		tst.Func(t, testFilesBashOutput)
	}
}

func TestMain(m *testing.M) {
	flag.Parse()

	tst.SetInitFunc(initFunc)
	tst.SetFiniFunc(finiFunc)
	os.Exit(m.Run())
}
