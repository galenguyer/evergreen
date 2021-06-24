package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/galenguyer/evergreen/core"
)

func Cleanup() {
	log.Println("[cleanup] running cleanup")
	files, err := ioutil.ReadDir("./metadata")
	if err != nil {
		log.Println(err)
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(fmt.Sprintf("./metadata/%s", file.Name()))
		if err != nil {
			log.Println(err)
		} else {
			var metadata core.Metadata
			err = json.Unmarshal(data, &metadata)
			if err != nil {
				log.Println(err)
			} else {
				if time.Now().After(metadata.Expiry) {
					err = os.Remove(fmt.Sprintf("./uploads/%s", metadata.Filename))
					if err != nil {
						log.Printf("[ERROR][cleanup] error removing file ./uploads/%s: %s", metadata.Filename, err)
					} else {
						err = os.Remove(fmt.Sprintf("./metadata/%s", file.Name()))
						if err != nil {
							log.Printf("[ERROR][cleanup] error removing file ./metadata/%s: %s", file.Name(), err)
						} else {
							log.Printf("[cleanup] %s expired, removing\n", metadata.Filename)
						}
					}
				}
			}
		}
	}
}
