import type { GeneratedTest } from '../api'
import { downloadUrl } from '../api'

export function GeneratedTestView({ generatedTest }: { generatedTest: GeneratedTest }) {
  return (
    <div className="flex flex-col gap-3">
      <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
        Generated test{generatedTest.expectedToFail ? ' (known failure)' : ''}
      </h2>
      {generatedTest.expectedToFail && (
        <p className="text-sm text-amber-700 dark:text-amber-400">
          This test encodes the behavior you confirmed is correct — it's expected to fail until the underlying bug
          is fixed.
        </p>
      )}
      <pre className="max-h-96 overflow-auto rounded-md bg-gray-900 p-4 text-xs text-gray-100">
        <code>{generatedTest.code}</code>
      </pre>
      <a
        href={downloadUrl(generatedTest.id)}
        className="self-start rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500"
      >
        Download .spec.ts
      </a>
    </div>
  )
}
