// https://smska.net/?mode=info&ul=api
package smska

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

const MaxRetries = 3
const SmskaApiEndpoint = "https://smska.net/stubs/handler_api.php"

// Possible responses
const BadKey = `BAD_KEY`
const BadAction = `BAD_ACTION`
const NoAction = `NO_ACTION`
const BadService = `BAD_SERVICE`
const ErrorSql = `ERROR_SQL`
const NoNumbers = `NO_NUMBERS`
const NoBalance = `NO_BALANCE`

const StatusWait = `STATUS_WAIT_CODE`

const BalancePat = `ACCESS_BALANCE:([0-9.,]+)`
const PhonePat = `ACCESS_NUMBER:(.+):(.+)`
const ActivationCodePat = `STATUS_OK:(.+)`

// Id - number of operation
// Phone - phone number
type SmskaNumber struct {
	Id    string
	Phone string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetBalance(balance *float64) error {
	smskaApiKey := os.Getenv("SMSKA_API_KEY")
	action := `getBalance`

	url := fmt.Sprintf(`%s?api_key=%s&action=%s`, SmskaApiEndpoint, smskaApiKey, action)

	return retry(MaxRetries, time.Second, func() error {
		resp, err := http.Get(url)

		if err != nil {
			log.Fatal(err)
			return err
		}

		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		serverResponse := string(body[:])

		if serverResponse == BadKey || serverResponse == ErrorSql {
			log.Fatal(serverResponse)
			return err
		}

		re := regexp.MustCompile(BalancePat)

		money, err := strconv.ParseFloat(re.FindStringSubmatch(serverResponse)[1], 64)

		if err != nil {
			log.Fatal(err)
			return err
		}

		*balance = money

		return nil
	})
}

func GetNumber(service string, smskaNumber *SmskaNumber) error {
	smskaApiKey := os.Getenv("SMSKA_API_KEY")
	action := `getNumber`

	url := fmt.Sprintf(`%s?api_key=%s&action=%s&service=%s&operator=any`,
		SmskaApiEndpoint, smskaApiKey, action, service)

	return retry(MaxRetries, time.Second, func() error {
		resp, err := http.Get(url)

		if err != nil {
			log.Fatal(err)
			return err
		}

		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		serverResponse := string(body[:])

		if serverResponse == BadKey ||
			serverResponse == ErrorSql ||
			serverResponse == BadAction ||
			serverResponse == NoNumbers ||
			serverResponse == NoBalance ||
			serverResponse == BadService {
			log.Fatal(serverResponse)
			return nil
		}

		re := regexp.MustCompile(PhonePat)

		result := re.FindStringSubmatch(serverResponse)

		*smskaNumber = SmskaNumber{Id: result[1], Phone: result[2]}

		return nil
	})
}

func GetStatus(id string, code *string) error {
	smskaApiKey := os.Getenv("SMSKA_API_KEY")
	action := `getStatus`

	url := fmt.Sprintf(`%s?api_key=%s&action=%s&id=%s`, SmskaApiEndpoint, smskaApiKey, action, id)

	return retry(12, time.Second, func() error {
		resp, err := http.Get(url)

		if err != nil {
			log.Fatal(err)
			return err
		}

		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		serverResponse := string(body[:])

		log.Print(serverResponse)

		if serverResponse == BadKey ||
			serverResponse == ErrorSql ||
			serverResponse == BadAction {
			log.Fatal(serverResponse)
			return nil
		}

		if serverResponse == StatusWait {
			log.Print("Try again...")
			return errors.New(serverResponse)
		}

		re := regexp.MustCompile(ActivationCodePat)

		result := re.FindStringSubmatch(serverResponse)

		log.Printf("Got: %s", result)

		*code = result[1]

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
