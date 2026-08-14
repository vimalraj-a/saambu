import { useState } from 'react'
import { CaptureForm } from './components/CaptureForm'
import { ScriptChat } from './components/ScriptChat'
import { RunStoryboard } from './components/RunStoryboard'
import { GeneratedTestView } from './components/GeneratedTestView'
import type { Capture, ChangeRequest, ExecutionRun, GeneratedTest, TestScript } from './api'
import { finalizeScript, lockKnownFailure, verifyScript } from './api'

type Stage = 'capture' | 'script' | 'run' | 'done'

function App() {
  const [stage, setStage] = useState<Stage>('capture')
  const [script, setScript] = useState<TestScript | null>(null)
  const [changeRequests, setChangeRequests] = useState<ChangeRequest[]>([])
  const [run, setRun] = useState<ExecutionRun | null>(null)
  const [generatedTest, setGeneratedTest] = useState<GeneratedTest | null>(null)
  const [verifying, setVerifying] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function handleScriptReady(_capture: Capture, newScript: TestScript) {
    setScript(newScript)
    setChangeRequests([])
    setStage('script')
  }

  function handleScriptUpdated(updatedScript: TestScript, changeRequest: ChangeRequest) {
    setScript(updatedScript)
    setChangeRequests((prev) => [...prev, changeRequest])
  }

  async function handleVerify() {
    if (!script) return
    setVerifying(true)
    setError(null)
    try {
      const newRun = await verifyScript(script.id)
      setRun(newRun)
      setStage('run')
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setVerifying(false)
    }
  }

  async function handleFinalize() {
    if (!script || !run) return
    setError(null)
    try {
      const gt = await finalizeScript(script.id, run.id)
      setGeneratedTest(gt)
      setStage('done')
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  async function handleLockKnownFailure() {
    if (!script || !run) return
    setError(null)
    try {
      const gt = await lockKnownFailure(script.id, run.id)
      setGeneratedTest(gt)
      setStage('done')
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  function handleFixDescription() {
    setStage('script')
  }

  return (
    <div className="min-h-screen bg-white px-4 py-10 dark:bg-gray-900">
      <div className="mx-auto flex max-w-3xl flex-col gap-8">
        <h1 className="text-3xl font-semibold text-gray-900 dark:text-gray-100">mongo-hack</h1>

        {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

        {stage === 'capture' && <CaptureForm onReady={handleScriptReady} />}

        {stage === 'script' && script && (
          <ScriptChat
            script={script}
            changeRequests={changeRequests}
            onScriptUpdated={handleScriptUpdated}
            onVerify={handleVerify}
            verifying={verifying}
          />
        )}

        {stage === 'run' && run && (
          <div className="flex flex-col gap-4">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
              Verification run — {run.overallPassed ? 'passed' : "didn't pass"}
            </h2>
            <RunStoryboard
              steps={run.steps}
              overallPassed={run.overallPassed}
              onLockKnownFailure={handleLockKnownFailure}
              onFixDescription={handleFixDescription}
            />
            {run.overallPassed && (
              <button
                onClick={handleFinalize}
                className="self-start rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500"
              >
                Generate Playwright test
              </button>
            )}
          </div>
        )}

        {stage === 'done' && generatedTest && <GeneratedTestView generatedTest={generatedTest} />}
      </div>
    </div>
  )
}

export default App
