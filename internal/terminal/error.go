package terminal

func errorView(err string) string {
	content := titleStyle.Render("Error")
	content += Red.Render(err)
	return largePaddingStyle.Render(content)
}
