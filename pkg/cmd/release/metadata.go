package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/storage"
)

type buildInfo struct {
	Tag       string `json:"tag"`
	SHA       string `json:"sha"`
	Timestamp string `json:"timestamp"`
}

func getBuildInfo(ctx context.Context, bucket string, obj string) (buildInfo, error) {
	__antithesis_instrumentation__.Notify(42779)
	client, err := storage.NewClient(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(42785)
		return buildInfo{}, fmt.Errorf("cannot create GCS client: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42786)
	}
	__antithesis_instrumentation__.Notify(42780)
	reader, err := client.Bucket(bucket).Object(obj).NewReader(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(42787)
		return buildInfo{}, fmt.Errorf("cannot create GCS reader: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42788)
	}
	__antithesis_instrumentation__.Notify(42781)
	defer func() {
		__antithesis_instrumentation__.Notify(42789)
		_ = reader.Close()
	}()
	__antithesis_instrumentation__.Notify(42782)

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		__antithesis_instrumentation__.Notify(42790)
		return buildInfo{}, fmt.Errorf("cannot read GCS object: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42791)
	}
	__antithesis_instrumentation__.Notify(42783)
	var info buildInfo
	if err := json.Unmarshal(data, &info); err != nil {
		__antithesis_instrumentation__.Notify(42792)
		return buildInfo{}, fmt.Errorf("cannot unmarshal metadata: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42793)
	}
	__antithesis_instrumentation__.Notify(42784)
	return info, nil
}

func publishReleaseCandidateInfo(
	ctx context.Context, next releaseInfo, bucket string, obj string,
) error {
	__antithesis_instrumentation__.Notify(42794)
	marshalled, err := json.MarshalIndent(next.buildInfo, "", "  ")
	if err != nil {
		__antithesis_instrumentation__.Notify(42799)
		return fmt.Errorf("cannot marshall buildInfo: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42800)
	}
	__antithesis_instrumentation__.Notify(42795)
	client, err := storage.NewClient(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(42801)
		return fmt.Errorf("cannot create storage client: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42802)
	}
	__antithesis_instrumentation__.Notify(42796)
	wc := client.Bucket(bucket).Object(obj).NewWriter(ctx)
	wc.ContentType = "application/json"
	if _, err := wc.Write(marshalled); err != nil {
		__antithesis_instrumentation__.Notify(42803)
		return fmt.Errorf("cannot write to bucket: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42804)
	}
	__antithesis_instrumentation__.Notify(42797)
	if err := wc.Close(); err != nil {
		__antithesis_instrumentation__.Notify(42805)
		return fmt.Errorf("cannot close storage writer filehandle: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42806)
	}
	__antithesis_instrumentation__.Notify(42798)
	return nil
}
