import Link from "next/link";
import { listProjectTasks } from "@/lib/api";

// Always render at request time — this page needs live data from apps/api,
// which is not reachable during `next build`'s static generation pass
// (see TASK-075's Отчёт).
export const dynamic = "force-dynamic";

export default async function ProjectPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const tasks = await listProjectTasks(id);

  return (
    <div>
      <h1>Проект {id}</h1>
      {tasks.length === 0 ? (
        <p>Задач пока нет.</p>
      ) : (
        <ul>
          {tasks.map((task) => (
            <li key={task.id}>
              <Link href={`/projects/${id}/tasks/${task.id}`}>
                {task.id} — {task.title}
              </Link>{" "}
              — <span>{task.state}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
