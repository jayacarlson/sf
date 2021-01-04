package main

import (
	"errors"
	"os"

	"github.com/jayacarlson/dbg"
)

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
