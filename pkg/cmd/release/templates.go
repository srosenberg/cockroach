package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"text/template"
)

func templateToText(templateText string, args interface{}) (string, error) {
	__antithesis_instrumentation__.Notify(42904)
	templ, err := template.New("").Parse(templateText)
	if err != nil {
		__antithesis_instrumentation__.Notify(42907)
		return "", fmt.Errorf("cannot parse template: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42908)
	}
	__antithesis_instrumentation__.Notify(42905)

	var buf bytes.Buffer
	err = templ.Execute(&buf, args)
	if err != nil {
		__antithesis_instrumentation__.Notify(42909)
		return "", fmt.Errorf("cannot execute template: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42910)
	}
	__antithesis_instrumentation__.Notify(42906)
	return buf.String(), nil
}

func templateToHTML(templateText string, args interface{}) (string, error) {
	__antithesis_instrumentation__.Notify(42911)
	templ, err := htmltemplate.New("").Parse(templateText)
	if err != nil {
		__antithesis_instrumentation__.Notify(42914)
		return "", fmt.Errorf("cannot parse template: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42915)
	}
	__antithesis_instrumentation__.Notify(42912)

	var buf bytes.Buffer
	err = templ.Execute(&buf, args)
	if err != nil {
		__antithesis_instrumentation__.Notify(42916)
		return "", fmt.Errorf("cannot execute template: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42917)
	}
	__antithesis_instrumentation__.Notify(42913)
	return buf.String(), nil
}
