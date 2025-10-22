// C:\_Projects_Go\AcousticLog\cmd\acousticlog\main.go

package main

import (
	"log"

	"acousticlog/internal/app"
	"acousticlog/internal/build"
)

func main() {
	cfg, err := app.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}
	build.PrintHeader(cfg.Timezone)
	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
