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

  // Ссылка на Pull Request под ревью — пустая, пока он неизвестен
  // (BUGFIX-009). Нужна, чтобы человек видел, что именно сольёт приёмочное
  // решение: в рамках задачи оно необратимо.
  repository: string;
  pullRequestId: string;
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

// post отправляет команду в apps/api. Вызывается только из Server Actions
// (src/lib/actions.ts): у apps/api нет CORS-заголовков, поэтому запрос из
// браузера был бы заблокирован — запись идёт через сервер Next.js, который и
// так общается с API.
async function post(path: string, body?: unknown): Promise<void> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    method: "POST",
    cache: "no-store",
    headers:
      body === undefined ? undefined : { "Content-Type": "application/json" },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (!res.ok) {
    // Тело ошибки API — {"error": "..."}; показываем его человеку, а не общий
    // код статуса: «переход недопустим» и «задача не найдена» требуют разных
    // действий от того, кто это увидит.
    let detail = `${res.status} ${res.statusText}`;
    try {
      const parsed = (await res.json()) as { error?: string };
      if (parsed.error) detail = parsed.error;
    } catch {
      // Тело не JSON — остаётся статус.
    }
    throw new ApiError(res.status, detail);
  }
}

// planTask — docs/api/tasks.md, «Запланировать задачу»: приём Definition of
// Ready, первая контрольная точка человека.
export function planTask(projectId: string, taskId: string): Promise<void> {
  return post(
    `/projects/${encodeURIComponent(projectId)}/tasks/${encodeURIComponent(taskId)}/plan`,
  );
}

// completeTesting — docs/api/tasks.md, «Завершить тестирование»: приёмочное
// решение, вторая контрольная точка. При passed=true платформа сливает Pull
// Request (ADR-008); ссылку она знает сама (BUGFIX-009), поэтому здесь её нет.
export function completeTesting(
  projectId: string,
  taskId: string,
  passed: boolean,
): Promise<void> {
  return post(
    `/projects/${encodeURIComponent(projectId)}/tasks/${encodeURIComponent(taskId)}/complete-testing`,
    { passed },
  );
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
