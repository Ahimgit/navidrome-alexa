package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets/*
var assets embed.FS

type EmbedFileSystem struct {
	http.FileSystem
}

func NewEmbedFileSystem() http.FileSystem {
	sub, _ := fs.Sub(assets, "assets")
	return &EmbedFileSystem{FileSystem: http.FS(sub)}
}

func (e *EmbedFileSystem) Open(name string) (http.File, error) {
	return e.FileSystem.Open(name)
}
