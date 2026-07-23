import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = {
  title: "AI Studio OS — Dashboard",
  description: "Наблюдение за проектами и задачами AI Studio OS",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ru">
      <body>
        <header className="site-header">
          <nav>
            <Link href="/">AI Studio OS</Link>
          </nav>
        </header>
        <main>{children}</main>
      </body>
    </html>
  );
}
