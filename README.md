# Api for smska.net

## Constraints

Api of Smska.net doesn't allow to retreive full text from sms

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

    // Max retries - 12 every 1 second
	smska.GetStatus(phoneNumber.Id, &code)
	log.Print(code)
}
```
