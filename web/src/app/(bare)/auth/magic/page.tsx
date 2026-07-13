"use client";

import { useEffect, useState } from "react";

/**
 * Magic-link entry for easy-easy SSO (SSH → JWT → bridge → Hydra → Storyden).
 *
 * Expected URL: /auth/magic?login_token=<JWT>
 * Flow:
 *   1. POST /v1/auth/oidc/bootstrap → authorize_url
 *   2. Redirect to bridge /realms/main/start?login_token=…&return_to=authorize_url
 *   3. Bridge + Hydra complete OIDC; callback lands on Storyden session
 */
export default function MagicAuthPage() {
  const [message, setMessage] = useState("Starting secure sign-in…");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function run() {
      const params = new URLSearchParams(window.location.search);
      const token = params.get("login_token");
      if (!token) {
        setError(
          "Missing login_token. Open the member dashboard via the SSH login link, then click Forum again within a few minutes.",
        );
        setMessage("");
        return;
      }

      try {
        setMessage("Requesting authorization…");
        const res = await fetch("/v1/auth/oidc/bootstrap", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Accept: "application/json",
          },
          body: JSON.stringify({ redirect_path: "/" }),
          credentials: "same-origin",
        });
        if (!res.ok) {
          const text = await res.text();
          throw new Error(`Bootstrap failed (${res.status}): ${text.slice(0, 200)}`);
        }
        const data = (await res.json()) as { authorize_url?: string };
        if (!data.authorize_url) {
          throw new Error("Bootstrap response missing authorize_url");
        }

        if (cancelled) return;
        setMessage("Redirecting to identity provider…");

        const start =
          "https://daltons-office.easyeasyspeakeasy.com/realms/main/start" +
          "?login_token=" +
          encodeURIComponent(token) +
          "&return_to=" +
          encodeURIComponent(data.authorize_url);

        window.location.replace(start);
      } catch (e) {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : String(e));
        setMessage("");
      }
    }

    void run();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <main
      style={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        fontFamily: "system-ui, sans-serif",
        background: "#0f1115",
        color: "#e8eaed",
        padding: "1.5rem",
      }}
    >
      <div style={{ maxWidth: "28rem", textAlign: "center" }}>
        <h1 style={{ fontSize: "1.25rem", marginBottom: "0.75rem" }}>
          easy-easy forum sign-in
        </h1>
        {message ? (
          <p style={{ color: "#9aa0a6", lineHeight: 1.5 }}>{message}</p>
        ) : null}
        {error ? (
          <p style={{ color: "#f87171", lineHeight: 1.5 }}>{error}</p>
        ) : null}
      </div>
    </main>
  );
}
