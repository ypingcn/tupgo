package tup

import (
	"errors"
)

// ErrNeedWriteTo - Need WriteTo function , generate struct from tars2go tools
var ErrNeedWriteTo = errors.New("NEED WriteTo")

// ErrNeedReadFrom - Need ReadFrom function , generate struct from tars2go tools
var ErrNeedReadFrom = errors.New("NEED ReadFrom")

// ErrTUPVersionNotSupported - Version not supported, only 2 and 3 currently
var ErrTUPVersionNotSupported = errors.New("TUP VERSION NOT SUPPORT")

// ErrEmptyServantNameFuncName - Empty ServantName and empty FuncName
var ErrEmptyServantNameFuncName = errors.New("EMPTY ServantName and EMPTY FuncName")
