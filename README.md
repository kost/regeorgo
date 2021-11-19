# regeorgo

Implementation of regeorg tunnel in golang (victim side).

### Usage: binary

build binary:
```
git clone https://github.com/kost/regeorgo
cd regeorgo/bin
go get
go build
```

run binary:

```
./regeorgo
```

it will listen on port 8111 for reGeorgSocksProxy.py to connect to.

### Usage in your code

If you want to embed regeorgo in your code/executable:
```
package main

import (
	"net/http"
	"github.com/kost/regeorgo"
)

func main() {
	// initialize regeorgo
	gh := &regeorgo.GeorgHandler{}
	gh.initHandler()

	// use it as standard handler for http
	http.HandleFunc("/regeorgo", gh.regHandler)
	http.ListenAndServe(":8111", nil)
}
```

### Requirement

You need to have:

- golang

# License

Distributed under MIT license

# Credits

Initial development of regeorgo in Go by kost.

## Links

- Original regeorg: https://github.com/sensepost/reGeorg

- Improved regeorg: https://github.com/kost/regeorg

- Refactored regeorg (not compatible with this): https://github.com/L-codes/Neo-reGeorg

