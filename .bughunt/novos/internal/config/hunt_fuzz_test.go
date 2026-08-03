package config

import (
	"bytes"
	"encoding/json"
	"testing"
)

// FuzzStripJSONCInvariants checks the properties StripJSONC's own doc comment
// promises, plus idempotence.
func FuzzStripJSONCInvariants(f *testing.F) {
	seeds := []string{
		`{}`,
		`{"a":1}`,
		`{"a":1,}`,
		`{"a":[1,2,]}`,
		`{ // c
"a":1}`,
		`{/* c */"a":1}`,
		`{"a":"//not a comment"}`,
		`{"a":"/*not a comment*/"}`,
		`{"a":"\\"}`,
		`{"a":"}"}`,
		`{"a":"]"}`,
		`[1,2,]`,
		`{"a":1} // tail`,
		`/**/{}`,
		`/*`,
		`//`,
		`{"a":1,/*x*/}`,
		"{\"a\":1,\n// c\n}",
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, src []byte) {
		orig := append([]byte(nil), src...)
		out := StripJSONC(src)

		// P1: byte offsets stay accurate -> length is preserved.
		if len(out) != len(orig) {
			t.Fatalf("length changed: in=%d out=%d\nin=%q\nout=%q", len(orig), len(out), orig, out)
		}

		// P2: newlines are preserved so line numbers stay correct.
		if got, want := bytes.Count(out, []byte{'\n'}), bytes.Count(orig, []byte{'\n'}); got != want {
			t.Fatalf("newline count changed: in=%d out=%d\nin=%q\nout=%q", want, got, orig, out)
		}

		// P3: idempotence over the meaningful domain — inputs that actually
		// strip to a JSON document. (Garbage like ",,}" is invalid before and
		// after; a second pass blanking another stray comma harms nobody.)
		if json.Valid(out) {
			again := StripJSONC(append([]byte(nil), out...))
			if !bytes.Equal(again, out) {
				t.Fatalf("not idempotent:\nin=%q\n1st=%q\n2nd=%q", orig, out, again)
			}
		}

		// P4: valid JSON contains no comments and no trailing commas, so it
		// must survive untouched.
		if json.Valid(orig) && !bytes.Equal(out, orig) {
			t.Fatalf("valid JSON was modified:\nin=%q\nout=%q", orig, out)
		}
	})
}

// FuzzJSONCCommentInjection is the metamorphic oracle for StripJSONC: take any
// valid JSON document, inject JSONC comments and trailing commas into
// whitespace positions, and require the stripped result to decode to the exact
// same value. This is what "JSONC support" means in practice.
func FuzzJSONCCommentInjection(f *testing.F) {
	seeds := []string{
		`{"a":1}`,
		`{"a":[1,2],"b":{"c":null}}`,
		`{"a":"//x","b":"/*y*/"}`,
		`[]`,
		`{}`,
		`{"a":{"b":{"c":[1,{"d":2}]}}}`,
		`{"a":"line\nbreak"}`,
	}
	for _, s := range seeds {
		f.Add([]byte(s), uint8(0))
	}

	f.Fuzz(func(t *testing.T, src []byte, mode uint8) {
		var want any
		if err := json.Unmarshal(src, &want); err != nil {
			return
		}

		var pretty bytes.Buffer
		if err := json.Indent(&pretty, src, "", "  "); err != nil {
			return
		}

		injected := injectJSONC(pretty.Bytes(), mode)
		stripped := StripJSONC(append([]byte(nil), injected...))

		var got any
		if err := json.Unmarshal(stripped, &got); err != nil {
			t.Fatalf("injected JSONC does not strip to valid JSON: %v\ninjected=%q\nstripped=%q", err, injected, stripped)
		}
		wantJSON, _ := json.Marshal(want)
		gotJSON, _ := json.Marshal(got)
		if !bytes.Equal(wantJSON, gotJSON) {
			t.Fatalf("value changed by comment injection:\ninjected=%q\nwant=%s\ngot=%s", injected, wantJSON, gotJSON)
		}
	})
}

// injectJSONC rewrites indented JSON into equivalent JSONC: line comments after
// a line's content, block comments before the indentation, and a trailing comma
// before every closing brace/bracket. mode selects which of the three are used.
func injectJSONC(pretty []byte, mode uint8) []byte {
	lineComments := mode&1 != 0
	blockComments := mode&2 != 0
	trailingCommas := mode&4 != 0
	if mode == 0 {
		lineComments, blockComments, trailingCommas = true, true, true
	}

	lines := bytes.Split(pretty, []byte{'\n'})
	var out bytes.Buffer
	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		closing := len(trimmed) > 0 && (trimmed[0] == '}' || trimmed[0] == ']')

		if trailingCommas && closing && i > 0 {
			prev := bytes.TrimRight(out.Bytes(), "\n")
			if len(prev) > 0 && prev[len(prev)-1] != '{' && prev[len(prev)-1] != '[' && prev[len(prev)-1] != ',' {
				out.Truncate(len(prev))
				out.WriteString(",\n")
			}
		}
		if blockComments {
			out.WriteString("/* b */")
		}
		out.Write(line)
		if lineComments && i < len(lines)-1 {
			out.WriteString(" // c")
		}
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	return out.Bytes()
}

// FuzzDocumentRoundTrip checks that a parsable document serializes to a stable
// canonical form: Bytes() must reparse, and the second serialization must be
// byte-identical to the first.
func FuzzDocumentRoundTrip(f *testing.F) {
	seeds := []string{
		`{}`,
		`{"$schema":"x"}`,
		`{"profiles":{"a":{"[opencode]":{}}}}`,
		`{"profiles":{"a":{"[opencode]":{"agents":{"x":{}}}}}}`,
		`{"[opencode]":{"disabled_mcps":[]}}`,
		`{"profiles":{}}`,
		`{"a":null}`,
		`{"a":1e999}`,
		`{"a":"\ud800"}`,
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, src []byte) {
		doc, err := ParseDocument(src)
		if err != nil {
			return // not a document; nothing promised
		}

		first, err := doc.Bytes()
		if err != nil {
			t.Fatalf("Bytes() failed on a parsed document: %v\nsrc=%q", err, src)
		}

		doc2, err := ParseDocument(first)
		if err != nil {
			t.Fatalf("canonical output does not reparse: %v\nsrc=%q\nout=%q", err, src, first)
		}
		second, err := doc2.Bytes()
		if err != nil {
			t.Fatalf("Bytes() failed on reparsed document: %v\nout=%q", err, first)
		}
		if !bytes.Equal(first, second) {
			t.Fatalf("canonical form unstable:\nsrc=%q\n1st=%q\n2nd=%q", src, first, second)
		}
	})
}
