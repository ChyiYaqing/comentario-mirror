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
