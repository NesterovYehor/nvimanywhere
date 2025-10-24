package webfs

import "embed"

//go:embed web/templates/*
var TmplFS embed.FS

//go:embed web/static
var StaticFS embed.FS
