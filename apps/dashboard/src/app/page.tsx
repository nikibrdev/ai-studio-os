import Link from "next/link";
import { listProjects } from "@/lib/api";

// Always render at request time — this page needs live data from apps/api,
// which is not reachable during `next build`'s static generation pass.
export const dynamic = "force-dynamic";

export default async function HomePage() {
  const projects = await listProjects();

  if (projects.length === 0) {
    return <p>Проектов пока нет.</p>;
  }

  return (
    <div>
      <h1>Проекты</h1>
      <ul>
        {projects.map((project) => (
          <li key={project.id}>
            <Link href={`/projects/${project.id}`}>{project.name}</Link> —{" "}
            <span>{project.state}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
