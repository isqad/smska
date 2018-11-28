# Api for smska.net

## Usage

```go
package main

import (
	"github.com/isqad/smska"
	"log"
)

func main() {
	var phoneNumber smska.SmskaNumber
	var code string

	smska.GetNumber(`oz`, &phoneNumber)
	log.Print(phoneNumber)

    // Max retries - 10 every 5 seconds
	smska.GetStatus(phoneNumber.Id, &code)
	log.Print(code)
}
```
