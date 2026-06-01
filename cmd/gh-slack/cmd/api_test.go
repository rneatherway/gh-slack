package cmd

import (
	"io"
	"os"
	"testing"
)

func TestMapFields(t *testing.T) {
	fields, err := mapFields([]string{"channel=C123", "text=hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fields["channel"] != "C123" || fields["text"] != "hello" {
		t.Fatalf("unexpected fields: %+v", fields)
	}
}

func TestMapFieldsMissingValue(t *testing.T) {
	if _, err := mapFields([]string{"channel="}); err == nil {
		t.Fatal("expected error")
	}
}

func TestFileParamFromFlag(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "photo-*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.WriteString("image"); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	param, file, err := fileParamFromFlag("image=@" + tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer file.Close()

	if param.Fieldname != "image" {
		t.Fatalf("unexpected field name: %s", param.Fieldname)
	}
	if param.Filename == "" {
		t.Fatal("expected filename")
	}
	data, err := io.ReadAll(param.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "image" {
		t.Fatalf("unexpected file data: %q", data)
	}
}

func TestFileParamFromFlagInvalidFormat(t *testing.T) {
	for _, fileFlag := range []string{"image=photo.jpg", "image=@", "=@photo.jpg", "image"} {
		if _, _, err := fileParamFromFlag(fileFlag); err == nil {
			t.Fatalf("expected error for %q", fileFlag)
		}
	}
}
