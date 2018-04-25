package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/yuin/gopher-lua"

	"github.com/zetamatta/go-box"

	"github.com/zetamatta/nyagos/readline"
	"github.com/zetamatta/nyagos/texts"
)

type KeyLuaFuncT struct {
	Chank *lua.LFunction
}

func getBufferForCallBack(L Lua) (*readline.Buffer, int) {
	table, ok := L.Get(1).(*lua.LTable)
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("bindKeyExec: call with : not ."))
		return nil, 2
	}
	userdata, ok := L.GetField(table, "buffer").(*lua.LUserData)
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("bindKey.Call: invalid object"))
		return nil, 2
	}
	buffer, ok := userdata.Value.(*readline.Buffer)
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("can not find readline.Buffer"))
		return nil, 2
	}
	return buffer, 0
}

func callReplace(L Lua) int {
	buffer, stackRc := getBufferForCallBack(L)
	if buffer == nil {
		return stackRc
	}
	pos, ok := L.Get(-2).(lua.LNumber)
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("not a number"))
		return 2
	}
	str := L.ToString(-1)
	pos_zero_base := int(pos) - 1
	if pos_zero_base > buffer.Length {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf(":replace: pos=%d: Too big.", pos)))
		return 2
	}
	buffer.ReplaceAndRepaint(pos_zero_base, string(str))
	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func callInsert(L Lua) int {
	buffer, stackRc := getBufferForCallBack(L)
	if buffer == nil {
		return stackRc
	}
	text := L.ToString(2)
	buffer.InsertAndRepaint(string(text))
	L.Push(lua.LTrue)
	return 1
}

func callKeyFunc(L Lua) int {
	buffer, stackRc := getBufferForCallBack(L)
	if buffer == nil {
		return stackRc
	}
	key := L.ToString(2)
	function, err := readline.GetFunc(key)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	ctx := context.Background()
	switch function.Call(ctx, buffer) {
	case readline.ENTER:
		L.Push(lua.LTrue)
		L.Push(lua.LTrue)
		return 2
	case readline.ABORT:
		L.Push(lua.LTrue)
		L.Push(lua.LFalse)
		return 2
	default:
		L.Push(lua.LNil)
		return 1
	}
}

func callLastWord(L Lua) int {
	this, stackCount := getBufferForCallBack(L)
	if this == nil {
		return stackCount
	}
	word, pos := this.CurrentWord()
	L.Push(lua.LString(word))
	L.Push(lua.LNumber(pos + 1))
	return 2
}

func callFirstWord(L Lua) int {
	this, stackCount := getBufferForCallBack(L)
	if this == nil {
		return stackCount
	}
	word := texts.FirstWord(this.String())
	L.Push(lua.LString(word))
	L.Push(lua.LNumber(0))
	return 2
}

func callBoxListing(L Lua) int {
	// stack +1: readline.Buffer
	// stack +2: table
	// stack +3: index or value
	this, stackCount := getBufferForCallBack(L)
	if this == nil {
		return stackCount
	}
	fmt.Print("\n")
	list := make([]string, 0, 100)
	table := L.ToTable(2)
	for i := 1; true; i++ {
		str, ok := L.GetTable(table, lua.LNumber(i)).(lua.LString)
		if !ok {
			break
		}
		list = append(list, string(str))
	}
	box.Print(nil, list, os.Stdout)
	this.RepaintAll()
	return 0
}

func (this KeyLuaFuncT) String() string {
	return this.Chank.String()
}
func (this *KeyLuaFuncT) Call(ctx context.Context, buffer *readline.Buffer) readline.Result {
	L, ok := ctx.Value(luaKey).(Lua)
	if !ok {
		println("(*mains.KeyLuaFuncT)Call: lua instance not found")
		return readline.CONTINUE
	}
	L.Push(this.Chank)
	pos := -1
	var text strings.Builder
	for i, c := range buffer.Buffer {
		if i >= buffer.Length {
			break
		}
		if i == buffer.Cursor {
			pos = text.Len() + 1
		}
		text.WriteRune(c)
	}
	if pos < 0 {
		pos = text.Len() + 1
	}

	table := L.NewTable()
	L.SetField(table, "pos", lua.LNumber(pos))
	L.SetField(table, "text", lua.LString(text.String()))
	userdata := L.NewUserData()
	userdata.Value = buffer
	L.SetField(table, "buffer", userdata)
	L.SetField(table, "call", L.NewFunction(callKeyFunc))
	L.SetField(table, "insert", L.NewFunction(callInsert))
	L.SetField(table, "replacefrom", L.NewFunction(callReplace))
	L.SetField(table, "lastword", L.NewFunction(callLastWord))
	L.SetField(table, "firstword", L.NewFunction(callFirstWord))
	L.SetField(table, "boxprint", L.NewFunction(callBoxListing))

	defer setContext(L, getContext(L))
	setContext(L, ctx)

	L.Push(table)
	err := L.PCall(1, 1, nil)
	if err != nil {
		println(err.Error())
	} else {
		switch value := L.Get(-1).(type) {
		case lua.LString:
			buffer.InsertAndRepaint(string(value))
		case lua.LBool:
			if !value {
				buffer.Buffer = []rune{}
				buffer.Length = 0
			}
			return readline.ENTER
		}
	}
	return readline.CONTINUE
}

func cmdBindKey(L Lua) int {
	keyTmp, ok := L.Get(-2).(lua.LString)
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("bindkey: key error"))
		return 2
	}
	key := strings.Replace(strings.ToUpper(string(keyTmp)), "-", "_", -1)
	switch value := L.Get(-1).(type) {
	case *lua.LFunction:
		if err := readline.BindKeyFunc(key, &KeyLuaFuncT{value}); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		} else {
			L.Push(lua.LTrue)
			return 1
		}
	default:
		val := L.ToString(-1)
		err := readline.BindKeySymbol(key, val)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		} else {
			L.Push(lua.LTrue)
			return 1
		}
	}
}
