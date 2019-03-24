package poster

import (
	"reflect"
	"regexp"
	"testing"
)

func Test_messageSplitter(t *testing.T) {
	type args struct {
		input  string
		msgLen int
		r      *regexp.Regexp
	}

	r := regexp.MustCompile("([.])")

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"basicTest",
			args{
				input:  `1,1 Szczęśliwy mąż, który nie idzie za radą występnych, nie wchodzi na drogę grzeszników i nie siada w kole szyderców, 1,2 lecz ma upodobanie w Prawie Pana, nad Jego Prawem rozmyśla dniem i nocą. 1,3 Jest on jak drzewo zasadzone nad płynącą wodą, które wydaje owoc w swoim czasie, a liście jego nie więdną: co uczyni, pomyślnie wypada. 1,4 Nie tak występni, nie tak: są oni jak plewa, którą wiatr rozmiata. 1,5 Toteż występni nie ostoją się na sądzie ani grzesznicy - w zgromadzeniu sprawiedliwych, 1,6 bo Pan uznaje drogę sprawiedliwych, a droga występnych zaginie.`,
				msgLen: 10,
				r:      r,
			},
			[]string{`1,1 Szczęśliwy mąż, który nie idzie za radą występnych, nie wchodzi na drogę grzeszników i nie siada w kole szyderców, 1,2 lecz ma upodobanie w Prawie Pana, nad Jego Prawem rozmyśla dniem i nocą.`,
				` 1,3 Jest on jak drzewo zasadzone nad płynącą wodą, które wydaje owoc w swoim czasie, a liście jego nie więdną: co uczyni, pomyślnie wypada.`,
				` 1,4 Nie tak występni, nie tak: są oni jak plewa, którą wiatr rozmiata.`,
				` 1,5 Toteż występni nie ostoją się na sądzie ani grzesznicy - w zgromadzeniu sprawiedliwych, 1,6 bo Pan uznaje drogę sprawiedliwych, a droga występnych zaginie.`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := messageSplitter(tt.args.input, tt.args.msgLen, tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("messageSplitter() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
