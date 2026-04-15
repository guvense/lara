import { NextResponse } from "next/server";

const backendBase = process.env.LARA_BACKEND_URL ?? "http://localhost:8898";

export async function GET() {
  try {
    const response = await fetch(`${backendBase}/api/editor`, { cache: "no-store" });
    const body = await response.text();
    return new NextResponse(body, {
      status: response.status,
      headers: { "content-type": "application/json" },
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : "backend unavailable";
    return NextResponse.json({ error: message }, { status: 502 });
  }
}
