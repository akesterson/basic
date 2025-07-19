package main

type BasicFunctionDef struct {
	arglist *BasicASTLeaf
	expression *BasicASTLeaf
	name string
	environment BasicEnvironment
	runtime *BasicRuntime
}
