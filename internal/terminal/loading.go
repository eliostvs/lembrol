package terminal

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
)

func loadingView(title string, spin spinner.Model) string {
	content := titleStyle.Render(title)
	content += normalTextStyle.Render(fmt.Sprintf("%s Loading...", spin.View()))
	return largePaddingStyle.Render(content)
}
