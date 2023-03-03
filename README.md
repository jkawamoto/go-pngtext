# go-pngtext
[![GoDoc](https://pkg.go.dev/badge/github.com/jkawamoto/go-pngtext)](https://pkg.go.dev/github.com/jkawamoto/go-pngtext)

Parse textual data from a PNG file.

## Usage
This code opens a file `image.png` and prints all textual data stored in the file:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jkawamoto/go-pngtext"
)

func main() {
	r, err := os.Open("image.png")
	if err != nil {
		log.Fatalf("failed to open a file: %v", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Fatalf("failed to close a file: %v", err)
		}
	}()

	res, err := pngtext.ParseTextualData(r)
	if err != nil {
		log.Fatalf("failed to parse textual data: %v", err)
	}

	for _, v:= range res{
        fmt.Printf("%v: %v\n", v.Keyword, v.Text)	
    }
}
```

## License
This software is released under the MIT License, see [LICENSE](https://github.com/jkawamoto/go-pngtext/blob/main/LICENSE).