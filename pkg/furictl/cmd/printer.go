/*
 * Copyright 2022 The Furiko Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"io"
	"math"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/xeonx/timeago"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	shorthandFormatter = timeago.Config{
		FuturePrefix: "in ",
		Periods: []timeago.FormatPeriod{
			{time.Second, "1s", "%ds"},
			{time.Minute, "1m", "%dm"},
			{time.Hour, "1d", "%dh"},
			{timeago.Day, "1d", "%dd"},
		},
		Zero:          "0s",
		Max:           math.MaxInt64, // no max
		DefaultLayout: time.RFC3339,
	}
)

// PrintTable prints the given header and rows to standard output.
func PrintTable(header []string, rows [][]string) {
	FprintTable(os.Stdout, header, rows)
}

// FprintTable prints the given header and rows to the given output.
func FprintTable(w io.Writer, header []string, rows [][]string) {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	style := table.StyleLight
	style.Options = table.Options{
		SeparateColumns: true,
	}
	style.Box.PaddingLeft = ""
	style.Box.PaddingRight = ""
	style.Box.MiddleVertical = "   "
	t.SetStyle(style)
	t.AppendHeader(makeTableRow(header))
	t.AppendRows(makeTableRows(rows))
	t.Render()
}

func makeTableRows(rows [][]string) []table.Row {
	output := make([]table.Row, 0, len(rows))
	for _, row := range rows {
		output = append(output, makeTableRow(row))
	}
	return output
}

func makeTableRow(cells []string) table.Row {
	output := make(table.Row, 0, len(cells))
	for _, cell := range cells {
		output = append(output, cell)
	}
	return output
}

// FormatTimeAgo formats a time as a string representing the duration since the
// given time.
func FormatTimeAgo(t *metav1.Time) string {
	if t.IsZero() {
		return ""
	}
	return shorthandFormatter.Format(t.Time)
}
