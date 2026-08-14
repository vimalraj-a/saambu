export interface ElementInfo {
  tag: string
  text?: string
  selector: string
}

export interface ActionInfo {
  type: string
  selector?: string
  value?: string
  reasoning?: string
}

export interface PlaywrightCheck {
  type: string
  selector?: string
  expectedText?: string
}

export interface Assertion {
  description: string
  held: boolean
  explanation?: string
  playwrightCheck?: PlaywrightCheck
}

export interface StepExecution {
  index: number
  nlText: string
  action: ActionInfo
  screenshotAfterBase64?: string
  assertion?: Assertion
  error?: string
}

export interface Capture {
  id: string
  url: string
  title?: string
  prerequisiteText?: string
  prerequisiteSteps?: string[]
  prerequisiteRun?: StepExecution[]
  screenshotBase64: string
  htmlSnapshot?: string
  elements?: ElementInfo[]
  createdAt: string
}

export type ScriptStatus = 'draft' | 'verifying' | 'resolved_pass' | 'resolved_known_failure'

export interface TestScript {
  id: string
  captureId: string
  steps: string[]
  version: number
  status: ScriptStatus
  createdAt: string
  updatedAt: string
}

export type ChangeResponseType = 'clarification' | 'updated'

export interface ChangeRequest {
  id: string
  testScriptId: string
  message: string
  responseType: ChangeResponseType
  responseText?: string
  previousSteps?: string[]
  newSteps?: string[]
  createdAt: string
}

export interface ExecutionRun {
  id: string
  testScriptId: string
  scriptVersion: number
  steps: StepExecution[]
  overallPassed: boolean
  createdAt: string
}

export interface GeneratedTest {
  id: string
  testScriptId: string
  sourceRunId?: string
  code: string
  expectedToFail: boolean
  createdAt: string
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...(options?.headers ?? {}) },
  })
  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`
    try {
      const body = (await res.json()) as { error?: string }
      if (body?.error) message = body.error
    } catch {
      // response wasn't JSON — keep the status text
    }
    throw new Error(message)
  }
  return res.json() as Promise<T>
}

export function createCapture(url: string, prerequisiteText: string): Promise<Capture> {
  return request<Capture>('/api/captures', {
    method: 'POST',
    body: JSON.stringify({ url, prerequisiteText }),
  })
}

export function generateScript(captureId: string): Promise<TestScript> {
  return request<TestScript>(`/api/captures/${captureId}/script`, { method: 'POST' })
}

export function getScript(scriptId: string): Promise<{ script: TestScript; changeRequests: ChangeRequest[] }> {
  return request(`/api/scripts/${scriptId}`)
}

export function sendMessage(
  scriptId: string,
  message: string,
): Promise<{ changeRequest: ChangeRequest; script: TestScript }> {
  return request(`/api/scripts/${scriptId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ message }),
  })
}

export function verifyScript(scriptId: string): Promise<ExecutionRun> {
  return request<ExecutionRun>(`/api/scripts/${scriptId}/verify`, { method: 'POST' })
}

export function finalizeScript(scriptId: string, runId: string): Promise<GeneratedTest> {
  return request<GeneratedTest>(`/api/scripts/${scriptId}/finalize`, {
    method: 'POST',
    body: JSON.stringify({ runId }),
  })
}

export function lockKnownFailure(scriptId: string, runId: string): Promise<GeneratedTest> {
  return request<GeneratedTest>(`/api/scripts/${scriptId}/lock-known-failure`, {
    method: 'POST',
    body: JSON.stringify({ runId }),
  })
}

export function downloadUrl(generatedTestId: string): string {
  return `/api/generated-tests/${generatedTestId}/download`
}
