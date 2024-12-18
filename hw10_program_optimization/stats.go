package hw10programoptimization

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/mailru/easyjson"
)

//easyjson:json
type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

type SearchResult struct {
	Name  string
	Count int
}

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	errChan := make(chan error, 1)
	stopChan := make(chan struct{}, 1)
	result := make(DomainStat)

	bufReader := bufio.NewReader(r)

	userChan := getUsers(bufReader, stopChan, errChan)
	searchResult := domainSearcher(userChan, domain, stopChan, errChan)

	doneChan := counter(searchResult, stopChan, result)

	select {
	case <-doneChan:
		return result, nil
	case err := <-errChan:
		close(stopChan)
		return nil, err
	}
}

func counter(searchCh <-chan SearchResult, stopChan <-chan struct{}, result DomainStat) <-chan struct{} {
	doneChan := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case searchResult, ok := <-searchCh:
				if ok {
					result[searchResult.Name] += searchResult.Count
				} else {
					// work done
					doneChan <- struct{}{}
				}
			case <-stopChan:
				return
			}
		}
	}()

	return doneChan
}

func getUsers(r io.Reader,
	stopChan <-chan struct{},
	errChan chan<- error,
) <-chan User {
	scanner := bufio.NewScanner(r)
	userChan := make(chan User)

	go func() {
		defer close(userChan)
		for scanner.Scan() {
			select {
			// stopChan (on error in countDomains())
			case <-stopChan:
				return
			default:
			}
			tokenText := scanner.Bytes()
			var user User
			err := easyjson.Unmarshal(tokenText, &user)
			if err != nil {
				errChan <- err
				return
			}
			userChan <- user
		}
	}()

	return userChan
}

func domainSearcher(userChan <-chan User,
	domain string,
	stopChan <-chan struct{},
	errChan chan<- error,
) <-chan SearchResult {
	searchResult := make(chan SearchResult, 1)
	go func() {
		defer close(searchResult)

		regCache := make(map[string]*regexp.Regexp)
		var err error

		for user := range userChan {
			select {
			// stopChan (on error in countDomains())
			case <-stopChan:
				return
			default:
			}

			reg := regCache[domain]
			if reg == nil {
				reg, err = regexp.Compile("\\." + domain)
				if err != nil {
					errChan <- err
					return
				}

				regCache[domain] = reg
			}

			if reg.MatchString(user.Email) {
				name := strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])
				searchResult <- SearchResult{
					Name:  name,
					Count: 1,
				}
			}
		}
	}()

	return searchResult
}
