import {
  getSessionId,
  getSubjectId,
} from "./storage";

interface ResolveConfig {
  apiUrl: string;
  appKey: string;
  subjectId?: string;
  props?: Record<string, unknown>;
}

export async function resolveTour(
  config: ResolveConfig,
) {
  const body = {
    url: window.location.href,
    subject_id: getSubjectId(config.subjectId),
    session_id: getSessionId(),
    props: config.props ?? {},
  };

  console.log(
    "[Onboarding SDK] RESOLVE REQUEST",
    body,
  );

  const response = await fetch(
    `${config.apiUrl}/v1/resolve`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-App-Key": config.appKey,
      },
      body: JSON.stringify(body),
    },
  );

  if (response.status === 204) {
    console.log(
      "[Onboarding SDK] RESOLVE 204",
    );

    return null;
  }

  if (!response.ok) {
    const text = await response.text();

    throw new Error(
      `Resolve failed: ${response.status} ${text}`,
    );
  }

  const data = await response.json();

  console.log(
    "[Onboarding SDK] RESOLVE 200",
    data,
  );

  return data;
}