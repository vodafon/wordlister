package wordlister

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/jinzhu/inflection"
	"github.com/microcosm-cc/bluemonday"
)

var (
	clearSymbols   = []string{".", "!", "?", ",", "(", ")", "[", "]", "\"", "'", ";", ":", "{", "}"}
	invalidSymbols = []string{"&", "#", "|", "/", "www", "http", "%", "@", "'", "’", "”", "+", "=", "0x", "x0", "_", "\\"}
	stopWords      = []string{
		"a", "the", "is", "are", "if", "what", "where", "of", "you", "me", "he", "she", "it", "to", "or", "can",
		"both", "and", "i", "from", "use", "let", "for", "add", "in", "be", "get", "either", "cannot", "do",
		"there", "no", "yes", "how", "on", "same", "in", "any", "so", "allow", "up", "all", "own", "which", "per",
		"not", "with", "within", "we", "then", "than", "they", "this", "through", "when", "will", "because",
	}
)

type Wordlist struct {
	*sync.Mutex
	wl         map[string]int64
	htmlPolicy *bluemonday.Policy
	lemmas     map[string]string
	sum        int64
}

func NewWordlist() *Wordlist {
	_, packetfile, _, _ := runtime.Caller(0)
	packetdir := path.Dir(packetfile)
	return &Wordlist{
		Mutex:      &sync.Mutex{},
		wl:         make(map[string]int64),
		htmlPolicy: bluemonday.NewPolicy(),
		lemmas:     loadLemmas(path.Join(packetdir, "data/lemmatization-en.txt")),
	}
}

func loadLemmas(path string) map[string]string {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(file))
	lemmas := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Split(line, "\t")
		if len(words) != 2 {
			words = strings.Split(line, " ")
			if len(words) != 2 {
				continue
			}
		}
		lemmas[words[1]] = words[0]
	}
	return lemmas
}

func (obj *Wordlist) FromHTML(page []byte) error {
	htmlB := obj.htmlPolicy.SanitizeBytes(page)
	scanner := bufio.NewScanner(bytes.NewReader(htmlB))
	scanner.Split(bufio.ScanWords)
	obj.Scan(scanner)
	return nil
}

func (obj *Wordlist) List() []string {
	obj.Lock()
	defer obj.Unlock()
	list := []string{}
	for k, _ := range obj.wl {
		list = append(list, k)
		k2 := inflection.Plural(k)
		if k != k2 {
			list = append(list, k2)
		}
	}
	return list
}

func (obj *Wordlist) SeqDict() map[string]float64 {
	obj.Lock()
	defer obj.Unlock()
	seq := make(map[string]float64)
	for k, v := range obj.wl {
		seq[k] = float64(v) / float64(obj.sum)
	}
	return seq
}

func (obj *Wordlist) Scan(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		text := strings.ToLower(scanner.Text())
		if IsInvalid(text) {
			continue
		}
		text = inflection.Singular(text)
		text = clearWord(text)
		if len(text) < 2 {
			continue
		}
		text = obj.lemmatization(text)
		if IsInSlice(text, stopWords) {
			continue
		}
		obj.Lock()
		obj.wl[text] += 1
		obj.sum += 1
		obj.Unlock()
	}
	return nil
}

func (obj *Wordlist) lemmatization(word string) string {
	v, ok := obj.lemmas[word]
	if ok {
		return v
	}
	return word
}

func IsInvalid(word string) bool {
	if len(word) > 20 || len(word) != len([]byte(word)) || len(strconv.Quote(word)) != len(strconv.QuoteToASCII(word)) {
		return true
	}
	for _, r := range word {
		if !unicode.IsPrint(r) {
			return true
		}
	}
	for _, s := range invalidSymbols {
		if strings.Contains(word, s) {
			return true
		}
	}
	return false
}

func IsInSlice(word string, sl []string) bool {
	for _, s := range sl {
		if word == s {
			return true
		}
	}
	return false
}

func clearWord(word string) string {
	for _, s := range clearSymbols {
		word = strings.ReplaceAll(word, s, "")
	}
	return word
}
