package smska

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"
)

const MaxRetries = 3
const SmskaApiEndpoint = "https://smska.net/stubs/handler_api.php"

// Possible responses
const BadKey = `BAD_KEY`
const ErrorSql = `ERROR_SQL`

const BalancePat = `ACCESS_BALANCE:([0-9.,]+)`

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetBalance(balance *string) error {
	smskaApiKey := os.Getenv("SMSKA_API_KEY")
	action := `getBalance`

	url := fmt.Sprintf(`%s?api_key=%s&action=%s`, SmskaApiEndpoint, smskaApiKey, action)

	return retry(MaxRetries, time.Second, func() error {
		resp, err := http.Get(url)

		if err != nil {
			log.Printf("Error: %s", err)
			return err
		}

		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		serverResponse := string(body[:])

		if serverResponse == BadKey || serverResponse == ErrorSql {
			log.Printf("Error: %s", serverResponse)
			return err
		}

		re := regexp.MustCompile(BalancePat)

		*balance = re.FindStringSubmatch(serverResponse)[1]

		return nil
	})
}

// https://upgear.io/blog/simple-golang-retry-function/
func retry(attempts int, sleep time.Duration, f func() error) error {
	if err := f(); err != nil {
		if s, ok := err.(stop); ok {
			// Return the original error for later checking
			return s.error
		}

		if attempts--; attempts > 0 {
			// Add some randomness to prevent creating a Thundering Herd
			jitter := time.Duration(rand.Int63n(int64(sleep)))
			sleep = sleep + jitter/2

			time.Sleep(sleep)
			return retry(attempts, 2*sleep, f)
		}
		return err
	}

	return nil
}

type stop struct {
	error
}
