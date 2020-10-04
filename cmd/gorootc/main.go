// SPDX-FileCopyrightText: 2020 Carson Black <uhhadd@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"

	. "github.com/dave/jennifer/jen"
)

func TypeToString(field ast.Expr) string {
	switch kind := field.(type) {
	case *ast.Ident:
		return kind.Name
	case *ast.ArrayType:
		return "[]" + TypeToString(kind.Elt)
	}
	panic("unhandled case")
}

func GetBody(fset *token.FileSet, funcDecl *ast.FuncDecl) string {
	start := fset.PositionFor(funcDecl.Body.Pos(), false)
	end := fset.PositionFor(funcDecl.Body.End(), false)

	fi, err := os.Open(start.Filename)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	buf := make([]byte, end.Offset-start.Offset)
	_, err = fi.ReadAt(buf, int64(start.Offset))
	if err != nil {
		log.Fatalf("%+v", err)
	}

	return string(buf)
}

func main() {
	fset := token.FileSet{}
	set, err := parser.ParseDir(&fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	for _, pkg := range set {
		for _, files := range pkg.Files {
			for _, funktion := range files.Decls {
				if val, ok := funktion.(*ast.FuncDecl); ok {
					// properties, err := ReadComments(val.Doc)
					// if err != nil {
					// 	log.Fatalf("%+v", err)
					// }

					worker := Id("worker").Op(":=").Func().Params()
					workerCall := Id("worker").Call()
					workerStructCall := Id("returnValue").Op(":=").Id("ReturnKind").Block()
					returnKind := Type().Id("ReturnKind").Struct()
					if val.Type.Results != nil {
						var ids []Code
						var names []Code
						var structMemmbs []Code
						for _, param := range val.Type.Results.List {
							if len(param.Names) < 0 {
								log.Fatalf("Named returns are required for action worker functions")
							}
							for _, name := range param.Names {
								ids = append(ids, Id(TypeToString(param.Type)))
								names = append(names, Id(name.Name))
								structMemmbs = append(structMemmbs, Id(name.Name).Id(TypeToString(param.Type)))
							}
						}
						workerCall = List(names...).Op(":=").Id("worker").Call()
						workerStructCall = Id("returnValue").Op(":=").Id("ReturnKind").Id("{").List(names...).Id("}")
						returnKind = Type().Id("ReturnKind").Struct(structMemmbs...)
						worker.Params(ids...)
					}
					worker.Id(GetBody(&fset, val))

					readInKind := Type().Id("ReadInKind").Struct()
					readInVar := Var().Id("readIn").Id("ReadInKind")
					readInCall := Id("err").Op("=").Qual("encoding/json", "Unmarshal").Call(Id("data"), Id("&readIn"))
					destructureCall := Null()
					if val.Type.Params != nil {
						var structMemmbs []Code
						for _, param := range val.Type.Params.List {
							if len(param.Names) < 0 {
								log.Fatalf("Named parameters are required for action worker functions")
							}
							for _, name := range param.Names {
								structMemmbs = append(structMemmbs, Id(name.Name).Id(TypeToString(param.Type)))
								destructureCall.Add(Id(name.Name).Op(":=").Id("readIn." + name.Name))
								destructureCall.Add(Id(";"))
							}
						}
						readInKind = Type().Id("ReadInKind").Struct(structMemmbs...)
					}

					f := NewFile("main")

					for _, imp := range files.Imports {
						if imp.Name != nil {
							f.Id("import").Id(imp.Name.Name).Id(imp.Path.Value)
						} else {
							f.Id("import").Id(imp.Path.Value)
						}
					}

					f.Add(returnKind)
					f.Add(readInKind)
					f.Func().Id("main").Params().Block(
						List(Id("data"), Err()).Op(":=").Qual("io/ioutil", "ReadAll").Call(Qual("os", "Stdin")),
						If(Err().Op("!=").Nil()).Block(
							Qual("log", "Fatalf").Call(Lit("Failed to receive input from stdin")),
						),
						readInVar,
						readInCall,
						If(Err().Op("!=").Nil()).Block(
							Qual("log", "Fatalf").Call(Lit("Failed to receive input from stdin")),
						),
						destructureCall,
						worker,
						workerCall,
						workerStructCall,
						List(Id("returnMarshalled"), Err()).Op(":=").Qual("encoding/json", "Marshal").Call(Id("returnValue")),
						If(Err().Op("!=").Nil()).Block(
							Qual("log", "Fatalf").Call(Lit("Failed to format value for stdout")),
						),
						Println(String().Call(Id("returnMarshalled"))),
					)
					fmt.Printf("%#v", f)
				}
			}
		}
	}
}
