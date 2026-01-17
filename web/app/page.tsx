'use client';

import { useState } from 'react';

export default function Home() {
  const [dbUrl, setDbUrl] = useState('http://localhost:8081');
  const [key, setKey] = useState('');
  const [value, setValue] = useState('');
  const [searchKey, setSearchKey] = useState('');
  const [searchResult, setSearchResult] = useState<string | null>(null);
  const [status, setStatus] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSet = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setStatus('');
    try {
      const res = await fetch(`${dbUrl}/set`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key, value }),
      });
      const data = await res.json();
      if (data.success) {
        setStatus(`Successfully set ${key}`);
        setKey('');
        setValue('');
      } else {
        setStatus(`Error: ${data.error || 'Unknown error'}`);
      }
    } catch (err) {
      setStatus(`Failed to connect to GrassDB at ${dbUrl}`);
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleGet = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setSearchResult(null);
    try {
      const res = await fetch(`${dbUrl}/get?key=${searchKey}`);
      if (res.status === 404) {
        setSearchResult('Key not found');
      } else if (!res.ok) {
        setSearchResult(`Error: ${res.statusText}`);
      } else {
        const data = await res.json();
        setSearchResult(data.value || 'Empty value');
      }
    } catch (err) {
      setSearchResult(`Failed to connect to GrassDB at ${dbUrl}`);
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="min-h-screen bg-gradient-to-br from-gray-900 to-black text-white p-8 font-sans">
      <div className="max-w-4xl mx-auto">
        <header className="mb-12 flex justify-between items-center">
          <h1 className="text-4xl font-extrabold tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-green-400 to-emerald-600">
            GrassDB
          </h1>
          <div className="flex items-center gap-2 bg-white/5 p-2 rounded-lg border border-white/10 backdrop-blur-sm">
            <span className="text-gray-400 text-sm">Node URL:</span>
            <input
              type="text"
              value={dbUrl}
              onChange={(e) => setDbUrl(e.target.value)}
              className="bg-transparent text-white focus:outline-none text-sm w-48"
            />
          </div>
        </header>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          {/* Write Section */}
          <section className="bg-white/5 p-6 rounded-2xl border border-white/10 shadow-xl backdrop-blur-md">
            <h2 className="text-2xl font-bold mb-6 flex items-center gap-2">
              <span className="w-8 h-8 rounded-full bg-green-500/20 text-green-400 flex items-center justify-center text-sm">RW</span>
              Write Data
            </h2>
            <form onSubmit={handleSet} className="space-y-4">
              <div>
                <label className="block text-sm text-gray-400 mb-1">Key</label>
                <input
                  type="text"
                  value={key}
                  onChange={(e) => setKey(e.target.value)}
                  className="w-full bg-black/40 border border-white/10 rounded-lg p-3 focus:border-green-500 focus:ring-1 focus:ring-green-500 transition-all outline-none"
                  placeholder="e.g., username"
                  required
                />
              </div>
              <div>
                <label className="block text-sm text-gray-400 mb-1">Value</label>
                <input
                  type="text"
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                  className="w-full bg-black/40 border border-white/10 rounded-lg p-3 focus:border-green-500 focus:ring-1 focus:ring-green-500 transition-all outline-none"
                  placeholder="e.g., admin"
                  required
                />
              </div>
              <button
                type="submit"
                disabled={loading}
                className="w-full bg-gradient-to-r from-green-600 to-emerald-600 hover:from-green-500 hover:to-emerald-500 text-white font-bold py-3 rounded-lg transition-all transform hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? 'Writing...' : 'Set Key'}
              </button>
              {status && (
                <div className={`mt-4 p-3 rounded-lg text-sm ${status.startsWith('Error') ? 'bg-red-500/20 text-red-200' : 'bg-green-500/20 text-green-200'}`}>
                  {status}
                </div>
              )}
            </form>
          </section>

          {/* Read Section */}
          <section className="bg-white/5 p-6 rounded-2xl border border-white/10 shadow-xl backdrop-blur-md">
            <h2 className="text-2xl font-bold mb-6 flex items-center gap-2">
              <span className="w-8 h-8 rounded-full bg-blue-500/20 text-blue-400 flex items-center justify-center text-sm">RO</span>
              Read Data
            </h2>
            <form onSubmit={handleGet} className="space-y-4">
              <div>
                <label className="block text-sm text-gray-400 mb-1">Key to Search</label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={searchKey}
                    onChange={(e) => setSearchKey(e.target.value)}
                    className="flex-1 bg-black/40 border border-white/10 rounded-lg p-3 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all outline-none"
                    placeholder="e.g., username"
                    required
                  />
                  <button
                    type="submit"
                    disabled={loading}
                    className="bg-blue-600 hover:bg-blue-500 text-white font-bold px-6 rounded-lg transition-colors disabled:opacity-50"
                  >
                    Get
                  </button>
                </div>
              </div>
            </form>

            <div className="mt-8">
              <h3 className="text-sm text-gray-400 mb-2">Reflected Value</h3>
              <div className="h-32 bg-black/60 rounded-lg border border-white/5 p-4 overflow-auto font-mono text-sm text-gray-300">
                {searchResult !== null ? (
                  searchResult
                ) : (
                  <span className="text-gray-600 italic">No query results yet...</span>
                )}
              </div>
            </div>
          </section>
        </div>

        <footer className="mt-16 text-center text-gray-500 text-sm">
          <p>Designed for GrassDB &bull; Host on Vercel</p>
        </footer>
      </div>
    </main>
  );
}
