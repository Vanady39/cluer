const SUBJECT_KEY = "onboarding_subject_id";
const SESSION_KEY = "onboarding_session_id";

function createId(prefix: string): string {
  const uuid = crypto.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).substring(2)}`;
  return `${prefix}_${uuid}`;
}

export function hasSubjectId() {
  return localStorage.getItem(SUBJECT_KEY) !== null;
}

export function getSubjectId(providedSubjectId?: string) {
  if (providedSubjectId) return providedSubjectId;

  const existing = localStorage.getItem(SUBJECT_KEY);
  if (existing) return existing;

  const subjectId = createId("anon");
  localStorage.setItem(SUBJECT_KEY, subjectId);
  return subjectId;
}

export function getSessionId() {
  const existing = sessionStorage.getItem(SESSION_KEY);
  if (existing) return existing;

  const sessionId = createId("sess");
  sessionStorage.setItem(SESSION_KEY, sessionId);
  return sessionId;
}