import { useState, type FormEvent } from 'react'
import type { Capture, TestScript } from '../api'
import { createCapture, generateScript } from '../api'
import { RunStoryboard } from './RunStoryboard'

type Phase = 'form' | 'capturing' | 'reviewSetup' | 'generating'

export function CaptureForm({ onReady }: { onReady: (capture: Capture, script: TestScript) => void }) {
  const [url, setUrl] = useState('')
  const [prerequisiteText, setPrerequisiteText] = useState('')
  const [phase, setPhase] = useState<Phase>('form')
  const [capture, setCapture] = useState<Capture | null>(null)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setPhase('capturing')
    try {
      const cap = await createCapture(url, prerequisiteText)
      setCapture(cap)
      if (cap.prerequisiteRun && cap.prerequisiteRun.length > 0) {
        setPhase('reviewSetup')
        return
      }
      setPhase('generating')
      const script = await generateScript(cap.id)
      onReady(cap, script)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
      setPhase('form')
    }
  }

  async function handleContinue() {
    if (!capture) return
    setError(null)
    setPhase('generating')
    try {
      const script = await generateScript(capture.id)
      onReady(capture, script)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
      setPhase('reviewSetup')
    }
  }

  if (phase === 'reviewSetup' && capture) {
    return (
      <div className="flex flex-col gap-4">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">Setup run</h2>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Here's how the agent reached the page under test, starting from {capture.url}.
        </p>
        <RunStoryboard steps={capture.prerequisiteRun ?? []} showAssertions={false} />
        {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}
        <button
          onClick={handleContinue}
          className="self-start rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500"
        >
          Looks right — generate a test script
        </button>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      <div>
        <label className="mb-1 block text-sm font-medium text-gray-900 dark:text-gray-100">URL</label>
        <input
          type="url"
          required
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://example.com"
          className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"
        />
      </div>
      <div>
        <label className="mb-1 block text-sm font-medium text-gray-900 dark:text-gray-100">
          How to reach the page under test (optional)
        </label>
        <textarea
          value={prerequisiteText}
          onChange={(e) => setPrerequisiteText(e.target.value)}
          placeholder="e.g. log in with test@example.com / password123, then open Settings"
          rows={3}
          className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"
        />
      </div>
      {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}
      <button
        type="submit"
        disabled={phase !== 'form'}
        className="self-start rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
      >
        {phase === 'capturing' ? 'Capturing…' : phase === 'generating' ? 'Generating script…' : 'Capture'}
      </button>
    </form>
  )
}
