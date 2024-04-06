package client

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/text/encoding/charmap"
)

// GetAnecdote returns anecdote from api
//
// http://rzhunemogu.ru/RandJSON.aspx?CType={categoryId} where categoryId - number from 1 to 18.
func GetAnecdote(category int) string {
	client := http.DefaultClient
	requestUrl := fmt.Sprintf("http://rzhunemogu.ru/RandJSON.aspx?CType=%d", category)
	resp, err := client.Get(requestUrl)
	if err != nil {
		var urlErr *url.Error
		errors.As(err, &urlErr)
		log.Printf("Error: op: %v err: %v\n", urlErr.Op, urlErr.Err)
		return ""
	}

	// make response normal...
	body := transformWindows1251ToUtf8(resp.Body)
	anecdote := GetContent(body)

	if anecdote != "" {
		return anecdote
	}
	return ""
}

// transformWindows1251ToUtf8 converts windows1251 to utf8
func transformWindows1251ToUtf8(r io.Reader) []byte {
	data, _ := io.ReadAll(r)
	dec := charmap.Windows1251.NewDecoder()
	out, _ := dec.Bytes(data)
	return out
}

// GetContent returns content from response
//
// data is in form `{"content":"text"}`
// json.Unmarshal work bad cause it's very cringe api with bad response
// in response we have:
// 1. unescaped `"` in value
// 2. \n in value of "content"
//
// so I just slice it by template
// it's more stable
func GetContent(data []byte) string {
	return string(data)[12 : len(string(data))-2]
}
