[![Go Reference](https://pkg.go.dev/badge/github.com/disgoorg/omit.svg)](https://pkg.go.dev/github.com/disgoorg/omit)
[![Go Report](https://goreportcard.com/badge/github.com/disgoorg/omit)](https://goreportcard.com/report/github.com/disgoorg/omit)
[![Go Version](https://img.shields.io/github/go-mod/go-version/disgoorg/omit)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/disgoorg/omit/blob/master/LICENSE)
[![omit Version](https://img.shields.io/github/v/tag/disgoorg/omit?label=release)](https://github.com/disgoorg/omit/releases/latest)
[![Discord](https://discord.com/api/guilds/817327181659111454/widget.png)](https://discord.gg/9tKpqXjYVC)

# omit

omit is a simple library to handle optional and nullable struct field values for json serialization and deserialization in golang.

## Installation

```bash
go get github.com/disgoorg/omit
```

## Usage

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/disgoorg/omit"
)

type User struct {
	ID    int                `json:"id"`
	Name  string             `json:"name"`
	Email omit.Omit[*string] `json:"email,omitzero"`
}

func main() {
	u := User{
		ID:    1,
		Name:  "John Doe",
		Email: omit.NewPtr("test@example.com"),
	}

	// Marshal
	data, err := json.Marshal(u)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
}
```

## Documentation

Documentation is wip and can be found under:

* [![Go Reference](https://pkg.go.dev/badge/github.com/disgoorg/omit.svg)](https://pkg.go.dev/github.com/disgoorg/omit)

## Troubleshooting

For help feel free to open an issue or reach out on [Discord](https://discord.gg/9tKpqXjYVC)

## Contributing

Contributions are welcomed but for bigger changes we recommend first reaching out via [Discord](https://discord.gg/9tKpqXjYVC) or create an issue to discuss your problems, intentions and ideas.

## License

Distributed under the [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE). See LICENSE for more information.
