import "./globals.css";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Lara Web Editor",
  description: "Edit Lara config.yml and mocks.json",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
