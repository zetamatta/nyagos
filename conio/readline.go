package conio

import "fmt"
import "strings"

type KeyFuncResult int

const (
	CONTINUE KeyFuncResult = iota
	ENTER    KeyFuncResult = iota
	ABORT    KeyFuncResult = iota
)

type KeyFuncT interface {
	Call(buffer *ReadLineBuffer) KeyFuncResult
}

type KeyGoFuncT struct {
	F func(buffer *ReadLineBuffer) KeyFuncResult
}

func (this *KeyGoFuncT) Call(buffer *ReadLineBuffer) KeyFuncResult {
	return this.F(buffer)
}

var KeyMap = map[rune]KeyFuncT{
	NAME2CHAR[K_CTRL_A]: NAME2FUNC[F_BEGINNING_OF_LINE],
	NAME2CHAR[K_CTRL_B]: NAME2FUNC[F_BACKWARD_CHAR],
	NAME2CHAR[K_CTRL_C]: NAME2FUNC[F_INTR],
	NAME2CHAR[K_CTRL_D]: NAME2FUNC[F_DELETE_OR_ABORT],
	NAME2CHAR[K_CTRL_E]: NAME2FUNC[F_END_OF_LINE],
	NAME2CHAR[K_CTRL_F]: NAME2FUNC[F_FORARD_CHAR],
	NAME2CHAR[K_CTRL_H]: NAME2FUNC[F_BACKWARD_DELETE_CHAR],
	NAME2CHAR[K_CTRL_K]: NAME2FUNC[F_KILL_LINE],
	NAME2CHAR[K_CTRL_L]: NAME2FUNC[F_CLEAR_SCREEN],
	NAME2CHAR[K_CTRL_M]: NAME2FUNC[F_ACCEPT_LINE],
	NAME2CHAR[K_CTRL_U]: NAME2FUNC[F_UNIX_LINE_DISCARD],
	NAME2CHAR[K_CTRL_Y]: NAME2FUNC[F_YANK],
	NAME2CHAR[K_DELETE]: NAME2FUNC[F_DELETE_CHAR],
	NAME2CHAR[K_ENTER]:  NAME2FUNC[F_ACCEPT_LINE],
	NAME2CHAR[K_ESCAPE]: NAME2FUNC[F_KILL_WHOLE_LINE],
}

var ZeroMap = map[uint16]KeyFuncT{
	NAME2SCAN[K_CTRL]:   NAME2FUNC[F_PASS],
	NAME2SCAN[K_DELETE]: NAME2FUNC[F_DELETE_CHAR],
	NAME2SCAN[K_END]:    NAME2FUNC[F_END_OF_LINE],
	NAME2SCAN[K_HOME]:   NAME2FUNC[F_BEGINNING_OF_LINE],
	NAME2SCAN[K_LEFT]:   NAME2FUNC[F_BACKWARD_CHAR],
	NAME2SCAN[K_RIGHT]:  NAME2FUNC[F_FORARD_CHAR],
	NAME2SCAN[K_SHIFT]:  NAME2FUNC[F_PASS],
}

func normWord(src string) string {
	return strings.Replace(strings.ToUpper(src), "-", "_", -1)
}

func BindKeyFunc(keyName string, funcValue KeyFuncT) error {
	keyName_ := normWord(keyName)
	if charValue, charOk := NAME2CHAR[keyName_]; charOk {
		KeyMap[charValue] = funcValue
		return nil
	} else if scanValue, scanOk := NAME2SCAN[keyName_]; scanOk {
		ZeroMap[scanValue] = funcValue
		return nil
	} else {
		return fmt.Errorf("%s: no such keyname", keyName)
	}
}

func BindKeySymbol(keyName, funcName string) error {
	funcValue, funcOk := NAME2FUNC[normWord(funcName)]
	if !funcOk {
		return fmt.Errorf("%s: no such function.", funcName)
	}
	return BindKeyFunc(keyName, funcValue)
}

func BindKeySymbolFunc(keyName, funcName string, funcValue KeyFuncT) error {
	NAME2FUNC[normWord(funcName)] = funcValue
	return BindKeyFunc(keyName, funcValue)
}

func ReadLine(prompt_ func() int) (string, KeyFuncResult) {
	this := ReadLineBuffer{Buffer: make([]rune, 20)}
	this.ViewWidth, _ = GetScreenSize()
	this.ViewWidth--

	this.Prompt = prompt_
	if this.Prompt != nil {
		this.ViewWidth = this.ViewWidth - this.Prompt()
	}
	for {
		stdOut.Flush()
		shineCursor()
		this.Unicode, this.Keycode = GetKey()
		var f KeyFuncT
		var ok bool
		if this.Unicode != 0 {
			f, ok = KeyMap[this.Unicode]
			if !ok {
				//f = KeyFuncInsertReport
				f = &KeyGoFuncT{KeyFuncInsertSelf}
			}
		} else {
			f, ok = ZeroMap[this.Keycode]
			if !ok {
				f = &KeyGoFuncT{KeyFuncPass}
			}
		}
		rc := f.Call(&this)
		if rc != CONTINUE {
			stdOut.WriteRune('\n')
			stdOut.Flush()
			return this.String(), rc
		}
	}
}
