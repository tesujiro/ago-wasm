// +build js,wasm

package main

import (
	"fmt"
	"html"
	"strings"
	"syscall/js"

	"github.com/tesujiro/ago/parser"
	"github.com/tesujiro/ago/vm"
)

var (
	result = js.Global().Get("document").Call("getElementById", "result")
	input  = js.Global().Get("document").Call("getElementById", "input")
)

func writeCommand(s string) {
	result.Set("innerHTML", result.Get("innerHTML").String()+"<p class='command'>"+html.EscapeString(s)+"</p>")
	result.Set("scrollTop", result.Get("scrollHeight").Int())
}

func writeStdout(s string) {
	result.Set("innerHTML", result.Get("innerHTML").String()+"<p class='stdout'>"+html.EscapeString(s)+"</p>")
	result.Set("scrollTop", result.Get("scrollHeight").Int())
}

func writeStderr(s string) {
	result.Set("innerHTML", result.Get("innerHTML").String()+"<p class='stderr'>"+html.EscapeString(s)+"</p>")
	result.Set("scrollTop", result.Get("scrollHeight").Int())
}

func main() {
	env := vm.NewEnv([]string{})

	env.Define("print", func(a ...interface{}) {
		writeStdout(fmt.Sprint(a...))
	})
	env.Define("printf", func(a string, b ...interface{}) {
		writeStdout(fmt.Sprintf(a, b...))
	})

	var following bool
	var source string

	parser.EnableErrorVerbose()

	ch := make(chan string)

	input.Call("addEventListener", "keypress", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		if e.Get("keyCode").Int() != 13 {
			return nil
		}
		s := e.Get("target").Get("value").String()
		e.Get("target").Set("value", "")
		writeCommand(s)
		ch <- s
		return nil
	}))
	input.Set("disabled", false)
	result.Set("innerHTML", "")

	go func() {
		for {
			text, ok := <-ch
			if !ok {
				break
			}
			source += text
			if source == "" {
				continue
			}
			if source == "quit()" {
				break
			}

			stmts, err := parser.ParseSrc(source)

			if e, ok := err.(*parser.Error); ok {
				es := e.Error()
				if strings.HasPrefix(es, "syntax error: unexpected") {
					if strings.HasPrefix(es, "syntax error: unexpected $end,") {
						following = true
						continue
					}
				} else {
					if e.Pos.Column == len(source) && !e.Fatal {
						writeStderr(e.Error())
						following = true
						continue
					}
					if e.Error() == "unexpected EOF" {
						following = true
						continue
					}
				}
			}

			following = false
			source = ""
			var v interface{}

			if err == nil {
				v, err = vm.Run(stmts, env)
			}
			if err != nil {
				/*
					if e, ok := err.(*vm.Error); ok {
						writeStderr(fmt.Sprintf("%d:%d %s\n", e.Pos.Line, e.Pos.Column, err))
					} else if e, ok := err.(*parser.Error); ok {
						writeStderr(fmt.Sprintf("%d:%d %s\n", e.Pos.Line, e.Pos.Column, err))
					} else {
				*/
				writeStderr(err.Error())
				//}
				continue
			}

			writeStdout(fmt.Sprintf("%#v\n", v))
		}
	}()

	select {}

}
