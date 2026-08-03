package config

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestStripJSONC(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want any // unmarshaled expected value
	}{
		{
			name: "line comments",
			src: `{
  // comment before
  "a": 1, // trailing comment
  "b": 2
}`,
			want: map[string]any{"a": float64(1), "b": float64(2)},
		},
		{
			name: "block comments",
			src: `{
  /* block */
  "a": /* mid */ 1,
  "b": 2
}`,
			want: map[string]any{"a": float64(1), "b": float64(2)},
		},
		{
			name: "multiline block comment",
			src: `{
  /* line1
     line2 */
  "a": 1
}`,
			want: map[string]any{"a": float64(1)},
		},
		{
			name: "url inside string untouched",
			src:  `{"a":"http://x"}`,
			want: map[string]any{"a": "http://x"},
		},
		{
			name: "block-comment-like text inside string untouched",
			src:  `{"a":"/* not a comment */"}`,
			want: map[string]any{"a": "/* not a comment */"},
		},
		{
			name: "line-comment-like text inside string untouched",
			src:  `{"a":"http://example.com // still string"}`,
			want: map[string]any{"a": "http://example.com // still string"},
		},
		{
			name: "escaped quotes in strings",
			src:  `{"a":"say \"hi\" // not a comment"}`,
			want: map[string]any{"a": `say "hi" // not a comment`},
		},
		{
			name: "escaped backslashes in strings",
			src:  `{"a":"C:\\path\\file // still string"}`,
			want: map[string]any{"a": `C:\path\file // still string`},
		},
		{
			name: "trailing comma in object",
			src: `{
  "a": 1,
  "b": 2,
}`,
			want: map[string]any{"a": float64(1), "b": float64(2)},
		},
		{
			name: "trailing comma in array",
			src: `{
  "items": [1, 2, 3,]
}`,
			want: map[string]any{"items": []any{float64(1), float64(2), float64(3)}},
		},
		{
			name: "nested trailing commas",
			src: `{
  "obj": {"x": 1,},
  "arr": [true,],
}`,
			want: map[string]any{
				"obj": map[string]any{"x": float64(1)},
				"arr": []any{true},
			},
		},
		{
			name: "plain JSON unchanged semantically",
			src:  `{"a":1,"b":[true,null],"c":{"d":"e"}}`,
			want: map[string]any{
				"a": float64(1),
				"b": []any{true, nil},
				"c": map[string]any{"d": "e"},
			},
		},
		{
			name: "comments and trailing commas together",
			src: `{
  // header
  "a": 1, /* keep */
  "b": [2, 3,], // end
}`,
			want: map[string]any{
				"a": float64(1),
				"b": []any{float64(2), float64(3)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := StripJSONC([]byte(tt.src))

			var got any
			if err := json.Unmarshal(out, &got); err != nil {
				t.Fatalf("json.Unmarshal after StripJSONC: %v\noutput: %q", err, out)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("value mismatch\ngot:  %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}

func TestStripJSONC_PreservesLineCount(t *testing.T) {
	src := []byte(`{
  // line comment
  "a": 1, /* block
     spanning
     lines */
  "b": 2,
}`)
	out := StripJSONC(src)

	srcLines := strings.Count(string(src), "\n")
	outLines := strings.Count(string(out), "\n")
	if srcLines != outLines {
		t.Errorf("line count changed: src=%d out=%d", srcLines, outLines)
	}
	if len(out) != len(src) {
		t.Errorf("byte length changed: src=%d out=%d (offsets would shift)", len(src), len(out))
	}

	if !bytes.Contains(out, []byte("\n")) {
		t.Fatal("expected newlines to remain after stripping")
	}

	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	want := map[string]any{"a": float64(1), "b": float64(2)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestStripJSONC_PlainJSONByteStableAsideFromTrailingCommas(t *testing.T) {
	src := []byte(`{"ok":true,"n":42}`)
	out := StripJSONC(src)

	var a, b any
	if err := json.Unmarshal(src, &a); err != nil {
		t.Fatalf("unmarshal src: %v", err)
	}
	if err := json.Unmarshal(out, &b); err != nil {
		t.Fatalf("unmarshal out: %v", err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("plain JSON changed semantically:\nsrc=%#v\nout=%#v", a, b)
	}
}
