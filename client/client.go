package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"golang.org/x/text/encoding/charmap"
)

func GetAnecdote(category int) string {
	client := http.DefaultClient
	requestUrl := fmt.Sprintf("http://rzhunemogu.ru/RandJSON.aspx?CType=%d", category)
	resp, err := client.Get(requestUrl)
	if err != nil {
		var urlErr *url.Error
		errors.As(err, &urlErr)
		log.Printf("Error: op: %v err: %v\n", urlErr.Op, urlErr.Err)
		return "Что-то пошло не так."
	}

	// make response normal...
	body := transformWindows1251ToUtf8(resp.Body)
	body = removeCRLF(body)

	return getAnecdote(body)
}

func transformWindows1251ToUtf8(r io.Reader) []byte {
	data, _ := io.ReadAll(r)
	dec := charmap.Windows1251.NewDecoder()
	out, _ := dec.Bytes(data)
	return out
}

func removeCRLF(in []byte) []byte {
	re := regexp.MustCompile(`\r?\n`)
	replaceString := url.QueryEscape("\n")
	out := re.ReplaceAll(in, []byte(replaceString))
	return out
}

func getAnecdote(data []byte) string {
	respJson := make(map[string]string)
	json.Unmarshal(data, &respJson)
	anecdote := respJson["content"]
	anecdote, _ = url.QueryUnescape(anecdote)
	return anecdote
}
