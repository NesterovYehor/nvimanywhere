package config

type HTTP struct {
	Host string
	Port string
}


type Container struct {
	Image   string 
	Workdir string
}

type Config struct {
	Htpp      *HTTP
	Container *Container
	BaseDir string
}
