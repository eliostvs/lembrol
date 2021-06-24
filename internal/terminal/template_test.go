package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_helpTag(t *testing.T) {
	type args struct {
		width    int
		sections []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "short line short width",
			args: args{
				width:    30,
				sections: []string{"esc: back", "q: quit"},
			},
			want: "   esc: back • q: quit",
		},
		{
			name: "long line large width",
			args: args{
				width:    120,
				sections: []string{"0: again", "1: very hard", "2: hard", "3: normal", "4: easy", "5: super easy", "esc: close", "q: quit"},
			},
			want: "   0: again • 1: very hard • 2: hard • 3: normal • 4: easy • 5: super easy • esc: close • q: quit",
		},
		{
			name: "long line few medium width",
			args: args{
				width:    35,
				sections: []string{"0: again", "1: very hard", "2: hard", "3: normal", "4: easy", "5: super easy", "esc: close", "q: quit"},
			},
			want: "   0: again • 1: very hard • 2: hard\n   3: normal • 4: easy • 5: super easy\n   esc: close • q: quit",
		},
		{
			name: "long line short width",
			args: args{
				width:    15,
				sections: []string{"0: again", "1: very hard", "2: hard", "3: normal", "4: easy", "5: super easy", "esc: close", "q: quit"},
			},
			want: "   0: again\n   1: very hard\n   2: hard\n   3: normal\n   4: easy\n   5: super easy\n   esc: close\n   q: quit",
		},
		{
			name: "empty line",
			args: args{
				width:    30,
				sections: []string{},
			},
			want: "",
		},
		{
			name: "many empty sections",
			args: args{
				width:    50,
				sections: []string{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},
			},
			want: "   ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helpTag(tt.args.width, tt.args.sections...)

			assert.Equal(t, tt.want, got)
		})
	}
}
