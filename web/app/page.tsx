"use client";

import { useEffect, useMemo, useState } from "react";

type EditorPayload = {
  config: string;
  mocks: string;
  mocksFile: string;
};

type Step = 1 | 2 | 3 | 4;

type MockDraft = {
  endpoint: string;
  method: string;
  responseStatus: number;
  responseDelay: string;
  responseBody: string;
  callbackRestEnabled: boolean;
  callbackRestEndpoint: string;
  callbackRestMethod: string;
  callbackRestDelay: string;
  callbackRestBody: string;
  callbackAuthEnabled: boolean;
  callbackAuthKey: string;
  callbackBrokerEnabled: boolean;
  callbackBrokerType: "kafka" | "rabbitmq";
  callbackBrokerDelay: string;
  callbackBrokerMessage: string;
  callbackKafkaTopic: string;
  callbackKafkaKey: string;
  callbackKafkaBrokers: string;
  callbackRabbitURL: string;
  callbackRabbitExchange: string;
  callbackRabbitRoutingKey: string;
};

const defaultDraft: MockDraft = {
  endpoint: "/users",
  method: "GET",
  responseStatus: 200,
  responseDelay: "0ms",
  responseBody: "{\n  \"ok\": true\n}",
  callbackRestEnabled: false,
  callbackRestEndpoint: "",
  callbackRestMethod: "POST",
  callbackRestDelay: "1000ms",
  callbackRestBody: "{\n  \"event\": \"created\"\n}",
  callbackAuthEnabled: false,
  callbackAuthKey: "",
  callbackBrokerEnabled: false,
  callbackBrokerType: "kafka",
  callbackBrokerDelay: "10000ms",
  callbackBrokerMessage: "{\n  \"event\": \"created\"\n}",
  callbackKafkaTopic: "users.events",
  callbackKafkaKey: "${value::request::body::username}",
  callbackKafkaBrokers: "localhost:9092",
  callbackRabbitURL: "amqp://guest:guest@localhost:5672/",
  callbackRabbitExchange: "events",
  callbackRabbitRoutingKey: "users.created",
};

function parseJSONObject(input: string): unknown {
  const trimmed = input.trim();
  if (!trimmed) {
    return {};
  }
  return JSON.parse(trimmed);
}

function ensureJSONArray(input: string): unknown[] {
  const trimmed = input.trim();
  if (!trimmed) {
    return [];
  }
  const parsed = JSON.parse(trimmed);
  if (!Array.isArray(parsed)) {
    throw new Error("mocks.json must be an array");
  }
  return parsed;
}

function upsertTokenGenerator(configText: string, key: string): string {
  const config = configText.trim();
  if (!key.trim()) {
    return configText;
  }

  const tokenBlock =
    "token-generator:\n" +
    `  ${key}:\n` +
    "    type: Password-Credential\n" +
    "    token-url: http://auth.local/token\n" +
    "    client-id: client-id\n" +
    "    client-secret: client-secret\n" +
    "    username: username\n" +
    "    password: password\n" +
    "    scope: read\n";

  if (!config.includes("token-generator:")) {
    const prefix = config.endsWith("\n") || config === "" ? config : `${config}\n`;
    return `${prefix}${tokenBlock}`;
  }

  if (config.includes(`\n  ${key}:`)) {
    return configText;
  }

  const appended = config.endsWith("\n") ? config : `${config}\n`;
  return `${appended}  ${key}:\n    type: Password-Credential\n    token-url: http://auth.local/token\n    client-id: client-id\n    client-secret: client-secret\n    username: username\n    password: password\n    scope: read\n`;
}

function validateDraft(step: Step, draft: MockDraft): string | null {
  if (step === 1) {
    if (!draft.endpoint.trim()) {
      return "Step 1: endpoint is required";
    }
    if (!draft.endpoint.startsWith("/")) {
      return "Step 1: endpoint must start with /";
    }
    return null;
  }

  if (step === 2) {
    if (draft.responseStatus < 100 || draft.responseStatus > 599) {
      return "Step 2: response status must be between 100 and 599";
    }
    if (!draft.responseDelay.trim()) {
      return "Step 2: response delay is required";
    }
    try {
      parseJSONObject(draft.responseBody);
    } catch {
      return "Step 2: response body must be valid JSON";
    }
    return null;
  }

  if (step === 3) {
    if (!draft.callbackRestEnabled) {
      return null;
    }
    if (!draft.callbackRestEndpoint.trim()) {
      return "Step 3: callback endpoint is required";
    }
    if (!draft.callbackRestDelay.trim()) {
      return "Step 3: callback delay is required";
    }
    if (draft.callbackAuthEnabled && !draft.callbackAuthKey.trim()) {
      return "Step 3: token generator key is required when auth is enabled";
    }
    try {
      parseJSONObject(draft.callbackRestBody);
    } catch {
      return "Step 3: callback body must be valid JSON";
    }
    return null;
  }

  if (!draft.callbackBrokerEnabled) {
    return null;
  }
  if (!draft.callbackBrokerDelay.trim()) {
    return "Step 4: broker delay is required";
  }
  try {
    parseJSONObject(draft.callbackBrokerMessage);
  } catch {
    return "Step 4: broker message must be valid JSON";
  }
  if (draft.callbackBrokerType === "kafka") {
    const brokers = draft.callbackKafkaBrokers
      .split(",")
      .map((x) => x.trim())
      .filter(Boolean);
    if (brokers.length === 0) {
      return "Step 4: at least one Kafka broker is required";
    }
    if (!draft.callbackKafkaTopic.trim()) {
      return "Step 4: Kafka topic is required";
    }
    return null;
  }

  if (!draft.callbackRabbitURL.trim()) {
    return "Step 4: RabbitMQ URL is required";
  }
  if (!draft.callbackRabbitRoutingKey.trim()) {
    return "Step 4: RabbitMQ routing key is required";
  }
  return null;
}

function buildMockObject(draft: MockDraft): unknown {
  const callbacks: Array<Record<string, unknown>> = [];

  if (draft.callbackRestEnabled) {
    const rest: Record<string, unknown> = {
      after: draft.callbackRestDelay,
      request: {
        endpoint: draft.callbackRestEndpoint,
        method: draft.callbackRestMethod,
        headers: { "content-type": "application/json" },
        body: parseJSONObject(draft.callbackRestBody),
      },
    };

    if (draft.callbackAuthEnabled && draft.callbackAuthKey.trim()) {
      rest["token-generator"] = draft.callbackAuthKey.trim();
    }
    callbacks.push({ rest });
  }

  if (draft.callbackBrokerEnabled) {
    if (draft.callbackBrokerType === "kafka") {
      const brokers = draft.callbackKafkaBrokers
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean);

      callbacks.push({
        "message-broker": {
          type: "kafka",
          after: draft.callbackBrokerDelay,
          connection: { brokers },
          message: parseJSONObject(draft.callbackBrokerMessage),
          kafka: {
            topic: draft.callbackKafkaTopic,
            key: draft.callbackKafkaKey,
          },
        },
      });
    } else {
      callbacks.push({
        "message-broker": {
          type: "rabbitmq",
          after: draft.callbackBrokerDelay,
          connection: { url: draft.callbackRabbitURL },
          message: parseJSONObject(draft.callbackBrokerMessage),
          rabbitmq: {
            exchange: draft.callbackRabbitExchange,
            "routing-key": draft.callbackRabbitRoutingKey,
          },
        },
      });
    }
  }

  return {
    rest: {
      request: {
        endpoint: draft.endpoint,
        method: draft.method,
        headers: { "content-type": "application/json" },
      },
      response: {
        status: draft.responseStatus,
        after: draft.responseDelay,
        body: parseJSONObject(draft.responseBody),
      },
      callback: callbacks,
    },
  };
}

const stepItems: Array<{ id: Step; title: string }> = [
  { id: 1, title: "Request" },
  { id: 2, title: "Response" },
  { id: 3, title: "REST Callback" },
  { id: 4, title: "Broker Callback" },
];

export default function HomePage() {
  const [config, setConfig] = useState("");
  const [mocks, setMocks] = useState("");
  const [mocksFile, setMocksFile] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [showBuilder, setShowBuilder] = useState(false);
  const [step, setStep] = useState<Step>(1);
  const [draft, setDraft] = useState<MockDraft>(defaultDraft);
  const [status, setStatus] = useState<{ text: string; type: "ok" | "err" | "" }>({
    text: "Loading editor content...",
    type: "",
  });

  const loadUrl = useMemo(() => "/api/editor", []);
  const saveUrl = useMemo(() => "/api/editor/save", []);

  useEffect(() => {
    const load = async () => {
      try {
        const response = await fetch(loadUrl, { cache: "no-store" });
        if (!response.ok) {
          throw new Error(`Load failed (${response.status})`);
        }
        const data: EditorPayload = await response.json();
        setConfig(data.config ?? "");
        setMocks(data.mocks ?? "");
        setMocksFile(data.mocksFile ?? "");
        setStatus({ text: "Loaded. Use stepper builder and then save.", type: "ok" });
      } catch (error) {
        const message = error instanceof Error ? error.message : "Failed to load";
        setStatus({ text: message, type: "err" });
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [loadUrl]);

  const goNext = () => {
    const validationError = validateDraft(step, draft);
    if (validationError) {
      setStatus({ text: validationError, type: "err" });
      return;
    }
    setStep((s: Step) => (s < 4 ? ((s + 1) as Step) : s));
    setStatus({ text: "", type: "" });
  };

  const goBack = () => {
    setStep((s: Step) => (s > 1 ? ((s - 1) as Step) : s));
    setStatus({ text: "", type: "" });
  };

  const applyBuilder = () => {
    for (const s of [1, 2, 3, 4] as Step[]) {
      const error = validateDraft(s, draft);
      if (error) {
        setStep(s);
        setStatus({ text: error, type: "err" });
        return;
      }
    }

    try {
      const arr = ensureJSONArray(mocks);
      arr.push(buildMockObject(draft));
      setMocks(JSON.stringify(arr, null, 2));

      if (draft.callbackAuthEnabled && draft.callbackAuthKey.trim()) {
        setConfig((current: string) => upsertTokenGenerator(current, draft.callbackAuthKey.trim()));
      }

      setShowBuilder(false);
      setStep(1);
      setStatus({ text: "Mock service generated and appended to mocks.json.", type: "ok" });
    } catch (error) {
      const message = error instanceof Error ? error.message : "Builder conversion failed";
      setStatus({ text: message, type: "err" });
    }
  };

  const save = async () => {
    setSaving(true);
    setStatus({ text: "Saving files...", type: "" });
    try {
      const response = await fetch(saveUrl, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ config, mocks, mocksFile }),
      });
      const result = (await response.json()) as { error?: string; mocksFile?: string };
      if (!response.ok) {
        throw new Error(result.error ?? "Save failed");
      }
      if (result.mocksFile) {
        setMocksFile(result.mocksFile);
      }
      setStatus({ text: "Saved. Watcher will restart backend automatically.", type: "ok" });
    } catch (error) {
      const message = error instanceof Error ? error.message : "Save failed";
      setStatus({ text: message, type: "err" });
    } finally {
      setSaving(false);
    }
  };

  return (
    <main className="page">
      <header className="head">
        <div>
          <h1>Lara Web Editor</h1>
          <div className="sub">Stepper builder for request, response, callback and authentication.</div>
        </div>
        <div className="actions">
          <input
            className="path"
            value={mocksFile}
            onChange={(e) => setMocksFile(e.target.value)}
            placeholder="mocks file path"
          />
          <button className="btn ghost" onClick={() => setShowBuilder((v: boolean) => !v)}>
            {showBuilder ? "Close Builder" : "Create Mock Service"}
          </button>
          <button className="btn" onClick={save} disabled={saving || loading}>
            {saving ? "Saving..." : "Save"}
          </button>
        </div>
      </header>

      <div className={`status ${status.type}`}>{status.text}</div>

      {showBuilder && (
        <section className="builder">
          <div className="builder-title">Mock Service Wizard</div>

          <div className="steps">
            {stepItems.map((item) => (
              <button
                key={item.id}
                className={`step ${step === item.id ? "active" : ""}`}
                onClick={() => setStep(item.id)}
                type="button"
              >
                {item.id}. {item.title}
              </button>
            ))}
          </div>

          {step === 1 && (
            <div className="builder-grid">
              <label>
                Endpoint
                <input value={draft.endpoint} onChange={(e) => setDraft({ ...draft, endpoint: e.target.value })} />
              </label>
              <label>
                Method
                <select value={draft.method} onChange={(e) => setDraft({ ...draft, method: e.target.value })}>
                  {["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"].map((m) => (
                    <option key={m} value={m}>
                      {m}
                    </option>
                  ))}
                </select>
              </label>
            </div>
          )}

          {step === 2 && (
            <>
              <div className="builder-grid">
                <label>
                  Response Status
                  <input
                    type="number"
                    value={draft.responseStatus}
                    onChange={(e) => setDraft({ ...draft, responseStatus: Number(e.target.value || 200) })}
                  />
                </label>
                <label>
                  Response Delay
                  <input
                    value={draft.responseDelay}
                    onChange={(e) => setDraft({ ...draft, responseDelay: e.target.value })}
                  />
                </label>
              </div>
              <label className="full">
                Response Body JSON
                <textarea
                  className="builder-json"
                  value={draft.responseBody}
                  onChange={(e) => setDraft({ ...draft, responseBody: e.target.value })}
                />
              </label>
            </>
          )}

          {step === 3 && (
            <>
              <label className="inline-check">
                <input
                  type="checkbox"
                  checked={draft.callbackRestEnabled}
                  onChange={(e) => setDraft({ ...draft, callbackRestEnabled: e.target.checked })}
                />
                Enable REST callback
              </label>

              {draft.callbackRestEnabled && (
                <>
                  <div className="builder-grid">
                    <label>
                      Callback Endpoint
                      <input
                        value={draft.callbackRestEndpoint}
                        onChange={(e) => setDraft({ ...draft, callbackRestEndpoint: e.target.value })}
                      />
                    </label>
                    <label>
                      Callback Method
                      <select
                        value={draft.callbackRestMethod}
                        onChange={(e) => setDraft({ ...draft, callbackRestMethod: e.target.value })}
                      >
                        {["POST", "PUT", "PATCH", "GET", "DELETE"].map((m) => (
                          <option key={m} value={m}>
                            {m}
                          </option>
                        ))}
                      </select>
                    </label>
                    <label>
                      Callback Delay
                      <input
                        value={draft.callbackRestDelay}
                        onChange={(e) => setDraft({ ...draft, callbackRestDelay: e.target.value })}
                      />
                    </label>
                  </div>

                  <label className="inline-check">
                    <input
                      type="checkbox"
                      checked={draft.callbackAuthEnabled}
                      onChange={(e) => setDraft({ ...draft, callbackAuthEnabled: e.target.checked })}
                    />
                    Enable callback authentication
                  </label>

                  {draft.callbackAuthEnabled && (
                    <label>
                      Token Generator Key
                      <input
                        value={draft.callbackAuthKey}
                        onChange={(e) => setDraft({ ...draft, callbackAuthKey: e.target.value })}
                        placeholder="example: keycloak"
                      />
                    </label>
                  )}

                  <label className="full">
                    Callback Body JSON
                    <textarea
                      className="builder-json"
                      value={draft.callbackRestBody}
                      onChange={(e) => setDraft({ ...draft, callbackRestBody: e.target.value })}
                    />
                  </label>
                </>
              )}
            </>
          )}

          {step === 4 && (
            <>
              <label className="inline-check">
                <input
                  type="checkbox"
                  checked={draft.callbackBrokerEnabled}
                  onChange={(e) => setDraft({ ...draft, callbackBrokerEnabled: e.target.checked })}
                />
                Enable message-broker callback
              </label>

              {draft.callbackBrokerEnabled && (
                <>
                  <div className="builder-grid">
                    <label>
                      Broker Type
                      <select
                        value={draft.callbackBrokerType}
                        onChange={(e) =>
                          setDraft({ ...draft, callbackBrokerType: e.target.value as "kafka" | "rabbitmq" })
                        }
                      >
                        <option value="kafka">Kafka</option>
                        <option value="rabbitmq">RabbitMQ</option>
                      </select>
                    </label>
                    <label>
                      Broker Delay
                      <input
                        value={draft.callbackBrokerDelay}
                        onChange={(e) => setDraft({ ...draft, callbackBrokerDelay: e.target.value })}
                      />
                    </label>
                  </div>

                  {draft.callbackBrokerType === "kafka" ? (
                    <div className="builder-grid">
                      <label>
                        Kafka Brokers (comma separated)
                        <input
                          value={draft.callbackKafkaBrokers}
                          onChange={(e) => setDraft({ ...draft, callbackKafkaBrokers: e.target.value })}
                        />
                      </label>
                      <label>
                        Kafka Topic
                        <input
                          value={draft.callbackKafkaTopic}
                          onChange={(e) => setDraft({ ...draft, callbackKafkaTopic: e.target.value })}
                        />
                      </label>
                      <label>
                        Kafka Key
                        <input
                          value={draft.callbackKafkaKey}
                          onChange={(e) => setDraft({ ...draft, callbackKafkaKey: e.target.value })}
                        />
                      </label>
                    </div>
                  ) : (
                    <div className="builder-grid">
                      <label>
                        RabbitMQ URL
                        <input
                          value={draft.callbackRabbitURL}
                          onChange={(e) => setDraft({ ...draft, callbackRabbitURL: e.target.value })}
                        />
                      </label>
                      <label>
                        Exchange
                        <input
                          value={draft.callbackRabbitExchange}
                          onChange={(e) => setDraft({ ...draft, callbackRabbitExchange: e.target.value })}
                        />
                      </label>
                      <label>
                        Routing Key
                        <input
                          value={draft.callbackRabbitRoutingKey}
                          onChange={(e) => setDraft({ ...draft, callbackRabbitRoutingKey: e.target.value })}
                        />
                      </label>
                    </div>
                  )}

                  <label className="full">
                    Broker Message JSON
                    <textarea
                      className="builder-json"
                      value={draft.callbackBrokerMessage}
                      onChange={(e) => setDraft({ ...draft, callbackBrokerMessage: e.target.value })}
                    />
                  </label>
                </>
              )}
            </>
          )}

          <div className="wizard-actions">
            <button className="btn ghost" onClick={goBack} disabled={step === 1} type="button">
              Back
            </button>
            {step < 4 ? (
              <button className="btn" onClick={goNext} type="button">
                Next
              </button>
            ) : (
              <button className="btn" onClick={applyBuilder} type="button">
                Apply Builder to mocks.json
              </button>
            )}
          </div>
        </section>
      )}

      <section className="grid">
        <article className="panel">
          <div className="title">config.yml</div>
          <textarea spellCheck={false} value={config} onChange={(e) => setConfig(e.target.value)} />
        </article>
        <article className="panel">
          <div className="title">mocks.json</div>
          <textarea spellCheck={false} value={mocks} onChange={(e) => setMocks(e.target.value)} />
        </article>
      </section>
    </main>
  );
}
