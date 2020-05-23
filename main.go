// +build js,wasm

package main

import (
	"fmt"
	"html"
	"reflect"
	"strings"
	"syscall/js"

	"github.com/tesujiro/ago/lib"
	"github.com/tesujiro/ago/parser"
	"github.com/tesujiro/ago/vm"
)

var (
	result = js.Global().Get("document").Call("getElementById", "result")
	input  = js.Global().Get("document").Call("getElementById", "input")
	infile = js.Global().Get("document").Call("getElementById", "infile")
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
	env = lib.Import(env)
	env.SetFS(" ")
	vm.SetGlobalVariables()

	printf := func(a string, b ...interface{}) {
		writeStdout(fmt.Sprintf(a, b...))
	}
	env.Define("printf", reflect.ValueOf(printf))

	var following bool
	var source string

	parser.EnableErrorVerbose()

	ch := make(chan []string)

	currBuffer := 0
	sourceBuffer := []string{}

	//input.Call("addEventListener", "keypress", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	input.Call("addEventListener", "keyup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		//fmt.Printf("keyCode=%v\n", e.Get("keyCode").Int())
		switch e.Get("keyCode").Int() {
		case 13: // CR
			src := e.Get("target").Get("value").String()
			e.Get("target").Set("value", "")
			writeCommand(src)
			sourceBuffer = append(sourceBuffer, src)
			currBuffer = len(sourceBuffer)
			//fmt.Println(infile.Get("value").String())
			in := infile.Get("value").String()
			ch <- []string{src, in}
		case 38: // UP arrow
			if currBuffer > 0 {
				currBuffer--
				src := sourceBuffer[currBuffer]
				//currBuffer = len(sourceBuffer)
				//writeCommand(s)
				e.Get("target").Set("value", src)
			}
		case 40: // DOWN arrow
			if currBuffer < len(sourceBuffer)-1 {
				currBuffer++
				src := sourceBuffer[currBuffer]
				//currBuffer = len(sourceBuffer)
				//writeCommand(src)
				e.Get("target").Set("value", src)
			}
		}
		return nil
	}))
	input.Set("disabled", false)
	result.Set("innerHTML", "")

	go func() {
		for {
			texts, ok := <-ch
			if !ok {
				break
			}
			text := texts[0]
			inputStr := texts[1]

			source += text
			if source == "" {
				continue
			}
			if source == "exit" {
				break
			}

			rules, err := parser.ParseSrc(source)
			//parser.Dump(rules)

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
				env.SetPseudoStdin(inputStr)
				v, err = vm.Run(rules, env)
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
