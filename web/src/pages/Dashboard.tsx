function Dashboard() {
  return (
    <div className="flex min-h-screen bg-gray-950">
      <aside className="w-64 border-r border-gray-800 bg-gray-900 p-4">
        <h2 className="text-lg font-semibold text-white">Orbita</h2>
        <p className="mt-2 text-sm text-gray-400">v0.1.0</p>
      </aside>
      <main className="flex-1 p-8">
        <h1 className="text-3xl font-bold text-white">Dashboard</h1>
        <p className="mt-4 text-gray-400">
          Welcome to Orbita. Dashboard will be built in Phase 14.
        </p>
      </main>
    </div>
  );
}

export default Dashboard;
