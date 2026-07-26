import Link from "next/link";
import { decisionLabels, listAwaitingDecision } from "@/lib/api";

// Always render at request time — this page needs live data from apps/api,
// which is not reachable during `next build`'s static generation pass.
export const dynamic = "force-dynamic";

export default async function DecisionsPage() {
  const awaiting = await listAwaitingDecision();

  return (
    <div>
      <h1>Ждут решения</h1>

      {awaiting.length === 0 ? (
        // An empty list is a normal state of the system, not an error and not
        // an unexplained blank screen: nothing is waiting on a human.
        <p>Задач, ожидающих решения, нет.</p>
      ) : (
        <ul>
          {awaiting.map(({ decision, task }) => (
            <li key={`${task.projectId}/${task.id}`}>
              <Link href={`/projects/${task.projectId}/tasks/${task.id}`}>
                {task.id}: {task.title}
              </Link>{" "}
              — <span>{decisionLabels[decision]}</span>{" "}
              <span>({task.projectId})</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
