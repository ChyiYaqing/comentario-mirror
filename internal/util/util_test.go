package util

import (
	"strings"
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
		{"abc@def.x", false},
		{"abc@def", false},
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
		{"single part, alpha               ", "example", false},
		{"single part, alphanumeric 1      ", "examp1e2", false},
		{"single part, alphanumeric 2      ", "2examp1e", false},
		{"single part 63 chars long        ", "4785tchn2w4g890hn-4t598-u2hxm08-u24htg0m82ug028u5gjkndsrfigjnsf", false},
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
		{"no port, single part, alpha                 ", "example", false, "", ""},
		{"with port, single part, alpha               ", "example:3128", false, "", ""},
		{"no port, single part, alphanumeric 1        ", "examp1e2", false, "", ""},
		{"with port, single part, alphanumeric 1      ", "examp1e2:3128", false, "", ""},
		{"no port, single part, alphanumeric 2        ", "2examp1e", false, "", ""},
		{"with port, single part, alphanumeric 2      ", "2examp1e:3128", false, "", ""},
		{"no port, single part 63 chars long          ", "4785tchn2w4g890hn-4t598-u2hxm08-u24htg0m82ug028u5gjkndsrfigjnsf", false, "", ""},
		{"with port, single part 63 chars long        ", "4785tchn2w4g890hn-4t598-u2hxm08-u24htg0m82ug028u5gjkndsrfigjnsf:3128", false, "", ""},
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
