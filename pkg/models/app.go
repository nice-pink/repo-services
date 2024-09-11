package models

// Exceptional Apps

type ExceptionalApps struct {
	Apps []App
}

type ExceptionalEnvDef struct {
	Name string
	Path string
	File string
}

// App

type App struct {
	Path      string
	Name      string
	Namespace string
	Tag       string
	Image     string
	File      string
	Env       string
	History   string
	Envs      []ExceptionalEnvDef
}
