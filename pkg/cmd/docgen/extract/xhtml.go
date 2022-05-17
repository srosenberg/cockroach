package extract

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func XHTMLtoHTML(r io.Reader) (string, error) {
	__antithesis_instrumentation__.Notify(39751)
	b := new(bytes.Buffer)
	z := html.NewTokenizer(r)
	for {
		__antithesis_instrumentation__.Notify(39756)
		tt := z.Next()
		if tt == html.ErrorToken {
			__antithesis_instrumentation__.Notify(39760)
			err := z.Err()
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(39762)
				break
			} else {
				__antithesis_instrumentation__.Notify(39763)
			}
			__antithesis_instrumentation__.Notify(39761)
			return "", z.Err()
		} else {
			__antithesis_instrumentation__.Notify(39764)
		}
		__antithesis_instrumentation__.Notify(39757)
		t := z.Token()
		switch t.Type {
		case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
			__antithesis_instrumentation__.Notify(39765)
			idx := strings.IndexByte(t.Data, ':')
			t.Data = t.Data[idx+1:]
		default:
			__antithesis_instrumentation__.Notify(39766)
		}
		__antithesis_instrumentation__.Notify(39758)
		var na []html.Attribute
		for _, a := range t.Attr {
			__antithesis_instrumentation__.Notify(39767)
			if strings.HasPrefix(a.Key, "xmlns") {
				__antithesis_instrumentation__.Notify(39769)
				continue
			} else {
				__antithesis_instrumentation__.Notify(39770)
			}
			__antithesis_instrumentation__.Notify(39768)
			na = append(na, a)
		}
		__antithesis_instrumentation__.Notify(39759)
		t.Attr = na
		b.WriteString(t.String())
	}
	__antithesis_instrumentation__.Notify(39752)

	doc, err := goquery.NewDocumentFromReader(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(39771)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(39772)
	}
	__antithesis_instrumentation__.Notify(39753)
	defs := doc.Find("defs")
	dhtml, err := defs.First().Html()
	if err != nil {
		__antithesis_instrumentation__.Notify(39773)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(39774)
	}
	__antithesis_instrumentation__.Notify(39754)
	doc.Find("head").AppendHtml(dhtml)
	defs.Remove()
	doc.Find("svg").First().Remove()
	doc.Find("meta[http-equiv]").Remove()
	doc.Find("head").PrependHtml(`<meta charset="UTF-8">`)
	doc.Find("a[name]:not([href])").Each(func(_ int, s *goquery.Selection) {
		__antithesis_instrumentation__.Notify(39775)
		name, exists := s.Attr("name")
		if !exists {
			__antithesis_instrumentation__.Notify(39777)
			return
		} else {
			__antithesis_instrumentation__.Notify(39778)
		}
		__antithesis_instrumentation__.Notify(39776)
		s.SetAttr("href", "#"+name)
	})
	__antithesis_instrumentation__.Notify(39755)
	s, err := doc.Find("html").Html()
	s = "<!DOCTYPE html><html>" + s + "</html>"
	return s, err
}

func Tag(r io.Reader, tag string) (string, error) {
	__antithesis_instrumentation__.Notify(39779)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		__antithesis_instrumentation__.Notify(39782)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(39783)
	}
	__antithesis_instrumentation__.Notify(39780)
	node := doc.Find(tag).Get(0)
	var b bytes.Buffer
	if err := html.Render(&b, node); err != nil {
		__antithesis_instrumentation__.Notify(39784)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(39785)
	}
	__antithesis_instrumentation__.Notify(39781)
	return b.String(), nil
}

func InnerTag(r io.Reader, tag string) (string, error) {
	__antithesis_instrumentation__.Notify(39786)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		__antithesis_instrumentation__.Notify(39788)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(39789)
	}
	__antithesis_instrumentation__.Notify(39787)
	return doc.Find(tag).Html()
}
