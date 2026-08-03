import CodeMirror from '@uiw/react-codemirror'
import { json, jsonParseLinter } from '@codemirror/lang-json'
import { linter, lintGutter } from '@codemirror/lint'

export function JsonEditor({
  value,
  onChange,
  readOnly = false,
}: {
  value: string
  onChange?: (v: string) => void
  readOnly?: boolean
}) {
  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <CodeMirror
        value={value}
        theme="dark"
        height="60vh"
        readOnly={readOnly}
        extensions={[json(), linter(jsonParseLinter()), lintGutter()]}
        onChange={onChange}
        basicSetup={{ lineNumbers: true, foldGutter: true, highlightActiveLine: !readOnly }}
      />
    </div>
  )
}
