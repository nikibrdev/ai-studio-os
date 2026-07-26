"use server";

import { revalidatePath } from "next/cache";
import { ApiError, completeTesting, planTask } from "./api";

// ActionResult — исход действия, возвращаемый значением, а не исключением.
//
// error.tsx — Error Boundary для отказов рендера: он заменяет страницу целиком.
// Для неудавшегося действия это неверно — человек теряет контекст и не видит,
// что именно не получилось. Поэтому ошибка возвращается сюда и показывается
// рядом с кнопкой, а действие можно повторить.
export type ActionResult = { ok: true } | { ok: false; error: string };

function failure(err: unknown): ActionResult {
  if (err instanceof ApiError) {
    return { ok: false, error: err.message };
  }
  // Недоступный API даёт не ApiError, а сетевую ошибку — её текст тоже
  // осмысленный («fetch failed»), и молчать о ней нельзя.
  return {
    ok: false,
    error: err instanceof Error ? err.message : "неизвестная ошибка",
  };
}

// refresh перечитывает данные, на которые повлияло решение: список ожидающих и
// страницу самой задачи. Без этого человек увидел бы прежнее состояние и решил,
// что действие не сработало.
function refresh(projectId: string, taskId: string): void {
  revalidatePath("/decisions");
  revalidatePath(`/projects/${projectId}`);
  revalidatePath(`/projects/${projectId}/tasks/${taskId}`);
}

// acceptDefinitionOfReady — первая контрольная точка: задача готова к работе
// (Backlog → Ready). Обратима существующим переходом, необратимых последствий
// не имеет.
export async function acceptDefinitionOfReady(
  projectId: string,
  taskId: string,
): Promise<ActionResult> {
  try {
    await planTask(projectId, taskId);
  } catch (err) {
    return failure(err);
  }
  refresh(projectId, taskId);
  return { ok: true };
}

// acceptTask — вторая контрольная точка: приёмочное решение. Платформа сливает
// Pull Request в основную ветку (ADR-008) и переводит задачу в Done. В рамках
// задачи это необратимо, поэтому интерфейс обязан предупредить до нажатия
// (docs/architecture/workflow.md).
export async function acceptTask(
  projectId: string,
  taskId: string,
): Promise<ActionResult> {
  try {
    await completeTesting(projectId, taskId, true);
  } catch (err) {
    return failure(err);
  }
  refresh(projectId, taskId);
  return { ok: true };
}

// returnForRework — отрицательный исход приёмочного решения: задача возвращается
// в In Progress. Выражается той же командой с passed=false — отрицательные
// исходы уже покрыты переходами state machine, новых команд не требуется.
export async function returnForRework(
  projectId: string,
  taskId: string,
): Promise<ActionResult> {
  try {
    await completeTesting(projectId, taskId, false);
  } catch (err) {
    return failure(err);
  }
  refresh(projectId, taskId);
  return { ok: true };
}
