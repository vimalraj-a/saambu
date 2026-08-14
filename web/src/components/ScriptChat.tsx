import { useState, type FormEvent } from 'react'
import type { ChangeRequest, TestScript } from '../api'
import { sendMessage } from '../api'

interface ScriptChatProps {
  script: TestScript
  changeRequests: ChangeRequest[]
  onScriptUpdated: (script: TestScript, changeRequest: ChangeRequest) => void
  onVerify: () => void
  verifying: boolean
}

export function ScriptChat({ script, changeRequests, onScriptUpdated, onVerify, verifying }: ScriptChatProps) {
  const [message, setMessage] = useState('')
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSend(e: FormEvent) {
    e.preventDefault()
    if (!message.trim()) return
    setSending(true)
    setError(null)
    try {
      const { script: updatedScript, changeRequest } = await sendMessage(script.id, message)
      setMessage('')
      onScriptUpdated(updatedScript, changeRequest)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div>
        <h2 className="mb-2 text-xl font-semibold text-gray-900 dark:text-gray-100">Test script (v{script.version})</h2>
        <ol className="list-decimal space-y-1 pl-5 text-sm text-gray-800 dark:text-gray-200">
          {script.steps.map((step, i) => (
            <li key={i}>{step}</li>
          ))}
        </ol>
      </div>

      {changeRequests.length > 0 && (
        <div className="flex flex-col gap-2">
          {changeRequests.map((cr) => (
            <div key={cr.id} className="flex flex-col gap-1 text-sm">
              <p className="self-end rounded-md bg-indigo-600 px-3 py-1.5 text-white">{cr.message}</p>
              {cr.responseType === 'clarification' ? (
                <p className="self-start rounded-md bg-gray-100 px-3 py-1.5 text-gray-900 dark:bg-gray-800 dark:text-gray-100">
                  {cr.responseText}
                </p>
              ) : (
                <p className="self-start rounded-md bg-green-100 px-3 py-1.5 text-green-900 dark:bg-green-900 dark:text-green-100">
                  Spec updated.
                </p>
              )}
            </div>
          ))}
        </div>
      )}

      <form onSubmit={handleSend} className="flex gap-2">
        <input
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="Describe a change, e.g. 'also check that an empty password shows an error'"
          className="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"
        />
        <button
          type="submit"
          disabled={sending}
          className="rounded-md bg-gray-700 px-4 py-2 text-sm font-medium text-white hover:bg-gray-600 disabled:opacity-50"
        >
          Send
        </button>
      </form>
      {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

      <button
        onClick={onVerify}
        disabled={verifying}
        className="self-start rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
      >
        {verifying ? 'Verifying…' : 'Verify'}
      </button>
    </div>
  )
}
