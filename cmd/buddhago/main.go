package main

import "../../pkg/buddhabrot"
import "image/png"
import "os"
import "flag"
import "runtime"

var filename = flag.String("output-filename", "render.png", "name of the output file")

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	image := buddhabrot.Generate()
	file, err := os.Create(*filename)
	if (err != nil) {
		panic(err)
	}
	png.Encode(file, image)
	defer file.Close()
}
