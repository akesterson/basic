package main

type BasicFunctionDef struct {
	arglist *BasicASTLeaf
	expression *BasicASTLeaf
	lineno int64
	name string
	environment BasicEnvironment
	runtime *BasicRuntime
}
