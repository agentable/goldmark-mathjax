package mathjax

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/yuin/goldmark"

	"github.com/stretchr/testify/assert"
)

type mathJaxTestCase struct {
	d   string // test description
	in  string // input markdown source
	out string // expected output html
}

func TestMathJax(t *testing.T) {

	tests := []mathJaxTestCase{
		{
			d:   "plain text",
			in:  "foo",
			out: `<p>foo</p>`,
		},
		{
			d:   "bold",
			in:  "**foo**",
			out: `<p><strong>foo</strong></p>`,
		},
		{
			d:   "math inline",
			in:  "$1+2$",
			out: `<p><span class="math inline">\(1+2\)</span></p>`,
		},
		{
			d:  "math display",
			in: "$$\n1+2\n$$",
			out: `<p><span class="math display">\[1+2
\]</span></p>`,
		},
		{
			// this input previously triggered a panic in block.go
			d:   "list-begin",
			in:  "*foo\n  ",
			out: "<p>*foo</p>",
		},
		// Same-line format tests
		{
			d:   "math display - same line simple",
			in:  `$$x+y$$`,
			out: `<p><span class="math display">\[x+y\]</span></p>`,
		},
		{
			d:   "math display - same line complex",
			in:  `$$\iint^{\infty}_{0}{xdxdy}$$`,
			out: `<p><span class="math display">\[\iint^{\infty}_{0}{xdxdy}\]</span></p>`,
		},
		{
			d:   "math display - same line with spaces",
			in:  `$$  a + b  $$`,
			out: `<p><span class="math display">\[  a + b  \]</span></p>`,
		},
		{
			d:   "math display - same line empty",
			in:  `$$$$`,
			out: `<p><span class="math display">\[\]</span></p>`,
		},
		// Consecutive blocks tests
		{
			d:  "math display - two same-line blocks",
			in: "$$x+y$$\n$$a+b$$",
			out: `<p><span class="math display">\[x+y\]</span></p>
<p><span class="math display">\[a+b\]</span></p>`,
		},
		{
			d:  "math display - two same-line blocks with blank line",
			in: "$$x+y$$\n\n$$a+b$$",
			out: `<p><span class="math display">\[x+y\]</span></p>
<p><span class="math display">\[a+b\]</span></p>`,
		},
		{
			d:  "math display - consecutive multi-line blocks with blank line",
			in: "$$\n1+2\n$$\n\n$$\n3+4\n$$",
			out: `<p><span class="math display">\[1+2
\]</span></p>
<p><span class="math display">\[3+4
\]</span></p>`,
		},
		// Mixed format tests
		{
			d:  "math display - multi-line then same-line",
			in: "$$\nx+y\n$$\n$$a+b$$",
			out: `<p><span class="math display">\[x+y
\]</span></p>
<p><span class="math display">\[a+b\]</span></p>`,
		},
		// Multi-line with content variations
		{
			d:  "math display - multi-line multiple lines content",
			in: "$$\na+b\\\\\nc+d\n$$",
			out: `<p><span class="math display">\[a+b\\
c+d
\]</span></p>`,
		},
		// Text mixing tests
		{
			d:  "math display - text before same-line",
			in: "text before\n$$x+y$$",
			out: `<p>text before</p>
<p><span class="math display">\[x+y\]</span></p>`,
		},
		{
			d:  "math display - text after same-line",
			in: "$$x+y$$\ntext after",
			out: `<p><span class="math display">\[x+y\]</span></p>
<p>text after</p>`,
		},
		{
			d:  "math display - text before and after same-line",
			in: "before\n$$x+y$$\nafter",
			out: `<p>before</p>
<p><span class="math display">\[x+y\]</span></p>
<p>after</p>`,
		},
		// vmatrix test - bug report case
		{
			d: "math display - vmatrix multiline",
			in: `Before matrix

$$\begin{vmatrix}
\vec{i} & \vec{j} & \vec{k} \\
1 & 2 & 3 \\
4 & 5 & 6
\end{vmatrix}$$

After matrix`,
			out: `<p>Before matrix</p>
<p><span class="math display">\[\begin{vmatrix}
\vec{i} & \vec{j} & \vec{k} \\
1 & 2 & 3 \\
4 & 5 & 6
\end{vmatrix}\]</span></p>
<p>After matrix</p>`,
		},
		// pmatrix test
		{
			d: "math display - pmatrix multiline",
			in: `Before matrix

$$\begin{pmatrix}
1 & 2 \\
3 & 4
\end{pmatrix}$$

After matrix`,
			out: `<p>Before matrix</p>
<p><span class="math display">\[\begin{pmatrix}
1 & 2 \\
3 & 4
\end{pmatrix}\]</span></p>
<p>After matrix</p>`,
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tc.d), func(t *testing.T) {
			out, err := renderMarkdown([]byte(tc.in))
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.out, strings.TrimSpace(string(out)))
		})
	}

}

func renderMarkdown(src []byte) ([]byte, error) {
	md := goldmark.New(
		goldmark.WithExtensions(MathJax),
	)

	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
