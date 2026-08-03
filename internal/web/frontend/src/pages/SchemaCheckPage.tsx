import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { CheckCircle2, RefreshCw } from 'lucide-react'
import { api } from '../lib/api'
import type { SchemaCheckResult } from '../lib/types'
import { Card } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Spinner } from '../components/ui/spinner'

export function SchemaCheckPage() {
  const [result, setResult] = useState<SchemaCheckResult | null>(null)
  const check = useMutation({
    mutationFn: api.schemaCheck,
    onSuccess: setResult,
  })

  return (
    <div className="mx-auto max-w-4xl space-y-5">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-text">Schema check</h1>
        <Button variant="primary" onClick={() => check.mutate()} disabled={check.isPending}>
          {check.isPending ? <Spinner /> : <RefreshCw className="h-4 w-4" />} Run check
        </Button>
      </div>

      <Card>
        <p className="text-sm text-muted">
          Compares the omo document schema embedded in this binary against the latest upstream{" "}
          <span className="font-mono text-text">omo.schema.json</span>. Drift means upstream changed; re-sync the
          embedded schema and rebuild to stay aligned.
        </p>
      </Card>

      {check.isError && <p className="text-sm text-danger">{(check.error as Error).message}</p>}

      {result &&
        (result.identical ? (
          <Card className="border-success/40 bg-success/10">
            <div className="flex items-center gap-2 text-success">
              <CheckCircle2 className="h-5 w-5" />
              <span className="text-sm font-medium">In sync — the embedded schema matches upstream.</span>
            </div>
          </Card>
        ) : (
          <Card className="p-0">
            <div className="border-b border-border px-4 py-2 text-sm font-medium text-warn">
              Drift detected — upstream differs from the embedded schema.
            </div>
            <pre className="max-h-[60vh] overflow-auto scrollbar-thin p-4 font-mono text-xs text-text">{result.diff}</pre>
          </Card>
        ))}
    </div>
  )
}
