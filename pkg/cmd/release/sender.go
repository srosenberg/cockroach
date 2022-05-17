package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	htmltemplate "html/template"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/jordan-wright/email"
)

const (
	templatePrefixPickSHA           = "pick-sha"
	templatePrefixPostBlockers      = "post-blockers"
	templatePrefixPostBlockersAlpha = "post-blockers.alpha"
)

type messageDataPickSHA struct {
	Version          string
	SHA              string
	TrackingIssue    string
	TrackingIssueURL htmltemplate.URL
	DiffURL          htmltemplate.URL
}

type ProjectBlocker struct {
	ProjectName string
	NumBlockers int
}

type messageDataPostBlockers struct {
	Version       string
	PrepDate      string
	ReleaseDate   string
	TotalBlockers int
	BlockersURL   string
	ReleaseBranch string
	BlockerList   []ProjectBlocker
}

type postBlockerTemplateArgs struct {
	BackportsUseBackboard       bool
	BackportsWeeklyTriageReview bool
}

type message struct {
	Subject  string
	TextBody string
	HTMLBody string
}

func loadTemplate(templatesDir, template string) (string, error) {
	__antithesis_instrumentation__.Notify(42865)
	file, err := os.ReadFile(filepath.Join(templatesDir, template))
	if err != nil {
		__antithesis_instrumentation__.Notify(42867)
		return "", fmt.Errorf("loadTemplate %s: %w", template, err)
	} else {
		__antithesis_instrumentation__.Notify(42868)
	}
	__antithesis_instrumentation__.Notify(42866)
	return string(file), nil
}

func newMessage(templatesDir string, templatePrefix string, data interface{}) (*message, error) {
	__antithesis_instrumentation__.Notify(42869)
	subjectTemplate, err := loadTemplate(templatesDir, templatePrefix+".subject")
	if err != nil {
		__antithesis_instrumentation__.Notify(42876)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(42877)
	}
	__antithesis_instrumentation__.Notify(42870)
	subject, err := templateToText(subjectTemplate, data)
	if err != nil {
		__antithesis_instrumentation__.Notify(42878)
		return nil, fmt.Errorf("templateToText %s: %w", templatePrefix+".subject", err)
	} else {
		__antithesis_instrumentation__.Notify(42879)
	}
	__antithesis_instrumentation__.Notify(42871)

	textTemplate, err := loadTemplate(templatesDir, templatePrefix+".txt")
	if err != nil {
		__antithesis_instrumentation__.Notify(42880)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(42881)
	}
	__antithesis_instrumentation__.Notify(42872)
	text, err := templateToText(textTemplate, data)
	if err != nil {
		__antithesis_instrumentation__.Notify(42882)
		return nil, fmt.Errorf("templateToText %s: %w", templatePrefix+".txt", err)
	} else {
		__antithesis_instrumentation__.Notify(42883)
	}
	__antithesis_instrumentation__.Notify(42873)

	htmlTemplate, err := loadTemplate(templatesDir, templatePrefix+".gohtml")
	if err != nil {
		__antithesis_instrumentation__.Notify(42884)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(42885)
	}
	__antithesis_instrumentation__.Notify(42874)
	html, err := templateToHTML(htmlTemplate, data)
	if err != nil {
		__antithesis_instrumentation__.Notify(42886)
		return nil, fmt.Errorf("templateToHTML %s: %w", templatePrefix+".gohtml", err)
	} else {
		__antithesis_instrumentation__.Notify(42887)
	}
	__antithesis_instrumentation__.Notify(42875)

	return &message{
		Subject:  subject,
		TextBody: text,
		HTMLBody: html,
	}, nil
}

type sendOpts struct {
	templatesDir string
	host         string
	port         int
	user         string
	password     string
	from         string
	to           []string
}

func sendMailPostBlockers(args messageDataPostBlockers, opts sendOpts) error {
	__antithesis_instrumentation__.Notify(42888)
	templatePrefix := templatePrefixPostBlockers

	backportsUseBackboard := false
	backportsWeeklyTriageReview := false

	switch {
	case strings.Contains(args.Version, "-alpha."):
		__antithesis_instrumentation__.Notify(42891)

		templatePrefix = templatePrefixPostBlockersAlpha
	case
		strings.Contains(args.Version, "-beta."),
		strings.Contains(args.Version, "-rc."):
		__antithesis_instrumentation__.Notify(42892)
		backportsWeeklyTriageReview = true
	default:
		__antithesis_instrumentation__.Notify(42893)
		backportsUseBackboard = true
	}
	__antithesis_instrumentation__.Notify(42889)

	data := struct {
		Args     messageDataPostBlockers
		Template postBlockerTemplateArgs
	}{
		Args: args,
		Template: postBlockerTemplateArgs{
			BackportsUseBackboard:       backportsUseBackboard,
			BackportsWeeklyTriageReview: backportsWeeklyTriageReview,
		},
	}
	msg, err := newMessage(opts.templatesDir, templatePrefix, data)
	if err != nil {
		__antithesis_instrumentation__.Notify(42894)
		return fmt.Errorf("newMessage: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42895)
	}
	__antithesis_instrumentation__.Notify(42890)
	return sendmail(msg, opts)
}

func sendMailPickSHA(args messageDataPickSHA, opts sendOpts) error {
	__antithesis_instrumentation__.Notify(42896)
	msg, err := newMessage(opts.templatesDir, templatePrefixPickSHA, args)
	if err != nil {
		__antithesis_instrumentation__.Notify(42898)
		return fmt.Errorf("newMessage: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42899)
	}
	__antithesis_instrumentation__.Notify(42897)
	return sendmail(msg, opts)
}

var sendmail = func(content *message, smtpOpts sendOpts) error {
	__antithesis_instrumentation__.Notify(42900)

	e := &email.Email{
		To:      smtpOpts.to,
		From:    smtpOpts.from,
		Subject: content.Subject,
		Text:    []byte(content.TextBody),
		HTML:    []byte(content.HTMLBody),
		Headers: textproto.MIMEHeader{},
	}
	emailAuth := smtp.PlainAuth("", smtpOpts.user, smtpOpts.password, smtpOpts.host)
	addr := fmt.Sprintf("%s:%d", smtpOpts.host, smtpOpts.port)
	if err := e.Send(addr, emailAuth); err != nil {
		__antithesis_instrumentation__.Notify(42902)
		return fmt.Errorf("cannot send email: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42903)
	}
	__antithesis_instrumentation__.Notify(42901)
	return nil
}
