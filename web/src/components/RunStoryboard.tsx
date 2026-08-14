import type { StepExecution } from '../api'

interface RunStoryboardProps {
  steps: StepExecution[]
  showAssertions?: boolean
  overallPassed?: boolean
  onLockKnownFailure?: () => void
  onFixDescription?: () => void
}

export function RunStoryboard({
  steps,
  showAssertions = true,
  overallPassed,
  onLockKnownFailure,
  onFixDescription,
}: RunStoryboardProps) {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-3">
        {steps.map((step) => (
          <div key={step.index} className="rounded-md border border-gray-200 p-3 dark:border-gray-700">
            <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
              {step.index + 1}. {step.nlText}
            </p>
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {step.action.type}
              {step.action.selector ? ` -> ${step.action.selector}` : ''}
              {step.action.value ? ` = "${step.action.value}"` : ''}
            </p>
            {step.error && <p className="mt-1 text-xs text-red-600 dark:text-red-400">Error: {step.error}</p>}
            {showAssertions && step.assertion && (
              <p
                className={`mt-1 text-xs ${
                  step.assertion.held ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
                }`}
              >
                {step.assertion.held ? '✓ held' : '✗ did not hold'} — {step.assertion.description}
                {step.assertion.explanation ? ` (${step.assertion.explanation})` : ''}
              </p>
            )}
            {step.screenshotAfterBase64 && (
              <img
                src={`data:image/png;base64,${step.screenshotAfterBase64}`}
                alt={`After: ${step.nlText}`}
                className="mt-2 max-h-64 rounded border border-gray-200 dark:border-gray-700"
              />
            )}
          </div>
        ))}
      </div>

      {overallPassed === false && (onLockKnownFailure || onFixDescription) && (
        <div className="flex flex-col gap-2 rounded-md border border-amber-300 bg-amber-50 p-4 dark:border-amber-700 dark:bg-amber-950">
          <p className="text-sm font-medium text-amber-900 dark:text-amber-200">This didn't verify. Which is right?</p>
          <div className="flex gap-2">
            <button
              onClick={onLockKnownFailure}
              className="rounded-md bg-amber-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-amber-500"
            >
              My description is right — lock as known bug
            </button>
            <button
              onClick={onFixDescription}
              className="rounded-md border border-amber-400 px-3 py-1.5 text-sm font-medium text-amber-900 hover:bg-amber-100 dark:text-amber-200 dark:hover:bg-amber-900"
            >
              Let me fix the description
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
