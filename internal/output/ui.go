package output

import (
	"fmt"
	"os"
)

func Header(out *os.File, styles Styles, title string, detail ...any) {
	if len(detail) == 0 {
		WriteLine(out, fmt.Sprintf("%s  %s", styles.Accent.Render("◆"), styles.Title.Render(title)))
		return
	}
	WriteLine(
		out,
		fmt.Sprintf(
			"%s  %s %s",
			styles.Accent.Render("◆"),
			styles.Title.Render(title),
			styles.Warm.Render(fmt.Sprint(detail...)),
		),
	)
}

func Intro(out *os.File, styles Styles, title string) {
	WriteLine(out, fmt.Sprintf("%s  %s", styles.Frame.Render("┌"), styles.Title.Render(title)))
	WriteLine(out, styles.Frame.Render("│"))
}

func Outro(out *os.File, styles Styles, message string) {
	WriteLine(out, fmt.Sprintf("%s  %s", styles.Frame.Render("└"), styles.Value.Render(message)))
}

func Divider(out *os.File, styles Styles) {
	WriteLine(out, styles.Frame.Render("│"))
}

func Info(out *os.File, styles Styles, message string) {
	WriteLine(out, fmt.Sprintf("%s  %s", styles.Secondary.Render("●"), styles.Value.Render(message)))
}

func Done(out *os.File, styles Styles, message string) {
	WriteLine(out, fmt.Sprintf("%s  %s", styles.Success.Render("◇"), styles.Value.Render(message)))
}

func Warn(out *os.File, styles Styles, message string) {
	WriteLine(out, fmt.Sprintf("%s  %s", styles.Warning.Render("▲"), styles.Value.Render(message)))
}

func Fail(out *os.File, styles Styles, message string) {
	WriteLine(out, fmt.Sprintf("%s  %s", styles.Error.Render("■"), styles.Value.Render(message)))
}
