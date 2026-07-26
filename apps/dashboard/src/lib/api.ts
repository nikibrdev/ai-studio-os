// Typed client for apps/api (docs/api/). Types mirror the response shapes
// documented there by hand — no OpenAPI generation (EPIC-008/009 leave
// that out, see EPIC-009 "Не входит").

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export type ProjectState = "created" | "active" | "archived";

export interface Project {
  id: string;
  name: string;
  state: ProjectState;
  createdAt: string;
}

// DecisionKind — какое решение человека ждёт задача (docs/api/decisions.md).
// Пустая строка означает, что решение не требуется. Значение приходит с
// сервера: соответствие состояния и решения — правило платформы, и выводить
// его здесь из state запрещено (docs/architecture/module-boundaries.md —
// «дублирование доменных правил в UI»).
export type DecisionKind = "definition-of-ready" | "acceptance" | "";

export interface TaskView {
  id: string;
  projectId: string;
  state: string;
  updatedAt: string;
  title: string;
  type: string;
  scope: string;
  acceptanceCriteria: string[];
  awaitingDecision: DecisionKind;
}

// AwaitingDecision — одна задача, ждущая решения человека
// (docs/api/decisions.md, «Список ожидающих решения»).
export interface AwaitingDecision {
  decision: Exclude<DecisionKind, "">;
  task: TaskView;
}

// Человекочитаемые названия решений — единственное место, где они заданы,
// чтобы формулировка не расходилась между экранами.
export const decisionLabels: Record<Exclude<DecisionKind, "">, string> = {
  "definition-of-ready": "Принять Definition of Ready",
  acceptance: "Приёмочное решение",
};

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, { cache: "no-store" });
  if (!res.ok) {
    throw new ApiError(
      res.status,
      `GET ${path}: ${res.status} ${res.statusText}`,
    );
  }
  return (await res.json()) as T;
}

// listProjects — docs/api/projects.md, "Список проектов".
export function listProjects(): Promise<Project[]> {
  return get<Project[]>("/projects");
}

// listProjectTasks — docs/api/tasks.md, "Список задач проекта".
export function listProjectTasks(projectId: string): Promise<TaskView[]> {
  return get<TaskView[]>(`/projects/${encodeURIComponent(projectId)}/tasks`);
}

// listAwaitingDecision — docs/api/decisions.md, «Список ожидающих решения».
// Сквозной по проектам: отвечает на «что ждёт решения вообще».
export async function listAwaitingDecision(): Promise<AwaitingDecision[]> {
  const body = await get<{ decisions: AwaitingDecision[] }>("/decisions");
  return body.decisions;
}

// getTask — docs/api/tasks.md, "Получить состояние задачи".
export function getTask(projectId: string, taskId: string): Promise<TaskView> {
  return get<TaskView>(
    `/projects/${encodeURIComponent(projectId)}/tasks/${encodeURIComponent(taskId)}`,
  );
}
