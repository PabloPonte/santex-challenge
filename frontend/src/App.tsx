import { FormEvent, type ReactNode, useEffect, useMemo, useState } from "react";
import { Link, Navigate, Route, Routes, useNavigate, useParams } from "react-router-dom";
import { ApiError, api, type Evaluation, type Feature, type FeatureInput, type FeatureStatus } from "./api";
import { formatDate, statusDescription, statusLabel } from "./status";

type Notice = { tone: "success" | "error"; message: string } | null;

export function App() {
  const [notice, setNotice] = useState<Notice>(null);
  return <Routes>
    <Route path="/" element={<Inventory notice={notice} setNotice={setNotice} />} />
    <Route path="/features/new" element={<FeatureEditor notice={setNotice} />} />
    <Route path="/features/:name" element={<FeatureEditor notice={setNotice} />} />
    <Route path="*" element={<Navigate to="/" replace />} />
  </Routes>;
}

function Shell({ children, notice }: { children: ReactNode; notice?: Notice }) {
  return <div className="app-shell">
    <aside className="rail">
      <Link className="wordmark" to="/"><span>Signal</span><em>Room</em></Link>
      <p className="rail-intro">Feature flags, kept deliberate.</p>
      <nav aria-label="Primary navigation"><Link className="nav-link active" to="/">Feature inventory</Link></nav>
      <div className="environment"><span className="pulse" /> Internal API<br /><strong>Trusted network</strong></div>
    </aside>
    <main className="workspace">
      {notice && <div className={`notice ${notice.tone}`} role="status">{notice.message}</div>}
      {children}
    </main>
  </div>;
}

function Inventory({ notice, setNotice }: { notice: Notice; setNotice: (notice: Notice) => void }) {
  const [features, setFeatures] = useState<Feature[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState<"all" | FeatureStatus>("all");
  const [confirmRefresh, setConfirmRefresh] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  const load = async () => {
    setLoading(true); setError(null);
    try { setFeatures(await api.listFeatures()); } catch (err) { setError(messageFor(err)); } finally { setLoading(false); }
  };
  useEffect(() => { void load(); }, []);
  const visible = useMemo(() => features.filter((feature) =>
    (status === "all" || feature.status === status) &&
    `${feature.name} ${feature.description}`.toLowerCase().includes(query.toLowerCase())
  ), [features, query, status]);
  const refresh = async () => {
    setRefreshing(true);
    try { const { refreshed } = await api.refreshCache(); setNotice({ tone: "success", message: `Cache refreshed with ${refreshed} ${refreshed === 1 ? "feature" : "features"}.` }); setConfirmRefresh(false); }
    catch (err) { setNotice({ tone: "error", message: messageFor(err) }); }
    finally { setRefreshing(false); }
  };

  return <Shell notice={notice}>
    <header className="page-head">
      <div><p className="kicker">Operations</p><h1>Feature inventory</h1><p>Decide who sees what, then verify the result from the cache.</p></div>
      <div className="head-actions"><button className="button ghost" onClick={() => setConfirmRefresh(true)}>Refresh cache</button><Link className="button primary" to="/features/new">Create feature</Link></div>
    </header>
    <section className="control-strip" aria-label="Feature filters">
      <label className="search"><span className="sr-only">Search features</span><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search name or description" /></label>
      <label className="select-label">Status<select value={status} onChange={(event) => setStatus(event.target.value as typeof status)}><option value="all">All states</option><option value="open">Open</option><option value="closed">Closed</option><option value="whitelisted">Whitelist only</option></select></label>
      <span className="result-count">{visible.length} shown</span>
    </section>
    {loading ? <div className="state-panel">Loading feature inventory…</div> : error ? <div className="state-panel error-state"><p>{error}</p><button className="button ghost" onClick={() => void load()}>Try again</button></div> : visible.length === 0 ? <Empty hasFeatures={features.length > 0} /> :
      <div className="table-wrap"><table><thead><tr><th>Feature</th><th>Access</th><th>Changed</th><th>Whitelist</th><th><span className="sr-only">Actions</span></th></tr></thead><tbody>
        {visible.map((feature) => <tr key={feature.name}><td><Link className="feature-name" to={`/features/${encodeURIComponent(feature.name)}`}>{feature.name}</Link><span className="description">{feature.description || "No description"}</span></td><td><StatusBadge status={feature.status} /></td><td>{formatDate(feature.statusDate)}</td><td>{feature.status === "whitelisted" ? `${feature.whitelist.length} users` : "—"}</td><td><Link className="text-link" to={`/features/${encodeURIComponent(feature.name)}`}>Edit</Link></td></tr>)}
      </tbody></table></div>}
    <EvaluationPanel features={features} />
    {confirmRefresh && <ConfirmDialog title="Refresh the Redis cache?" detail="The cache will be synchronously rebuilt from PostgreSQL." confirmLabel="Refresh cache" pending={refreshing} onConfirm={() => void refresh()} onCancel={() => setConfirmRefresh(false)} />}
  </Shell>;
}

function Empty({ hasFeatures }: { hasFeatures: boolean }) { return <div className="state-panel"><h2>{hasFeatures ? "No matching features" : "Your inventory is empty"}</h2><p>{hasFeatures ? "Change the search or status filter to see more features." : "Create the first feature flag to begin controlling access."}</p>{!hasFeatures && <Link className="button primary" to="/features/new">Create feature</Link>}</div>; }

function FeatureEditor({ notice }: { notice: (notice: Notice) => void }) {
  const { name } = useParams(); const navigate = useNavigate(); const editing = Boolean(name);
  const [initial, setInitial] = useState<Feature | null>(null); const [loading, setLoading] = useState(editing);
  const [error, setError] = useState<string | null>(null); const [deleting, setDeleting] = useState(false); const [deletePending, setDeletePending] = useState(false);
  useEffect(() => { if (!name) return; void (async () => { try { setInitial(await api.getFeature(name)); } catch (err) { setError(messageFor(err)); } finally { setLoading(false); } })(); }, [name]);
  if (loading) return <Shell><div className="state-panel">Loading feature…</div></Shell>;
  if (error || (editing && !initial)) return <Shell><div className="state-panel error-state"><p>{error ?? "Feature not found."}</p><Link className="button ghost" to="/">Back to inventory</Link></div></Shell>;
  const remove = async () => { if (!name) return; setDeletePending(true); try { await api.deleteFeature(name); notice({ tone: "success", message: `Deleted ${name}.` }); navigate("/"); } catch (err) { notice({ tone: "error", message: messageFor(err) }); setDeletePending(false); } };
  return <Shell>
    <header className="page-head editor-head"><div><Link className="back-link" to="/">Feature inventory</Link><p className="kicker">{editing ? "Edit feature" : "New feature"}</p><h1>{editing ? initial!.name : "Create a feature"}</h1><p>{editing ? "Names identify flags permanently. Adjust access without changing the identifier." : "Describe the audience and choose the access rule."}</p></div>{editing && <button className="button danger" onClick={() => setDeleting(true)}>Delete feature</button>}</header>
    <FeatureForm initial={initial ?? undefined} onSaved={(feature, created) => { notice({ tone: "success", message: created ? `Created ${feature.name}.` : `Saved changes to ${feature.name}.` }); navigate("/"); }} />
    {deleting && <ConfirmDialog title={`Delete ${name}?`} detail="This removes the feature from PostgreSQL and Redis. This action cannot be undone." confirmLabel="Delete feature" destructive pending={deletePending} onConfirm={() => void remove()} onCancel={() => setDeleting(false)} />}
  </Shell>;
}

function FeatureForm({ initial, onSaved }: { initial?: Feature; onSaved: (feature: Feature, created: boolean) => void }) {
  const [name, setName] = useState(initial?.name ?? ""); const [description, setDescription] = useState(initial?.description ?? ""); const [status, setStatus] = useState<FeatureStatus>(initial?.status ?? "open"); const [users, setUsers] = useState(initial?.whitelist.join("\n") ?? ""); const [error, setError] = useState<string | null>(null); const [saving, setSaving] = useState(false);
  const submit = async (event: FormEvent) => {
    event.preventDefault(); const whitelist = users.split("\n").map((user) => user.trim()).filter(Boolean);
    if (!initial && !/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(name)) return setError("Use lowercase letters, numbers, and hyphens for the feature name.");
    if (description.length > 2000) return setError("Description must be 2,000 characters or fewer.");
    if (new Set(whitelist).size !== whitelist.length) return setError("Each whitelist user must appear only once.");
    setSaving(true); setError(null); const input: FeatureInput = { description, status, whitelist, ...(initial ? {} : { name }) };
    try { const feature = initial ? await api.updateFeature(initial.name, input) : await api.createFeature(input); onSaved(feature, !initial); } catch (err) { setError(messageFor(err)); } finally { setSaving(false); }
  };
  return <form className="feature-form" onSubmit={submit}><div className="form-column"><label>Feature name<input value={name} disabled={Boolean(initial)} onChange={(event) => setName(event.target.value)} maxLength={100} required aria-describedby="name-help" /></label><p id="name-help" className="hint">Lowercase letters, numbers, and hyphens. This cannot change later.</p><label>Description<textarea value={description} onChange={(event) => setDescription(event.target.value)} maxLength={2000} rows={5} required /></label></div><fieldset><legend>Access rule</legend><div className="status-options">{(["open", "closed", "whitelisted"] as FeatureStatus[]).map((value) => <label className={`status-choice ${status === value ? "selected" : ""}`} key={value}><input type="radio" name="status" value={value} checked={status === value} onChange={() => setStatus(value)} /><span><strong>{statusLabel[value]}</strong><small>{statusDescription[value]}</small></span></label>)}</div></fieldset><label>Whitelist<textarea value={users} onChange={(event) => setUsers(event.target.value)} rows={7} placeholder="One user ID per line" disabled={status !== "whitelisted"} /><span className="hint">Only used by the whitelist-only rule. Up to 10,000 unique user IDs.</span></label>{error && <div className="form-error" role="alert">{error}</div>}<div className="form-actions"><Link className="button ghost" to="/">Cancel</Link><button className="button primary" disabled={saving}>{saving ? "Saving…" : initial ? "Save changes" : "Create feature"}</button></div></form>;
}

function EvaluationPanel({ features }: { features: Feature[] }) {
  const [name, setName] = useState(""); const [userId, setUserId] = useState(""); const [result, setResult] = useState<Evaluation | null>(null); const [error, setError] = useState<string | null>(null); const [pending, setPending] = useState(false);
  const submit = async (event: FormEvent) => { event.preventDefault(); setPending(true); setError(null); setResult(null); try { setResult(await api.evaluate(name, userId)); } catch (err) { setError(messageFor(err)); } finally { setPending(false); } };
  return <section className="evaluator"><div><p className="kicker">Cache check</p><h2>Test a user’s access</h2><p>This request uses the external Redis-only endpoint.</p></div><form onSubmit={submit}><label>Feature<select value={name} onChange={(event) => setName(event.target.value)} required><option value="" disabled>Select a feature</option>{features.map((feature) => <option key={feature.name} value={feature.name}>{feature.name}</option>)}</select></label><label>User ID<input value={userId} onChange={(event) => setUserId(event.target.value)} maxLength={255} required /></label><button className="button primary" disabled={pending || !features.length}>{pending ? "Checking…" : "Check access"}</button></form>{result && <div className={`evaluation-result ${result.enabled ? "enabled" : "disabled"}`}><strong>{result.enabled ? "Enabled" : "Disabled"}</strong><span>{result.userId} · {statusLabel[result.status]}</span></div>}{error && <p className="form-error" role="alert">{error}</p>}</section>;
}

function StatusBadge({ status }: { status: FeatureStatus }) { return <span className={`badge ${status}`}>{statusLabel[status]}</span>; }
function ConfirmDialog({ title, detail, confirmLabel, destructive, pending, onConfirm, onCancel }: { title: string; detail: string; confirmLabel: string; destructive?: boolean; pending: boolean; onConfirm: () => void; onCancel: () => void }) { return <div className="dialog-backdrop" role="presentation"><section className="dialog" role="dialog" aria-modal="true" aria-labelledby="dialog-title"><h2 id="dialog-title">{title}</h2><p>{detail}</p><div><button className="button ghost" onClick={onCancel}>Cancel</button><button className={`button ${destructive ? "danger" : "primary"}`} onClick={onConfirm} disabled={pending}>{pending ? "Working…" : confirmLabel}</button></div></section></div>; }
function messageFor(error: unknown) { if (error instanceof ApiError) { if (error.status === 503) return "A dependency is unavailable. Try again after PostgreSQL and Redis recover."; if (error.status === 409) return "A feature with this name already exists."; return error.message; } return "Something unexpected happened. Try again."; }
