package main

import "go/parser"

// ParseFile is to reduce diff with the main compiler
var ParseFile = parser.ParseFile
