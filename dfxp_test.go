package caps

import (
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yosssi/gohtml"
)

const sampleDFXP string = `<?xml version="1.0" encoding="utf-8"?>
<tt xml:lang="en" xmlns="http://www.w3.org/ns/ttml" xmlns:tts="http://www.w3.org/ns/ttml#styling">
  <head>
    <styling>
      <style xml:id="p" tts:fontfamily="Arial" tts:fontsize="10pt" tts:textAlign="center" tts:color="#ffeedd"></style>
    </styling>
    <layout>
      <region tts:displayAlign="after" tts:textAlign="center" xml:id="bottom"></region>
    </layout>
  </head>
  <body>
    <div xml:lang="en-US">
      <p begin="00:00:14.848" end="00:00:17.000" style="p">
        MAN:
        <br/>
        When we think
        <br/>
        ♪ ...say bow, wow, ♪
      </p>
     <p begin="00:00:17.000" end="00:00:18.752" style="p">
      <span tts:textalign="right">we have this vision of Einstein</span>
     </p>
     <p begin="00:00:18.752" end="00:00:20.887" style="p">
       <br/>
       as an old, wrinkly man
       <br/>
       with white hair.
     </p>
     <p begin="00:00:20.887" end="00:00:26.760" style="p">
      MAN 2:
      <br/>
      E equals m c-squared is
      <br/>
      not about an old Einstein.
     </p>
     <p begin="00:00:26.760" end="00:00:32.200" style="p">
      MAN 2:
      <br/>
      It's all about an eternal Einstein.  pois é
     </p>
     <p begin="00:00:32.200" end="00:00:36.200" style="p">
      &lt;LAUGHING &amp; WHOOPS!&gt;
     </p>
     <p begin="00:00:34.400" end="00:00:38.400" region="bottom" style="p">
      some more text
     </p>
    </div>
  </body>
</tt>`

const sampleDFXPSyntaxError = `
  <?xml version="1.0" encoding="UTF-8"?>
  <tt xml:lang="en" xmlns="http://www.w3.org/ns/ttml">
  <body>
    <div>
      <p begin="0:00:02.07" end="0:00:05.07">>>THE GENERAL ASSEMBLY'S 2014</p>
      <p begin="0:00:05.07" end="0:00:06.21">SESSION GOT OFF TO A LATE START,</p>
    </div>
   </body>
  </tt>
`

const sampleDFXPEmpty = `
  <?xml version="1.0" encoding="utf-8"?>
  <tt xml:lang="en" xmlns="http://www.w3.org/ns/ttml"
      xmlns:tts="http://www.w3.org/ns/ttml#styling">
   <head>
    <styling>
     <style xml:id="p" tts:color="#ffeedd" tts:fontfamily="Arial"
          tts:fontsize="10pt" tts:textAlign="center"/>
    </styling>
    <layout>
    </layout>
   </head>
   <body>
    <div xml:lang="en-US">
    </div>
   </body>
  </tt>
`

func TestDection(t *testing.T) {
	assert.True(t, NewDFXPReader().Detect(sampleDFXP))
}

func TestCaptionLength(t *testing.T) {
	captionSet, err := NewDFXPReader().Read(sampleDFXP)
	assert.Nil(t, err)
	assert.Equal(t, 7, len(captionSet.GetCaptions("en-US")))
}

func TestEmptyFile(t *testing.T) {
	set, err := NewDFXPReader().Read(sampleDFXPEmpty)
	assert.NotNil(t, err)
	assert.True(t, set.IsEmpty())
}

func TestProperTimestamps(t *testing.T) {
	captionSet, err := NewDFXPReader().Read(sampleDFXP)
	assert.Nil(t, err)

	paragraph := captionSet.GetCaptions("en-US")[2]
	assert.Equal(t, 18752000, paragraph.Start)
	assert.Equal(t, 20887000, paragraph.End)
}

func TestInvalidMarkupIsProperlyHandled(t *testing.T) {
	captionSet, err := NewDFXPReader().Read(sampleDFXPSyntaxError)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(captionSet.GetCaptions("en-US")))
}

func TestCaptionNodes(t *testing.T) {
	captionSet, err := NewDFXPReader().Read(sampleDFXP)
	assert.Nil(t, err)
	styles := captionSet.GetStyles()
	assert.Equal(t, 1, len(styles))
	assert.Equal(t, styles[0].ID, "p")
	assert.Equal(t, styles[0].FontFamily, "Arial")
	assert.Equal(t, styles[0].FontFamily, "Arial")
	assert.Equal(t, styles[0].FontSize, "10pt")
	assert.Equal(t, styles[0].TextAlign, "center")
	assert.Equal(t, styles[0].Color, "#ffeedd")
	type captionNodeTest struct {
		wantFormatStart string
		wantFormatEnd   string
		wantText        string
	}
	captionTests := []captionNodeTest{
		{wantFormatStart: "00:00:14.848", wantFormatEnd: "00:00:17.000", wantText: "MAN:\nWhen we think\n♪ ...say bow, wow, ♪"},
		{wantFormatStart: "00:00:17.000", wantFormatEnd: "00:00:18.752", wantText: "we have this vision of Einstein"},
		{wantFormatStart: "00:00:18.752", wantFormatEnd: "00:00:20.887", wantText: "\nas an old, wrinkly man\nwith white hair."},
		{wantFormatStart: "00:00:20.887", wantFormatEnd: "00:00:26.760", wantText: "MAN 2:\nE equals m c-squared is\nnot about an old Einstein."},
		{wantFormatStart: "00:00:26.760", wantFormatEnd: "00:00:32.200", wantText: "MAN 2:\nIt's all about an eternal Einstein.  pois é"},
		{wantFormatStart: "00:00:32.200", wantFormatEnd: "00:00:36.200", wantText: "<LAUGHING & WHOOPS!>"},
		{wantFormatStart: "00:00:34.400", wantFormatEnd: "00:00:38.400", wantText: "some more text"},
	}

	type nodeTest struct {
		kind    Kind
		content string
	}

	nodeTests := [][]nodeTest{
		[]nodeTest{
			{kind: Text, content: "MAN:"},
			{kind: LineBreak, content: "\n"},
			{kind: Text, content: "When we think"},
			{kind: LineBreak, content: "\n"},
			{kind: Text, content: "♪ ...say bow, wow, ♪"},
		},
		[]nodeTest{
			{kind: CapStyle, content: `
	class: \n
	text-align: right\n
	font-family: \n
	font-size: \n
	color: \n
	italics: false\n
	bold: false\n
	underline: false\n
	`,
			},
			{kind: Text, content: "we have this vision of Einstein"},
		},
		[]nodeTest{
			{kind: LineBreak, content: "\n"},
			{kind: Text, content: "as an old, wrinkly man"},
			{kind: LineBreak, content: "\n"},
			{kind: Text, content: "with white hair."},
		},
		[]nodeTest{
			{kind: Text, content: "MAN 2:"},
			{kind: LineBreak, content: "\n"},
			{kind: Text, content: "E equals m c-squared is"},
			{kind: LineBreak, content: "\n"},
			{kind: Text, content: "not about an old Einstein."},
		},
		[]nodeTest{
			{kind: Text, content: "MAN 2:"},
			{kind: LineBreak, content: "\n"},
			{kind: Text, content: "It's all about an eternal Einstein.  pois é"},
		},
		[]nodeTest{
			{kind: Text, content: "<LAUGHING & WHOOPS!>"},
		},
		[]nodeTest{
			{kind: Text, content: "some more text"},
		},
	}

	captions := captionSet.GetCaptions("en-US")
	for i, caption := range captions {
		assert.Equal(t, caption.FormatStart(), captionTests[i].wantFormatStart)
		assert.Equal(t, caption.FormatEnd(), captionTests[i].wantFormatEnd)
		assert.Equal(t, caption.GetText(), captionTests[i].wantText)
		for j, node := range caption.Nodes {
			assert.Equal(t, node.Kind(), nodeTests[i][j].kind)
			assert.Equal(t, node.GetContent(), nodeTests[i][j].content)
		}
	}
}

func TestStructToXML(t *testing.T) {
	base := dfxpSpan{xml.Name{}, "bla blablab bla alba", dfxpStyle{TTSTextAlign: "center"}}
	output, err := xml.MarshalIndent(base, "  ", "    ")
	fmt.Println(err)
	fmt.Println(string(output))
}

func TestDFXPWriter(t *testing.T) {
	captionSet, err := NewDFXPReader().Read(sampleDFXP)
	assert.Nil(t, err)
	data, _ := NewDFXPWriter().Write(captionSet)
	output, err := xml.MarshalIndent(data, "  ", "    ")
	fmt.Println(sampleDFXP)
	fmt.Println("--------------------------")
	fmt.Println(string(output))
}
