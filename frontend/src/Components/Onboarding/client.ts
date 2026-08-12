import { getSessionId, getSubjectId } from "./storage";
import type{ ResolveConfig } from "../../types/sdk";

function buildUrl(apiUrl: string) {
  return `${apiUrl.replace(/\/$/, "")}/resolve`;
}

function buildHeaders(appKey: string): HeadersInit {
  return {
    "Content-Type": "application/json",
    "X-App-Key": appKey,
  };
}

export async function resolveTour(config: ResolveConfig) {
  const controller = new AbortController();
  const { signal } = controller;
  window.addEventListener("beforeunload", () => controller.abort());

  const body = {
    url: window.location.href,
    subject_id: getSubjectId(config.subjectId),
    session_id: getSessionId(),
    props: config.props ?? {},
  };


  const response = await fetch(buildUrl(config.apiUrl), {
    method: "POST",
    headers: buildHeaders(config.appKey),
    body: JSON.stringify(body),
    signal,
  });

  if (response.status === 204) return null;
  if (!response.ok) throw new Error(`Resolve failed: ${response.status} ${await response.text()}`);

  return await response.json();
}