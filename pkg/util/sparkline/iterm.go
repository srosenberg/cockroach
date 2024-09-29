// Copyright 2024 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package sparkline

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
)

type ITermImage struct {
	img image.Image
}

// WriteTo implements the io.WriterTo interface, writing an iTerm 1337 escaped .
func (i *ITermImage) WriteTo(w io.Writer) (int64, error) {
	str, err := ITermEncodePNGToString(i.img, "[iTerm Image]")
	if err != nil {
		return 0, err
	}
	n, err := w.Write([]byte(str))
	return int64(n), err
}

func (i *ITermImage) String() string {
	str, err := ITermEncodePNGToString(i.img, "[iTerm Image]")
	if err != nil {
		return err.Error()
	}
	return str
}

func ITermEncodePNGToString(img image.Image, alt string) (str string, err error) {
	b := new(bytes.Buffer)
	err = png.Encode(b, img)
	if err != nil {
		return
	}
	bytes := b.Bytes()
	fmt.Printf("Encoded image size: %d\n", len(bytes))
	base64str := base64.StdEncoding.EncodeToString(bytes)
	str = fmt.Sprintf("\033]1337;File=inline=1;size=%d:%s\a%s", len(bytes), base64str, alt)
	return
}

func RenderImg(img image.Image, out io.Writer) error {
	itermImg := &ITermImage{img}
	if _, err := itermImg.WriteTo(out); err != nil {
		return err
	}
	return nil
}

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, _, err := image.Decode(f)
	return image, err
}
