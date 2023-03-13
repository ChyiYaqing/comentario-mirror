package util

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"abc@def.ru", true},
		{"abc@yktoo.solutions", true},
		{"abc.lastname@example.com", true},
		{"abc.last+new@def.ru", true},
		{"abc.last+new.long@def.ru", true},
		{"abc.last=oi@def.ru", true},
		{"abc@def-.com", true}, // Is it true actually?
		{"abc@de.f.com", true},
		{"abc@def.x", true},
		{"abc@def", true},
		{"abc[@def.com", false},
		{"abc]@def.com", false},
		{"abc @def.com", false},
		{"abc\n@def.com", false},
		{"abc\t@def.com", false},
		{"abc<@def.com", false},
		{"abc>@def.com", false},
		{"abc(@def.com", false},
		{"abc)@def.com", false},
		{"abc[@def.com", false},
		{"abc]@def.com", false},
		{"abc\\@def.com", false},
		{"abc.@def.com", false},
		{"abc,@def.com", false},
		{"abc;@def.com", false},
		{"abc:@def.com", false},
		{"abc@@def.com", false},
		{"abc\"@def.com", false},
		{"abc%@def.com", false},
		{"abc@def..com", false},
		{"abc@de!f.com", false},
		{"abc@de@f.com", false},
		{"abc@de#f.com", false},
		{"abc@de$f.com", false},
		{"abc@de%f.com", false},
		{"abc@de^f.com", false},
		{"abc@de&f.com", false},
		{"abc@de*f.com", false},
		{"abc@de(f.com", false},
		{"abc@de)f.com", false},
		{"abc@de_f.com", false},
		{"abc@de+f.com", false},
		{"abc@de=f.com", false},
		{"abc@de{f.com", false},
		{"abc@de}f.com", false},
		{"abc@de[f.com", false},
		{"abc@de]f.com", false},
		{"abc@de'f.com", false},
		{"abc@de\"f.com", false},
		{"abc@de\\f.com", false},
		{"abc@de|f.com", false},
		{"abc@de:f.com", false},
		{"abc@de;f.com", false},
		{"abc@de<f.com", false},
		{"abc@de>f.com", false},
		{"abc@de,f.com", false},
		{"abc@de/f.com", false},
		{"abc@de?f.com", false},
		{"abc@de~f.com", false},
		{"abc@de`f.com", false},
		{"abc@de f.com", false},
		{"abc@de\nf.com", false},
		{"abc@de\tf.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := IsValidEmail(tt.s); got != tt.want {
				t.Errorf("IsValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidHexID(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"empty string           ", "", false},
		{"string of 63 digits    ", "012345678901234567890123456789012345678901234567890123456789012", false},
		{"string of 64 bad chars ", "012345678901234567890123456789012345678901234567890123456789012g", false},
		{"string of 65 digits    ", "01234567890123456789012345678901234567890123456789012345678901234", false},
		{"string of 64 digits    ", "0123456789012345678901234567890123456789012345678901234567890123", true},
		{"string of 64 hex digits", "1dae2342c9255a4ecc78f2f54380d90508aa49761f3471e94239f178a210bcba", true},
	}
	for _, tt := range tests {
		t.Run(strings.TrimSpace(tt.name), func(t *testing.T) {
			if got := IsValidHexID(tt.s); got != tt.want {
				t.Errorf("IsValidHexID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidHostname(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"empty string                     ", "", false},
		{"single .                         ", ".", false},
		{"single -                         ", "-", false},
		{"single part starting with -      ", "-example", false},
		{"single part containing _         ", "ex_ample", false},
		{"single part, alpha               ", "example", true},
		{"single part, alphanumeric 1      ", "examp1e2", true},
		{"single part, alphanumeric 2      ", "2examp1e", true},
		{"single part 63 chars long        ", "4785tchn2w4g890hn-4t598-u2hxm08-u24htg0m82ug028u5gjkndsrfigjnsf", true},
		{"single part too long             ", "4785tchn2w4g890hn-4t598-u2hxm08-u24htg0m82ug028u5gjikndsrfigjnsf", false},
		{"two dots in a row                ", "e..a", false},
		{"two parts 1                      ", "e.ax", true},
		{"two parts 2                      ", "ex.ample", true},
		{"two parts, second starting with -", "ex.-ample", false},
		{"two parts, one 63 chars long     ", "ex.mplehasdifjhakdhfakjhdfgkajlfhgamplehasdifjhakdhfakjhdfgkajlfhg", true},
		{"two parts, one too long          ", "ex.amplehasdifjhakdhfakjhdfgkajlfhgamplehasdifjhakdhfakjhdfgkajlfhg", false},
		{"many parts                       ", "ex.a.m.p.l.e.h.a.s.d.i.f.j.h.a.k.d.h.f.a.k.j.h.d.fgk.ajl.fhgam.pl.eh.a.sdi.fjh.akdh.fa.kj.h.dfgkajlfhg.nl", true},
		{"many parts, length 253 chars     ", "a.very.very.very.loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong.loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong.loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong.dooooooooooooooooooooooooooooooomain.name.nl", true},
		{"many parts, length 254 chars     ", "a.very.very.very.loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong.loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong.loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong.doooooooooooooooooooooooooooooooomain.name.nl", false},
		{"google.com                       ", "google.com", true},
		{"google.com:80                    ", "google.com:80", false},
	}
	for _, tt := range tests {
		t.Run(strings.TrimSpace(tt.name), func(t *testing.T) {
			if got := IsValidHostname(tt.s); got != tt.want {
				t.Errorf("IsValidHostname() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidHostPort(t *testing.T) {
	tests := []struct {
		name      string
		s         string
		wantValid bool
		wantHost  string
		wantPort  string
	}{
		{"empty string                                ", "", false, "", ""},
		{"only port                                   ", ":80", false, "", ""},
		{"no port, single .                           ", ".", false, "", ""},
		{"with port, single .                         ", ".:3128", false, "", ""},
		{"no port, single -                           ", "-", false, "", ""},
		{"with port, single -                         ", "-:3128", false, "", ""},
		{"no port, single part starting with -        ", "-example", false, "", ""},
		{"with port, single part starting with -      ", "-example:3128", false, "", ""},
		{"no port, single part containing _           ", "ex_ample", false, "", ""},
		{"with port, single part containing _         ", "ex_ample:3128", false, "", ""},
		{"no port, single part, alpha                 ", "example", true, "example", ""},
		{"with port, single part, alpha               ", "example:3128", true, "example", "3128"},
		{"no port, single part, alphanumeric 1        ", "examp1e2", true, "examp1e2", ""},
		{"with port, single part, alphanumeric 1      ", "examp1e2:3128", true, "examp1e2", "3128"},
		{"no port, single part, alphanumeric 2        ", "2examp1e", true, "2examp1e", ""},
		{"with port, single part, alphanumeric 2      ", "2examp1e:3128", true, "2examp1e", "3128"},
		{"no port, single part 63 chars long          ", "4785tchn2w4g890hn-4t598-u2hxm08-u24htg0m82ug028u5gjkndsrfigjnsf", true, "4785tchn2w4g890hn-4t598-u2hxm08-u24htg0m82ug028u5gjkndsrfigjnsf", ""},
		{"with port, single part 63 chars long        ", "4785tchn2w4g890hn-4t598-u2hxm08-u24htg0m82ug028u5gjkndsrfigjnsf:3128", true, "4785tchn2w4g890hn-4t598-u2hxm08-u24htg0m82ug028u5gjkndsrfigjnsf", "3128"},
		{"no port, single part too long               ", "4785tchn2w4g890hn-4t598-u2hxm08-u24htg0m82ug028u5gjikndsrfigjnsf", false, "", ""},
		{"with port, single part too long             ", "4785tchn2w4g890hn-4t598-u2hxm08-u24htg0m82ug028u5gjikndsrfigjnsf:3128", false, "", ""},
		{"no port, two parts 1                        ", "e.ax", true, "e.ax", ""},
		{"with port, two parts 1                      ", "e.ax:3128", true, "e.ax", "3128"},
		{"with empty port, two parts 1                ", "e.ax:", false, "", ""},
		{"no port, two parts 2                        ", "ex.ample", true, "ex.ample", ""},
		{"with port, two parts 2                      ", "ex.ample:3128", true, "ex.ample", "3128"},
		{"with empty port, two parts 2                ", "ex.ample:", false, "", ""},
		{"with zero port, two parts 2                 ", "ex.ample:0", false, "", ""},
		{"with big port, two parts 2                  ", "ex.ample:65536", false, "", ""},
		{"no port, two parts, second starting with -  ", "ex.-ample", false, "", ""},
		{"with port, two parts, second starting with -", "ex.-ample:3128", false, "", ""},
		{"no port, two parts, one 63 chars long       ", "ex.mplehasdifjhakdhfakjhdfgkajlfhgamplehasdifjhakdhfakjhdfgkajlfhg", true, "ex.mplehasdifjhakdhfakjhdfgkajlfhgamplehasdifjhakdhfakjhdfgkajlfhg", ""},
		{"with port, two parts, one 63 chars long     ", "ex.mplehasdifjhakdhfakjhdfgkajlfhgamplehasdifjhakdhfakjhdfgkajlfhg:3128", true, "ex.mplehasdifjhakdhfakjhdfgkajlfhgamplehasdifjhakdhfakjhdfgkajlfhg", "3128"},
		{"no port, two parts, one too long            ", "ex.amplehasdifjhakdhfakjhdfgkajlfhgamplehasdifjhakdhfakjhdfgkajlfhg", false, "", ""},
		{"with port, two parts, one too long          ", "ex.amplehasdifjhakdhfakjhdfgkajlfhgamplehasdifjhakdhfakjhdfgkajlfhg:3128", false, "", ""},
		{"no port, many parts                         ", "ex.a.m.p.l.e.h.a.s.d.i.f.j.h.a.k.d.h.f.a.k.j.h.d.fgk.ajl.fhgam.pl.eh.a.sdi.fjh.akdh.fa.kj.h.dfgkajlfhg.nl", true, "ex.a.m.p.l.e.h.a.s.d.i.f.j.h.a.k.d.h.f.a.k.j.h.d.fgk.ajl.fhgam.pl.eh.a.sdi.fjh.akdh.fa.kj.h.dfgkajlfhg.nl", ""},
		{"with port, many parts                       ", "ex.a.m.p.l.e.h.a.s.d.i.f.j.h.a.k.d.h.f.a.k.j.h.d.fgk.ajl.fhgam.pl.eh.a.sdi.fjh.akdh.fa.kj.h.dfgkajlfhg.nl:3128", true, "ex.a.m.p.l.e.h.a.s.d.i.f.j.h.a.k.d.h.f.a.k.j.h.d.fgk.ajl.fhgam.pl.eh.a.sdi.fjh.akdh.fa.kj.h.dfgkajlfhg.nl", "3128"},
		{"no port, many parts                         ", "ex.a.m.p.l.e.h.a.s.d.i.f.j.h.a.k.d.h.f.a.k.j.h.d.fgk.ajl.fhgam.pl.eh.a.sdi.fjh.akdh.fa.kj.h.dfgkajlfhg.nl", true, "ex.a.m.p.l.e.h.a.s.d.i.f.j.h.a.k.d.h.f.a.k.j.h.d.fgk.ajl.fhgam.pl.eh.a.sdi.fjh.akdh.fa.kj.h.dfgkajlfhg.nl", ""},
		{"with port, many parts                       ", "ex.a.m.p.l.e.h.a.s.d.i.f.j.h.a.k.d.h.f.a.k.j.h.d.fgk.ajl.fhgam.pl.eh.a.sdi.fjh.akdh.fa.kj.h.dfgkajlfhg.nl:3128", true, "ex.a.m.p.l.e.h.a.s.d.i.f.j.h.a.k.d.h.f.a.k.j.h.d.fgk.ajl.fhgam.pl.eh.a.sdi.fjh.akdh.fa.kj.h.dfgkajlfhg.nl", "3128"},
		{"no port, comentario.app                     ", "comentario.app", true, "comentario.app", ""},
		{"with port, comentario.app                   ", "comentario.app:3128", true, "comentario.app", "3128"},
	}
	for _, tt := range tests {
		t.Run(strings.TrimSpace(tt.name), func(t *testing.T) {
			if gotValid, gotHost, gotPort := IsValidHostPort(tt.s); gotValid != tt.wantValid || gotHost != tt.wantHost || gotPort != tt.wantPort {
				t.Errorf("IsValidHostPort() = (%v, %v, %v), want (%v, %v, %v)", gotValid, gotHost, gotPort, tt.wantValid, tt.wantHost, tt.wantPort)
			}
		})
	}
}

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"empty string   ", "", false},
		{"alpha          ", "cc", false},
		{"alphanumeric 1 ", "a12", false},
		{"alphanumeric 2 ", "8f", false},
		{"zero           ", "0", false},
		{"too big number ", "65536", false},
		{"small number OK", "1", true},
		{"big number OK  ", "65535", true},
	}
	for _, tt := range tests {
		t.Run(strings.TrimSpace(tt.name), func(t *testing.T) {
			if got := IsValidPort(tt.str); got != tt.want {
				t.Errorf("IsValidPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUILang(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want bool
	}{
		{"empty string", "", false},
		{"1 char", "e", false},
		{"3 chars", "ene", false},
		{"af is not", "af", false},
		{"am is not", "am", false},
		{"ar is not", "ar", false},
		{"az is not", "az", false},
		{"bg is not", "bg", false},
		{"bn is not", "bn", false},
		{"ca is not", "ca", false},
		{"cs is not", "cs", false},
		{"da is not", "da", false},
		{"de is not", "de", false},
		{"el is not", "el", false},
		{"en is supported", "en", true},
		{"es is not", "es", false},
		{"et is not", "et", false},
		{"fa is not", "fa", false},
		{"fi is not", "fi", false},
		{"fr is not", "fr", false},
		{"gu is not", "gu", false},
		{"he is not", "he", false},
		{"hi is not", "hi", false},
		{"hr is not", "hr", false},
		{"hu is not", "hu", false},
		{"hy is not", "hy", false},
		{"id is not", "id", false},
		{"is is not", "is", false},
		{"it is not", "it", false},
		{"ja is not", "ja", false},
		{"ka is not", "ka", false},
		{"kk is not", "kk", false},
		{"km is not", "km", false},
		{"kn is not", "kn", false},
		{"ko is not", "ko", false},
		{"ky is not", "ky", false},
		{"lo is not", "lo", false},
		{"lt is not", "lt", false},
		{"lv is not", "lv", false},
		{"mk is not", "mk", false},
		{"ml is not", "ml", false},
		{"mn is not", "mn", false},
		{"mr is not", "mr", false},
		{"ms is not", "ms", false},
		{"my is not", "my", false},
		{"ne is not", "ne", false},
		{"nl is not", "nl", false},
		{"no is not", "no", false},
		{"pa is not", "pa", false},
		{"pl is not", "pl", false},
		{"pt is not", "pt", false},
		{"ro is not", "ro", false},
		{"ru is not", "ru", false},
		{"si is not", "si", false},
		{"sk is not", "sk", false},
		{"sl is not", "sl", false},
		{"sq is not", "sq", false},
		{"sr is not", "sr", false},
		{"sv is not", "sv", false},
		{"sw is not", "sw", false},
		{"ta is not", "ta", false},
		{"te is not", "te", false},
		{"th is not", "th", false},
		{"tr is not", "tr", false},
		{"uk is not", "uk", false},
		{"ur is not", "ur", false},
		{"uz is not", "uz", false},
		{"vi is not", "vi", false},
		{"zh is not", "zh", false},
		{"zu is not", "zu", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUILang(tt.lang); got != tt.want {
				t.Errorf("IsUILang() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarkdownToHTML(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{"Bare text   ", "Foo", "<p>Foo</p>"},
		{"Paragraphs  ", "Foo\n\nBar", "<p>Foo</p>\n\n<p>Bar</p>"},
		{"Script      ", "XSS: <script src='http://example.com/script.js'></script> Foo", "<p>XSS:  Foo</p>"},
		{"Regular link", "Regular [Link](http://example.com)", "<p>Regular <a href=\"http://example.com\" rel=\"nofollow noopener\" target=\"_blank\">Link</a></p>"},
		{"XSS link    ", "XSS [Link](data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pgo=)", "<p>XSS <tt>Link</tt></p>"},
		{"Image       ", "![Images disallowed](http://example.com/image.jpg)", "<p></p>"},
		{"Formatting  ", "**bold** *italics*", "<p><strong>bold</strong> <em>italics</em></p>"},
		{"URL         ", "http://example.com/autolink", "<p><a href=\"http://example.com/autolink\" rel=\"nofollow noopener\" target=\"_blank\">http://example.com/autolink</a></p>"},
		{"HTML        ", "<b>not bold</b>", "<p>not bold</p>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Trim leading/trailing whitespace explicitly before comparing (because it doesn't matter in the resulting
			// HTML)
			if got := strings.TrimSpace(MarkdownToHTML(tt.markdown)); got != tt.want {
				t.Errorf("MarkdownToHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeStringMap(t *testing.T) {
	m := SafeStringMap{}
	var wg sync.WaitGroup

	// Run 500 parallel threads
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// Insert 500 values into the map
			for j := 0; j < 500; j++ {
				m.Put(fmt.Sprintf("v-%d-%d", idx, j), "xyz")
			}
			// Now take all those values, in reverse order
			for j := 499; j >= 0; j-- {
				if got, ok := m.Take(fmt.Sprintf("v-%d-%d", idx, j)); !ok {
					t.Errorf("SafeStringMap.Take() misses a value for i = %d, j = %d", idx, j)
				} else if got != "xyz" {
					t.Errorf("SafeStringMap.Take() returned %v for i = %d, j = %d (want \"xyz\")", got, idx, j)
				}
			}
		}(i)
	}

	// Wait for the crunch to finish
	wg.Wait()

	// Verify the map is empty
	if got := m.Len(); got != 0 {
		t.Errorf("SafeStringMap.Len() returned %d, want 0", got)
	}
}
