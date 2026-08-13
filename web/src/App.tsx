import { useEffect, useState } from 'react'

type HealthStatus = 'checking' | 'ok' | 'down'

function App() {
  const [status, setStatus] = useState<HealthStatus>('checking')

  useEffect(() => {
    fetch('/health')
      .then((res) => setStatus(res.ok ? 'ok' : 'down'))
      .catch(() => setStatus('down'))
  }, [])

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-white dark:bg-gray-900">
      <h1 className="text-3xl font-semibold text-gray-900 dark:text-gray-100">
        mongo-hack
      </h1>
      <p className="text-gray-600 dark:text-gray-400">
        API status:{' '}
        <span
          className={
            status === 'ok'
              ? 'text-green-600 dark:text-green-400'
              : status === 'down'
                ? 'text-red-600 dark:text-red-400'
                : 'text-gray-500'
          }
        >
          {status}
        </span>
      </p>
    </div>
  )
}

export default App
