'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { toast } from '@/components/Toast';
import { MdAdd, MdClose } from 'react-icons/md';

export default function KeysPage() {
  const [keys, setKeys] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [newName, setNewName] = useState('');

  async function loadKeys() {
    try { const d = await api.keys(); setKeys(d?.keys || d || []); } catch {}
    setLoading(false);
  }
  useEffect(() => { loadKeys(); }, []);

  async function handleDelete(id) {
    try { await api.deleteKey(id); toast('Key deleted'); loadKeys(); } catch (err) { toast(err.message, 'error'); }
  }

  async function handleCreate() {
    try { await api.createKey({ name: newName }); toast('Key created'); setShowModal(false); setNewName(''); loadKeys(); } catch (err) { toast(err.message, 'error'); }
  }

  return (
    <>
      <div className="flex items-center justify-between mb-4">
        <div className="text-xs font-semibold text-text-muted/60 uppercase tracking-wider">API Keys</div>
        <button onClick={() => setShowModal(true)}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded text-sm font-semibold transition-all active:scale-[0.97] bg-primary text-white hover:bg-primary-hover">
          <MdAdd className="text-[16px]" />
          New Key
        </button>
      </div>

      <div className="overflow-x-auto rounded-lg border border-border-subtle">
        <table className="w-full">
          <thead>
            <tr className="bg-bg-alt">
              <th className="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Name</th>
              <th className="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Key</th>
              <th className="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Status</th>
              <th className="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Created</th>
              <th className="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={5} className="text-center py-8 text-sm text-text-subtle">Loading...</td></tr>
            ) : keys.length === 0 ? (
              <tr><td colSpan={5} className="text-center py-8 text-sm text-text-subtle">No API keys found</td></tr>
            ) : keys.map(k => (
              <tr key={k.id || k.key_id} className="border-t border-border">
                <td className="px-4 py-3 text-sm font-medium text-text-main">{k.name || '—'}</td>
                <td className="px-4 py-3 text-sm font-mono text-text-muted">{(k.key || k.key_prefix || '').substring(0, 20)}...</td>
                <td className="px-4 py-3">
                  <span className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold ${k.is_active ? 'bg-success/10 text-success' : 'bg-danger/10 text-danger'}`}>{k.is_active ? 'Active' : 'Inactive'}</span>
                </td>
                <td className="px-4 py-3 text-sm text-text-muted">{k.created_at ? new Date(k.created_at).toLocaleDateString() : '—'}</td>
                <td className="px-4 py-3">
                  <button onClick={() => handleDelete(k.id || k.key_id)} className="text-sm font-semibold text-danger hover:text-danger/80 transition-colors">Delete</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4" style={{ background: 'rgba(0,0,0,0.6)', backdropFilter: 'blur(4px)' }} onClick={() => setShowModal(false)}>
          <div className="card-elev w-full max-w-md p-6" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-text-main">Create API Key</h2>
              <button onClick={() => setShowModal(false)} className="text-text-muted hover:text-text-main transition-colors">
                <MdClose className="text-xl" />
              </button>
            </div>
            <div className="mb-6">
              <label className="block text-sm mb-1.5 text-text-muted">Name</label>
              <input type="text" value={newName} onChange={e => setNewName(e.target.value)}
                className="w-full px-3 py-2 rounded text-sm outline-none bg-bg-alt border border-border-subtle text-text-main focus:border-primary focus:shadow-focus transition-colors"
                placeholder="My Key" autoFocus />
            </div>
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowModal(false)} className="px-4 py-2 rounded text-sm font-semibold transition-all active:scale-[0.97] bg-surface-2 text-text-main hover:bg-surface-3">Cancel</button>
              <button onClick={handleCreate} className="px-4 py-2 rounded text-sm font-semibold transition-all active:scale-[0.97] bg-primary text-white hover:bg-primary-hover">Create</button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
